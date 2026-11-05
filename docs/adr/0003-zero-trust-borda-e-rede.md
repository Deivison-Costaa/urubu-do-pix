# ADR-0003: Zero Trust: gateway Envoy, segmentação de rede e, no P03, JWT, OPA e mTLS

- **Status:** Aceito
- **Data:** 2026-11-04
- **Responsável:** Herlan Alef de Lima Nascimento

## Contexto

Um sistema de pagamentos é alvo óbvio. O critério C6 exige que 100% das chamadas sem token ou certificado válido sejam recusadas e que a varredura de chaves no DICT seja bloqueada (no Pix real, a consulta em massa de chaves é usada para levantar dados pessoais e aplicar golpes). O enunciado dá peso especial a Zero Trust e DevSecOps. O princípio de Zero Trust é não confiar em nada por estar "dentro da rede": cada chamada se autentica e cada componente só alcança o que precisa.

## Decisão

**Envoy como API Gateway**, ponto único de entrada:

- só os PSPs são expostos (`/psp/alfa/...`, `/psp/beta/...`); SPI, DICT, antifraude e auditor não têm rota na borda;
- rate limit global (1.000 req/s) e um limite estrito nas consultas de chave (10 req/s por banco) contra varredura;
- `/metrics` bloqueado na borda;
- health check ativo em `/readyz`: o gateway tira do balanceamento a instância que não está pronta;
- `outlier_detection`: depois de 5 erros 5xx seguidos, a instância é ejetada por 10 s;
- normalização de caminho e rejeição de barras codificadas, contra desvio de rota;
- interface administrativa do Envoy só no loopback do container; o Prometheus lê as métricas por um listener à parte que só aceita `GET /stats/prometheus`. Sem isso, um PSP comprometido poderia desligar os limites ou derrubar o gateway pela rede.

**Segmentação de rede no Docker Compose:** cada banco de dados fica numa rede `internal` que só o seu serviço alcança. Um PSP não alcança o banco do outro, e nenhum banco é alcançável da borda.

**Hardening dos containers:** imagem distroless sem shell, usuário sem privilégio, sistema de arquivos somente leitura, `no-new-privileges` e todas as capabilities removidas.

**No Projeto 03**, sobre a mesma base:

1. **Keycloak (OIDC)** emite JWT para os clientes; o Envoy valida assinatura, emissor, audiência e expiração (`jwt_authn`) antes de rotear.
2. **OPA** decide a autorização pelo `ext_authz` do Envoy, com as políticas em Rego versionadas no repositório (por exemplo, o cliente só movimenta a própria conta).
3. **mTLS entre serviços** com CA própria; cada serviço só aceita conexões de quem tem certificado assinado por ela. A evolução prevista é SPIFFE/SPIRE, para emitir identidade por workload.

## Alternativas consideradas

- **Kong ou Traefik:** mais simples de configurar, mas o Envoy tem `jwt_authn` e `ext_authz` nativos e é o mesmo proxy dos sidecars de service mesh (Istio), então o conhecimento serve para a evolução.
- **Service mesh completo (Istio, Linkerd) já no Compose:** daria mTLS automático, mas depende de Kubernetes. Fica para ser reavaliado no kind, no P03.
- **Autenticação dentro de cada serviço:** cada serviço reimplementaria validação de token, com risco de divergência. O gateway centraliza, e o mTLS cobre o tráfego interno.
- **Rede única no Compose:** é o padrão, mas qualquer container comprometido alcançaria todos os bancos.

## Consequências

- **Positivas:** superfície de ataque mínima na borda; um serviço comprometido não lê o banco de outro; segurança testável com testes negativos automáticos.
- **Negativas:** a configuração do Envoy é verbosa; mais redes deixam o Compose mais difícil de ler; no P03 o mTLS exige distribuir e renovar certificados.
- **Obrigatório:** todo serviço novo entra só nas redes de que precisa; toda rota nova na borda passa pelo gateway com limite definido.

## Validação

- Teste de fumaça (já no CI): métricas e SPI inacessíveis na borda; admin do Envoy inacessível pela rede; consultas de chave acima do limite recebem 429; bancos inalcançáveis fora da própria rede.
- P03: chamadas sem JWT, com JWT expirado ou com certificado de CA estranha são recusadas em 100% dos casos (C6).
