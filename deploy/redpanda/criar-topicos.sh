#!/bin/bash
# Uma caixa de entrada por participante, como no SPI real:
#   pix.spi.entrada            ordens e confirmações enviadas pelos PSPs ao SPI
#   pix.psp.<ISPB>.entrada     créditos, devoluções e resultados enviados pelo SPI a cada PSP
# Cada tópico tem a sua DLQ (.dlq) para mensagens que esgotaram os retries.
# A chave das mensagens é o EndToEndId: as mensagens de um mesmo pagamento
# caem na mesma partição e são consumidas em ordem.
set -euo pipefail

PARTICOES=6
topicos=(pix.spi.entrada pix.psp.10000001.entrada pix.psp.20000002.entrada)

for t in "${topicos[@]}"; do
  for nome in "$t" "$t.dlq"; do
    if rpk topic describe "$nome" >/dev/null 2>&1; then
      echo "tópico já existe: $nome"
    else
      rpk topic create "$nome" --partitions "$PARTICOES" --replicas 1
    fi
  done
done
rpk topic list
