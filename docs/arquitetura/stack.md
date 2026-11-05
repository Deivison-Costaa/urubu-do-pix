# Stack tecnológico

Cada escolha está ligada a um critério de sucesso da proposta (C1 a C7) ou a um ADR. As versões são as usadas no `compose.yaml` e no `go.mod`.

| Camada | Tecnologia | Versão | Por que | Alternativas descartadas |
|---|---|---|---|---|
| Linguagem | Go | 1.26 | Concorrência nativa (goroutines) para consumidores e relays; binário estático que cabe em imagem distroless; biblioteca padrão com HTTP, `slog` e roteamento por padrão. Atende C3 com pouca memória | Java/Spring (mais memória por instância e partida lenta, o que pesa com 6 serviços no notebook); Node.js (aritmética monetária e CPU-bound mais frágeis) |
| Banco de dados | PostgreSQL, **um por serviço** | 18 | Transações ACID com `CHECK` e `UNIQUE` que reforçam os invariantes (saldo ≥ 0, lançamento único por pagamento); `SELECT ... FOR UPDATE` para a liquidação | MySQL (equivalente, mas a equipe conhece melhor o Postgres); banco compartilhado (acopla serviços e quebra o isolamento de falhas) |
| Driver Postgres | pgx | v5 | Driver nativo mais rápido em Go, com pool e suporte a `COPY` | `database/sql` + lib/pq (lib/pq em manutenção) |
| Mensageria | Redpanda (API Kafka) | 26.2 | Log durável, ordem por chave, releitura; binário único e leve no Compose (ADR-0002) | Kafka (JVM, mais pesado); RabbitMQ (sem releitura do log); NATS |
| Cliente Kafka | franz-go | 1.22 | Cliente Go puro (sem CGO, compatível com distroless), rápido e mantido | confluent-kafka-go (CGO/librdkafka); sarama |
| Cache | Redis | 8.8 | Cache-aside de chaves no DICT e contadores de rate limit por PSP | Memcached (sem estruturas para contadores com expiração) |
| API Gateway | Envoy | 1.39 | Rate limit, health check ativo, `jwt_authn` e `ext_authz` nativos; o mesmo proxy de service mesh (ADR-0003) | Kong, Traefik, NGINX |
| Identidade (P03) | Keycloak | — | OAuth 2.0/OIDC completo e gratuito, emite JWT com JWKS | Auth0 (externo e pago), implementação própria (arriscado) |
| Autorização (P03) | OPA (Rego) | — | Políticas como código, versionadas e testáveis, integradas ao Envoy | Regras espalhadas no código dos serviços |
| Métricas | Prometheus + client_golang | 3.15 | Padrão de fato, com PromQL para p99 e taxa de erro (C3, C5, C7) | — |
| Painéis | Grafana | 13 | Painéis e alertas sobre o Prometheus | — |
| Tracing (P03) | OpenTelemetry + Jaeger | — | Rastrear um pagamento atravessando PSP, SPI e PSP | Zipkin |
| Containers | Docker Compose | v2 | Ambiente completo com um comando, perfis opcionais e redes segmentadas | — |
| Orquestração (P03) | Kubernetes (kind) + Argo Rollouts | — | Canary com análise automática e rollback (C7) | Docker Swarm (sem canary nativo) |
| CI/CD | GitHub Actions | — | Integrado ao repositório; OIDC para assinatura keyless (ADR-0004) | GitLab CI, Jenkins |
| Segurança no pipeline | golangci-lint (gosec), govulncheck, Trivy, Syft, Cosign, Dependabot | — | SAST, SCA, scan de imagem, SBOM e assinatura: shift-left completo (ADR-0004) | Snyk (pago para uso completo) |
| Testes de carga (P03) | k6 | — | Scripts em JS, métricas de percentil e limites (thresholds) que quebram o teste | JMeter (mais pesado de versionar), Locust |
| Injeção de falhas (P03) | Toxiproxy | — | Latência, corte e banda limitada entre serviços específicos (C4) | Pumba, Chaos Mesh (precisa Kubernetes) |
| Testes de integração (P03) | Testcontainers for Go | — | Postgres e Redpanda reais em cada teste, descartados no fim | Mocks (não pegam erros de SQL nem de concorrência) |

## Princípios que guiaram as escolhas

- **Dinheiro em inteiros:** todo valor é `int64` em centavos (`internal/dominio/dinheiro`), no código e no banco. Ponto flutuante perde centavos em somas.
- **Uma imagem para todos os serviços:** o mesmo Dockerfile compila qualquer pasta de `cmd/`, o que padroniza build, scan e assinatura.
- **O mesmo código de PSP para todos os bancos:** cada banco é uma implantação separada (single-tenant), com o seu banco de dados, ISPB e tópico. Adicionar um banco é configuração, não código.
- **Biblioteca padrão primeiro:** roteamento, logs e healthcheck usam `net/http` e `log/slog`; dependências externas só onde trazem algo que a biblioteca padrão não tem (drivers, Kafka, Prometheus).
