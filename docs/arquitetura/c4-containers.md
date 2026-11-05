# C4 nível 2: containers

Cada serviço é um processo Go com banco próprio. A comunicação síncrona (HTTP) fica nas consultas que precisam de resposta imediata (chave no DICT, score no antifraude). As etapas do pagamento que mexem em saldo andam por mensagens no Redpanda, gravadas antes num Outbox na mesma transação do saldo (ADR-0001 e ADR-0002).

Caixas claras com borda tracejada (Keycloak e OPA) e as setas tracejadas que chegam nelas entram no Projeto 03. O auditor também consulta os PSPs (seta omitida para não poluir o desenho).

```mermaid
flowchart TB
    cliente["👤 <b>Cliente</b><br/><i>[Pessoa]</i>"]
    operador["👤 <b>Equipe de operação</b><br/><i>[Pessoa]</i>"]

    subgraph borda["Borda"]
        direction LR
        idp["<b>Provedor de identidade</b><br/><i>[Keycloak]</i><br/>OIDC, emite JWT"]
        gateway["<b>API Gateway</b><br/><i>[Envoy]</i><br/>Roteamento, rate limit,<br/>health check ativo"]
        opa["<b>Políticas</b><br/><i>[OPA]</i><br/>Autorização como código"]
    end

    subgraph bancos["Bancos simulados: mesmo código, uma instância por banco"]
        direction LR
        pspA["<b>PSP Banco Alfa</b><br/><i>[Go]</i><br/>Contas, débito e crédito,<br/>Outbox e Inbox"]
        pspB["<b>PSP Banco Beta</b><br/><i>[Go]</i><br/>Contas, débito e crédito,<br/>Outbox e Inbox"]
    end

    subgraph nucleo["Núcleo"]
        direction LR
        dict["<b>DICT</b><br/><i>[Go]</i><br/>Chave Pix → conta"]
        af["<b>Antifraude</b><br/><i>[Go, sem estado]</i><br/>Score de risco"]
        broker[["<b>Mensageria</b><br/><i>[Redpanda / API Kafka]</i><br/>Uma caixa de entrada por<br/>participante + DLQ"]]
        spi["<b>SPI</b><br/><i>[Go]</i><br/>Orquestra a SAGA e liquida<br/>entre contas de reserva"]
        auditor["<b>Auditor</b><br/><i>[Go]</i><br/>Conservação e<br/>idempotência"]
    end

    subgraph dados["Dados: um banco por serviço, cada um numa rede interna"]
        direction LR
        dbA[("<b>BD Alfa</b><br/><i>[PostgreSQL]</i>")]
        dbB[("<b>BD Beta</b><br/><i>[PostgreSQL]</i>")]
        dbD[("<b>BD DICT</b><br/><i>[PostgreSQL]</i>")]
        cache[("<b>Cache</b><br/><i>[Redis]</i>")]
        dbS[("<b>BD SPI</b><br/><i>[PostgreSQL]</i><br/>Reservas e ordens")]
    end

    obs["<b>Observabilidade</b><br/><i>[Prometheus, Grafana;<br/>Jaeger no P03]</i><br/>Coleta /metrics de todos"]

    cliente -- "HTTPS/JSON + JWT" --> gateway
    cliente -. "login" .-> idp
    gateway -. "valida JWT" .-> idp
    gateway -. "ext_authz" .-> opa
    gateway -- "HTTP" --> pspA & pspB

    pspA & pspB -- "consulta chave [HTTP]" --> dict
    pspA & pspB -- "score [HTTP, circuit breaker]" --> af
    pspA & pspB <-- "ordens, aceites, resultados<br/>[Kafka, via Outbox]" --> broker
    broker <-- "pix.spi.entrada /<br/>pix.psp.#lt;ISPB#gt;.entrada" --> spi
    auditor -. "saldos e ordens<br/>[HTTP interno]" .-> spi

    pspA --- dbA
    pspB --- dbB
    dict --- dbD
    dict --- cache
    spi --- dbS

    operador -- "painéis" --> obs
    operador -- "relatório" --> auditor

    classDef pessoa fill:#08427b,stroke:#052e56,color:#fff
    classDef container fill:#438dd5,stroke:#2e6295,color:#fff
    classDef futuro fill:#dbe7f5,stroke:#2e6295,color:#1b3a5c,stroke-dasharray: 5 5
    class cliente,operador pessoa
    class gateway,pspA,pspB,spi,dict,af,broker,auditor,obs,dbA,dbB,dbS,dbD,cache container
    class idp,opa futuro
    style borda fill:#f5f8fc,stroke:#2e6295
    style bancos fill:#f5f8fc,stroke:#2e6295,stroke-dasharray: 3 3
    style nucleo fill:#f5f8fc,stroke:#2e6295
    style dados fill:#f5f8fc,stroke:#2e6295
```

## Containers

| Container | Tecnologia | Responsabilidade | Dados | Exposto na borda? |
|---|---|---|---|---|
| API Gateway | Envoy | Ponto único de entrada: roteamento por banco, rate limit global e por rota, health check ativo, bloqueio de `/metrics` | não | sim (porta 8080) |
| PSP (Alfa, Beta) | Go | Contas de clientes, débito e crédito, Outbox e Inbox; inicia o pagamento | PostgreSQL próprio | sim, via gateway |
| SPI | Go | Recebe ordens, consulta o aceite do recebedor, liquida de forma atômica e notifica os dois lados | PostgreSQL próprio | não |
| DICT | Go | Resolve chave Pix em conta; cache-aside e limite de consultas por PSP | PostgreSQL + Redis | não |
| Antifraude | Go | Score de risco por pagamento; sem estado, escala horizontalmente e recebe canary no P03 | não | não |
| Auditor | Go | Verifica conservação de saldo e ausência de liquidação duplicada | não (lê APIs) | não |
| Redpanda | Redpanda | Tópicos `pix.spi.entrada` e `pix.psp.<ISPB>.entrada`, com DLQ | log próprio | não |
| Observabilidade | Prometheus, Grafana | Coleta `/metrics` de todos os serviços e do gateway | séries temporais | só em 127.0.0.1 |
| Keycloak, OPA | (P03) | Identidade (OIDC/JWT) e autorização como código | — | — |

## Redes do Docker Compose

A segmentação aplica o Zero Trust já no ambiente local (ADR-0003):

| Rede | Quem participa |
|---|---|
| `borda` | gateway, PSPs |
| `nucleo` | PSPs, SPI, DICT, antifraude, auditor, Redpanda |
| `db-psp-alfa`, `db-psp-beta`, `db-spi`, `db-dict` (internas) | cada serviço com o seu banco; o DICT também com o Redis |
| `observabilidade` | todos os serviços, gateway, Prometheus e Grafana |

Um PSP não alcança o banco do outro PSP, e nenhum banco é alcançável pela borda. O teste de fumaça (`scripts/fumaca.sh`) verifica isso.

## Fluxo de um pagamento

```mermaid
sequenceDiagram
    autonumber
    actor C as Cliente
    participant A as PSP Alfa (pagador)
    participant D as DICT
    participant F as Antifraude
    participant S as SPI
    participant B as PSP Beta (recebedor)

    C->>A: POST /v1/pagamentos (chave, valor)
    A->>D: resolve chave
    A->>F: score de risco
    Note over A: 1 transação local: debita o cliente,<br/>pagamento PENDENTE, ordem no Outbox
    A-)S: ordem (pix.spi.entrada)
    Note over S: ordem RECEBIDA
    S-)B: pedido de aceite
    B-)S: ACEITA (ou REJEITADA)
    Note over S: 1 transação ACID: debita reserva Alfa,<br/>credita reserva Beta, ordem LIQUIDADA
    S-)B: liquidado: credita o cliente
    S-)A: liquidado: pagamento LIQUIDADO
    Note over S,A: Rejeição ou falta de aceite no prazo:<br/>ordem REJEITADA e Alfa estorna o cliente
```

O recebedor só credita o cliente depois da notificação de liquidação. Assim um aceite que chega atrasado, depois do prazo, não gera crédito sem dinheiro correspondente na reserva.
