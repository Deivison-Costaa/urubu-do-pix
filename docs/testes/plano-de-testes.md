# Plano de testes

Cada critério de sucesso da proposta tem um teste automático que o valida. Os testes de carga e de caos rodam numa máquina de referência, descrita no README junto com os resultados (P03). Todo cenário termina chamando o auditor de invariantes (ADR-0005).

## Níveis de teste

| Nível | Ferramenta | O que cobre | Onde roda | Estado |
|---|---|---|---|---|
| Unitário | `go test` (tabelas, `-race`) | Domínio (valor, EndToEndId, chave Pix), sondas de saúde, servidor, configuração | CI, em todo push e PR | **Feito**: 8 pacotes, cobertura entre 81% e 100% |
| Lint e SAST | golangci-lint (gosec) | Erros comuns, falhas de segurança no código | CI | **Feito** |
| SCA e imagem | govulncheck, Trivy | CVEs em dependências e imagens | CI | **Feito** |
| Fumaça | `scripts/fumaca.sh` | Ambiente sobe inteiro; gateway, rate limit, segmentação de rede, tópicos, invariante nos dados iniciais | CI e local | **Feito** |
| Integração | Testcontainers for Go | Repositórios SQL, Outbox, Inbox e consumidores contra Postgres e Redpanda reais | CI | P03 |
| Contrato | testes de esquema das mensagens | Produtores e consumidores concordam no formato dos eventos | CI | P03 |
| Carga | k6 + Prometheus | Latência p99 e vazão (C3) | máquina de referência | P03 |
| Resiliência | Toxiproxy, `docker compose kill/pause` | Partição de rede, queda de serviço, broker lento (C4, C5) | máquina de referência | P03 |
| Segurança negativa | k6 ou scripts | Sem token, token expirado, certificado estranho, varredura de chaves (C6) | CI | Parcial (rate limit e exposição na borda) |
| Deployment | Argo Rollouts | Canary com rollback automático (C7) | kind | P03 |

## Matriz critério → teste

| # | Critério | Cenário | Medida | Aprovação |
|---|---|---|---|---|
| C1 | Conservação | Todos os cenários abaixo, inclusive os de caos | Auditor: invariantes 1 a 3 | soma idêntica antes e depois |
| C2 | Idempotência | Reenviar 10.000 vezes a mesma ordem (mesmo EndToEndId) direto no tópico e pela API | Contagem de liquidações no SPI e de lançamentos nos PSPs | exatamente 1 de cada |
| C3 | Desempenho | k6 com 300 pagamentos/s por 10 min, chaves distribuídas entre os dois bancos | `http_duracao_segundos` e tempo até `LIQUIDADO` | p99 fim a fim ≤ 2 s |
| C4 | Resiliência | Toxiproxy corta SPI↔PSP Beta por 60 s sob 100 pagamentos/s e depois restaura | Tempo até toda ordem afetada ficar em estado final | 100% finais em ≤ 30 s após reconectar |
| C5 | Degradação | Derrubar o antifraude sob carga | Estado do circuit breaker; aprovações abaixo do limite configurado | breaker aberto em ≤ 5 s; pagamentos pequenos seguem |
| C6 | Segurança | Requisições sem JWT, com JWT expirado ou de outro emissor; serviço sem certificado; rajada de consultas de chave | Códigos de resposta | 100% recusadas; varredura recebe 429 |
| C7 | Deployment | Publicar versão do antifraude com 5% de erro induzido | Análise do Argo Rollouts | rollback automático, sem intervenção |

## Cenários extras de caos (P03)

- Matar o SPI no meio de uma liquidação (`docker compose kill spi`) e reiniciar: nenhuma ordem liquidada em dobro.
- Matar o relay do Outbox de um PSP: as mensagens saem quando ele voltar, nenhuma se perde.
- Mensagem envenenada em `pix.spi.entrada`: vai para a DLQ e não trava as próximas.
- Pausar o Redpanda por 30 s: pagamentos ficam `PENDENTE` e terminam depois.

## Como rodar o que já existe

```bash
go test -race -cover ./...           # unitários
cp .env.example .env
docker compose up -d --build --wait  # ambiente
./scripts/fumaca.sh                  # fumaça
```
