# ADR-0002: Mensageria com Redpanda, uma caixa de entrada por participante e entrega pelo menos uma vez

- **Status:** Aceito
- **Data:** 2026-11-04
- **Responsável:** Giancarlo Silveira Cavalcante

## Contexto

As etapas da SAGA (ADR-0001) trocam mensagens entre PSPs e SPI. Precisamos de:

- **durabilidade:** uma mensagem aceita não pode sumir quando um consumidor está fora do ar (C4);
- **ordem por pagamento:** o resultado de um pagamento não pode ser processado antes da ordem dele;
- **releitura:** para reprocessar depois de um bug ou investigar um teste de caos;
- **vazão:** 300 pagamentos/s com p99 ≤ 2 s (C3), numa máquina de desenvolvimento.

## Decisão

Usar o **Redpanda** (compatível com a API do Kafka) com esta topologia:

| Tópico | Quem publica | Quem consome |
|---|---|---|
| `pix.spi.entrada` | PSPs (ordens, aceites) | SPI |
| `pix.psp.<ISPB>.entrada` | SPI (pedidos de aceite, resultados) | o PSP daquele ISPB |
| `<tópico>.dlq` | consumidor, após esgotar os retries | operação (análise manual) |

- **Uma caixa de entrada por participante**, como no SPI real: cada PSP só lê o próprio tópico e não recebe mensagens de outros bancos.
- **Chave da mensagem = EndToEndId:** todas as mensagens de um pagamento caem na mesma partição e chegam em ordem.
- **Entrega pelo menos uma vez:** o consumidor só confirma o offset depois de gravar o efeito; a duplicação é tratada pela Inbox (ADR-0001). Não dependemos de exactly-once do broker.
- **Retry com backoff exponencial e DLQ:** depois de N tentativas a mensagem vai para `<tópico>.dlq` com o motivo do erro, e a partição segue andando.
- **EndToEndId como chave de idempotência em todas as etapas:** gerado pelo PSP pagador no formato do Pix (`E` + ISPB + data e hora UTC + 11 caracteres; `internal/dominio/e2eid`). Ele é chave primária no SPI e parte da chave única dos lançamentos.
- 6 partições por tópico. É o teto de paralelismo por consumidor e pode crescer sem mudar código.

## Alternativas consideradas

- **Apache Kafka (KRaft):** mesma API, mas JVM e mais memória por broker. Como o cliente (franz-go) fala a API do Kafka, trocar o Redpanda pelo Kafka é só configuração.
- **RabbitMQ:** DLQ e retry nativos e ótimo para filas de trabalho, mas sem log persistente para releitura, e a ordem por chave exige uma fila por partição feita à mão.
- **NATS JetStream:** leve e rápido, mas a equipe teria que aprender outro modelo de consumo, e o ecossistema de ferramentas (CDC, conectores) é menor.
- **Um tópico por tipo de evento** (`ordens`, `aceites`, `resultados`): todo PSP leria mensagens de todos os bancos e filtraria no código. Isso vaza dados entre participantes e não segue o modelo do Pix.

## Consequências

- **Positivas:** binário único, sobe em segundos no Compose, com o modo `dev-container`; releitura de qualquer ponto do log para depuração; ordem garantida por pagamento.
- **Negativas:** uma instância só no ambiente local (sem replicação, então a queda do broker para tudo). No Kubernetes (P03) é possível rodar 3 brokers. A ordem é por partição, não global, mas isso não importa para pagamentos independentes.
- **Obrigatório:** todo produtor usa o EndToEndId como chave; todo consumidor é idempotente; nenhuma mensagem com erro bloqueia a partição (vai para a DLQ).

## Validação

- Teste de integração (Testcontainers): reentrega forçada da mesma mensagem 10.000 vezes resulta em um efeito só (C2).
- Teste de resiliência: derrubar o SPI por 60 s sob carga; ao voltar, ele consome o atraso sem perder ordens (C4).
- Mensagem envenenada (payload inválido) termina na DLQ sem travar as seguintes.
