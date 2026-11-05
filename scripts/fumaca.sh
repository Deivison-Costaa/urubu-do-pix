#!/usr/bin/env bash
# Teste de fumaça do ambiente local. Pressupõe `docker compose up -d --wait`.
# Verifica sondas, roteamento e limites do gateway, segmentação de rede,
# tópicos do Redpanda e o invariante de conservação nos dados iniciais.
set -euo pipefail

GATEWAY="http://localhost:${GATEWAY_PORTA:-8080}"
falhas=0
PROJETO="${COMPOSE_PROJECT_NAME:-urubu-do-pix}"
IMAGEM_PG="postgres:18.6-alpine"

ok()   { printf '  \033[32mok\033[0m    %s\n' "$1"; }
erro() { printf '  \033[31mFALHA\033[0m %s\n' "$1"; falhas=$((falhas + 1)); }
checar() { local desc=$1; shift; if "$@" >/dev/null 2>&1; then ok "$desc"; else erro "$desc"; fi; }
interno() { docker compose run --rm --no-deps -T curl -fsS --max-time 5 "$@"; }
status()  { curl -s -o /dev/null -w '%{http_code}' --max-time 5 "$1"; }
psql_em() { docker compose exec -T "$1" psql -U "$2" -d "$2" -tAc "$3"; }

# O --wait garante os healthchecks, mas o gateway ainda pode estar fazendo
# o primeiro health check ativo nos PSPs. Espera até 60 s antes de testar.
printf 'Aguardando o gateway'
for _ in $(seq 1 60); do
  if curl -fsS --max-time 2 "$GATEWAY/psp/alfa/readyz" >/dev/null 2>&1 &&
     curl -fsS --max-time 2 "$GATEWAY/psp/beta/readyz" >/dev/null 2>&1; then
    break
  fi
  printf '.'; sleep 1
done
echo

echo "Sondas de prontidão"
for psp in alfa beta; do
  checar "gateway -> psp-$psp /readyz" curl -fsS --max-time 5 "$GATEWAY/psp/$psp/readyz"
done
for svc in spi dict antifraude auditor; do
  checar "$svc /readyz (rede interna)" interno "http://$svc:8080/readyz"
done

echo "Gateway"
[[ $(status "$GATEWAY/psp/alfa/metrics") == 404 ]] && ok "métricas não expostas na borda" || erro "métricas expostas na borda"
[[ $(status "$GATEWAY/spi/readyz") == 404 ]] && ok "SPI não exposto na borda" || erro "SPI exposto na borda"
limitadas=0
for _ in $(seq 1 30); do
  [[ $(status "$GATEWAY/psp/alfa/v1/chaves/52998224725") == 429 ]] && limitadas=$((limitadas + 1))
done
(( limitadas > 0 )) && ok "rate limit de consulta de chaves ($limitadas de 30 recusadas com 429)" || erro "rate limit de consulta de chaves não atuou"

[[ $(docker compose run --rm --no-deps -T curl -s -o /dev/null -w '%{http_code}' --max-time 5 http://gateway:9902/stats/prometheus) == 200 ]] \
  && ok "métricas do gateway disponíveis para o Prometheus" || erro "métricas do gateway indisponíveis"
[[ $(docker compose run --rm --no-deps -T curl -s -o /dev/null -w '%{http_code}' --max-time 5 -X POST http://gateway:9902/quitquitquit) == 404 ]] \
  && ok "admin do Envoy inacessível pela rede" || erro "admin do Envoy acessível pela rede"

echo "Segmentação de rede"
# pg_isready a partir de redes diferentes: só a rede do próprio banco alcança.
alcanca() { docker run --rm --network "${PROJETO}_$1" "$IMAGEM_PG" pg_isready -q -h "$2" -t 3; }
checar "pg-spi alcançável pela rede db-spi (controle)" alcanca db-spi pg-spi
for rede in nucleo borda; do
  if alcanca "$rede" pg-spi 2>/dev/null; then erro "pg-spi alcançável pela rede $rede"; else ok "pg-spi inalcançável pela rede $rede"; fi
done
if alcanca db-psp-alfa pg-psp-beta 2>/dev/null; then erro "banco do Beta alcançável pela rede do Alfa"; else ok "banco do Beta inalcançável pela rede do Alfa"; fi

echo "Mensageria"
topicos=$(docker compose exec -T redpanda rpk topic list)
for t in pix.spi.entrada pix.psp.10000001.entrada pix.psp.20000002.entrada; do
  grep -q "^$t " <<<"$topicos" && grep -q "^$t.dlq " <<<"$topicos" && ok "tópico $t e DLQ" || erro "tópico $t ou DLQ ausente"
done

echo "Invariante de conservação (dados iniciais)"
for par in "pg-psp-alfa 10000001" "pg-psp-beta 20000002"; do
  read -r db ispb <<<"$par"
  clientes=$(psql_em "$db" psp "SELECT COALESCE(SUM(saldo_centavos),0) FROM contas")
  reserva=$(psql_em pg-spi spi "SELECT saldo_centavos FROM contas_reserva WHERE ispb='$ispb'")
  [[ "$clientes" == "$reserva" ]] && ok "$ispb: clientes ($clientes) = reserva ($reserva)" || erro "$ispb: clientes ($clientes) != reserva ($reserva)"
done

echo
if (( falhas > 0 )); then echo "$falhas verificação(ões) falharam"; exit 1; fi
echo "Todas as verificações passaram"
