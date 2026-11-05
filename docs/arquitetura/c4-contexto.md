# C4 nível 1: contexto

O Urubu do Pix simula o ecossistema Pix inteiro numa só implantação: os bancos (PSPs), o liquidante central (SPI), o diretório de chaves (DICT) e o antifraude. Por isso o sistema aparece como uma caixa só neste nível. Quem está fora dele são as pessoas que pagam e recebem, a equipe que opera e audita, e as ferramentas de teste e de entrega.

```mermaid
flowchart TB
    pagador["👤 <b>Cliente pagador</b><br/><i>[Pessoa]</i><br/>Envia Pix pelo app<br/>do seu banco"]
    recebedor["👤 <b>Cliente recebedor</b><br/><i>[Pessoa]</i><br/>Recebe o crédito e<br/>consulta o extrato"]
    operador["👤 <b>Equipe de operação</b><br/><i>[Pessoa]</i><br/>Acompanha métricas e<br/>verifica os invariantes"]

    sistema["<b>Urubu do Pix</b><br/><i>[Sistema de software]</i><br/>Pagamentos instantâneos entre bancos<br/>simulados, liquidados em segundos sem<br/>criar nem perder nenhum centavo"]

    testes["<b>Ferramentas de teste</b><br/><i>[Sistema externo]</i><br/>k6 (carga) e Toxiproxy<br/>(falhas de rede)"]
    github["<b>GitHub</b><br/><i>[Sistema externo]</i><br/>Actions (CI/CD) e GHCR<br/>(imagens assinadas)"]

    pagador -- "Envia pagamentos e consulta chaves<br/>[HTTPS/JSON]" --> sistema
    sistema -- "Credita e notifica" --> recebedor
    operador -- "Painéis, alertas e relatório<br/>do auditor [HTTPS]" --> sistema
    testes -- "Gera carga e injeta falhas" --> sistema
    github -- "Publica imagens<br/>verificadas" --> sistema

    classDef pessoa fill:#08427b,stroke:#052e56,color:#fff
    classDef sistema fill:#1168bd,stroke:#0b4884,color:#fff
    classDef externo fill:#999,stroke:#6b6b6b,color:#fff
    class pagador,recebedor,operador pessoa
    class sistema sistema
    class testes,github externo
```

## Elementos

| Elemento | Tipo | Responsabilidade |
|---|---|---|
| Cliente pagador | Pessoa | Inicia o pagamento informando a chave Pix do recebedor e o valor. |
| Cliente recebedor | Pessoa | Recebe o crédito na conta do seu banco. |
| Equipe de operação | Pessoa | Acompanha latência, erros e filas; roda o auditor depois de cada teste. |
| Urubu do Pix | Sistema | Resolve a chave, avalia o risco, debita, liquida e credita, mantendo a soma dos saldos. |
| Ferramentas de teste | Externo | Medem os critérios de sucesso C3 (desempenho) e C4 (resiliência). |
| GitHub | Externo | Executa o pipeline e guarda as imagens assinadas que o ambiente usa. |

O nível 2 abre a caixa do sistema: [c4-containers.md](c4-containers.md).
