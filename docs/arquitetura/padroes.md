# Padrões arquiteturais

O enunciado pede no mínimo 3 padrões da Seção 5. O projeto aplica 12, distribuídos pelas cinco áreas técnicas. A coluna "Estado" indica o que já existe no código ou na configuração no Projeto 02 e o que entra no Projeto 03.

| # | Padrão | Área | Onde e como | Estado | ADR |
|---|---|---|---|---|---|
| 1 | **SAGA (orquestração)** | Confiabilidade | O SPI conduz cada pagamento por uma máquina de estados; rejeição ou prazo esgotado dispara o estorno no pagador | Esquema de estados nas tabelas `ordens` e `pagamentos`; lógica no P03 | 0001 |
| 2 | **Transactional Outbox** | Confiabilidade | Tabela `outbox` em PSPs e SPI, gravada na mesma transação da mudança de saldo; relay publica no Redpanda | Tabelas e índice criados; relay no P03 | 0001 |
| 3 | **Retry + DLQ** | Confiabilidade | Backoff exponencial no consumidor; cada tópico tem `<tópico>.dlq` | Tópicos e DLQs criados no Compose; consumidor no P03 | 0002 |
| 4 | **Database-per-service** | Desempenho | Quatro PostgreSQL independentes, cada um numa rede interna própria | Implementado | — |
| 5 | **API Gateway + Rate Limit** | Segurança, Desempenho | Envoy: só PSPs na borda; limite global e limite estrito em consultas de chave | Implementado e testado no CI | 0003 |
| 6 | **Zero Trust** (micro-segmentação; JWT, OPA e mTLS no P03) | Segurança | Redes isoladas por banco, containers somente leitura e sem privilégios | Segmentação implementada; JWT, OPA e mTLS no P03 | 0003 |
| 7 | **Event-Driven Architecture / Messaging** | Desempenho, Escalabilidade | Etapas da SAGA trocadas por mensagens; uma caixa de entrada por participante | Tópicos criados; produtores e consumidores no P03 | 0002 |
| 8 | **Cache-Aside** | Desempenho | DICT consulta o Redis primeiro e o Postgres no miss; invalidação na alteração da chave | Redis conectado ao DICT; lógica no P03 | — |
| 9 | **Circuit Breaker + Request Hedging** | Confiabilidade, Desempenho | PSP → antifraude: o breaker abre em ≤ 5 s com o antifraude fora (C5); hedging repete a consulta idempotente se a primeira passar do p95 | P03 | — |
| 10 | **Single-tenant / Multi-tenant** | Escalabilidade | O mesmo código de PSP implantado uma vez por banco, com dados e tópicos isolados | Implementado (Alfa e Beta) | — |
| 11 | **Stateless** | Escalabilidade | Antifraude e SPI não guardam estado em memória, então podem ter várias réplicas | Implementado (antifraude); SPI no P03 | — |
| 12 | **Canary Release** | Deployment | Nova versão do antifraude recebe parte do tráfego e volta sozinha se o erro passar de 1% (C7) | Imagens versionadas e assinadas no CI; Argo Rollouts no P03 | 0004 |

Também aparecem padrões de apoio: **Health Check** (sondas `/healthz` e `/readyz` separadas), **Idempotent Consumer** (Inbox), **ACID** na liquidação, **Strong vs Eventual Consistency** (forte dentro do SPI, eventual entre bancos) e **Immutable Deployment** (imagens distroless versionadas pelo commit).

## Cobertura das áreas técnicas

| Área | Padrões |
|---|---|
| Segurança | API Gateway, Rate Limit, Zero Trust, DevSecOps e supply chain (ADR-0004) |
| Escalabilidade | Database-per-service, single/multi-tenant, stateless, containers, messaging |
| Desempenho | Cache-Aside, Request Hedging, Rate Limit, EDA |
| Confiabilidade | SAGA, Outbox, Retry + DLQ, Circuit Breaker, ACID |
| Deployment | Pipeline CI/CD, Canary Release, Immutable Deployment |
