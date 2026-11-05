# Divisão de responsabilidades

Cada integrante é dono de uma frente: responde pelas decisões (ADRs), pelo código e pelos testes dela, e apresenta essa parte. Todos revisam PRs de outras frentes. A nota individual considera os commits, então cada um commita o próprio trabalho.

A equipe começou com cinco integrantes. Em novembro de 2026 ficou com três, e as cinco frentes previstas no Projeto 01 foram redistribuídas.

| Integrante | Frente | Entregue no Projeto 02 | Responsável no Projeto 03 |
|---|---|---|---|
| **Deivison Costa** ([@Deivison-Costaa](https://github.com/Deivison-Costaa)) | Liquidação, observabilidade e coordenação | Plataforma comum dos serviços (`internal/plataforma`), tipo monetário, esqueleto do SPI e esquema das reservas, C4 níveis 1 e 2, ADR-0001, stack, padrões, README, apresentação | Liquidação atômica no SPI, máquina de estados da SAGA, prazo de aceite e compensação; OpenTelemetry e Jaeger; testes de carga (k6) |
| **Giancarlo Silveira Cavalcante** ([@GianBala](https://github.com/GianBala)) | Bancos, mensageria e qualidade | EndToEndId, esqueleto do PSP e esquema com Outbox e Inbox, tópicos do Redpanda, ADR-0002; esqueleto do auditor, Prometheus e Grafana, teste de fumaça, plano de testes, ADR-0005 | API de pagamentos do PSP, relay do Outbox, consumidores idempotentes, retry e DLQ; auditor de invariantes; testes de caos (Toxiproxy) |
| **Herlan Alef de Lima Nascimento** ([@HerlanLima](https://github.com/HerlanLima)) | Borda, segurança e entrega | Esqueleto do antifraude, Dockerfile distroless, pipeline de CI, lint, Dependabot, ADR-0004; validação de chaves Pix, esqueleto do DICT, gateway Envoy, Docker Compose com segmentação de rede, ADR-0003 | DICT com cache-aside e limite por PSP; Keycloak e JWT no gateway, OPA, mTLS; antifraude com circuit breaker e hedging; Kubernetes (kind) e canary com Argo Rollouts |

## Combinados

- Trabalho em branch, PR para a `main` com o CI verde e revisão de pelo menos uma pessoa de outra frente.
- Commits no padrão Conventional Commits (`feat:`, `fix:`, `docs:`, `test:`, `ci:`, `build:`).
- Uso de IA registrado em [docs/ia/registro.md](../ia/registro.md) por quem usou, no momento do uso.
