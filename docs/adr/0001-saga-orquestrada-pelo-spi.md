# ADR-0001: SAGA orquestrada pelo SPI, com Transactional Outbox e Inbox

- **Status:** Aceito
- **Data:** 2026-11-04
- **Responsável:** Deivison Costa

## Contexto

Um pagamento altera três bancos de dados independentes: o do PSP pagador (débito do cliente), o do SPI (contas de reserva) e o do PSP recebedor (crédito do cliente). Instituições diferentes não compartilham transação, e o critério C1 exige que a soma dos saldos seja a mesma antes e depois de qualquer cenário, inclusive com quedas e mensagens duplicadas. O C4 exige que, depois de uma partição de rede, todo pagamento afetado termine em estado final em até 30 s.

Três problemas precisam ser resolvidos juntos:

1. **Atomicidade entre serviços:** sem 2PC, alguém precisa conduzir as etapas e decidir quando desfazer.
2. **Dual write:** gravar o débito e publicar a mensagem são duas operações; se o processo cair entre elas, uma acontece sem a outra.
3. **Reentrega:** o broker entrega pelo menos uma vez (ADR-0002), então cada etapa pode ser executada duas vezes.

## Decisão

**SAGA orquestrada pelo SPI.** O SPI mantém a máquina de estados de cada ordem (`RECEBIDA → LIQUIDADA → CONFIRMADA`, ou `RECEBIDA → REJEITADA`). É também o papel que o SPI tem no Pix real: o liquidante central decide se o pagamento aconteceu.

Etapas (ver o diagrama de sequência em [c4-containers.md](../arquitetura/c4-containers.md)):

1. O PSP pagador, numa transação local, debita o cliente, grava o pagamento `PENDENTE` e grava a ordem no Outbox.
2. O SPI registra a ordem como `RECEBIDA` e pede o aceite ao PSP recebedor.
3. O recebedor valida a conta de destino e responde `ACEITA` ou `REJEITADA`.
4. Com o aceite, o SPI debita a reserva do pagador e credita a do recebedor **numa única transação ACID** no seu banco. A ordem passa a `LIQUIDADA`.
5. O SPI notifica os dois PSPs. O recebedor credita o cliente e o pagador marca o pagamento como `LIQUIDADO`.

**Compensação:** se o recebedor rejeitar ou não responder dentro do prazo, o SPI marca a ordem como `REJEITADA` e o pagador estorna o cliente com um lançamento `ESTORNO`. O recebedor só credita após a notificação de liquidação, então um aceite atrasado nunca gera crédito sem lastro.

**Transactional Outbox:** toda mensagem que sai de um serviço é gravada na tabela `outbox` na mesma transação da mudança de estado. Um relay lê as linhas com `publicada_em IS NULL` e publica no Redpanda. Se o processo cair, o relay republica depois: a mensagem pode duplicar, mas nunca se perde.

**Inbox idempotente:** cada consumidor grava o id da mensagem em `mensagens_processadas` na mesma transação do efeito. As restrições `UNIQUE (end_to_end_id, tipo)` nos lançamentos são a segunda barreira contra débito ou crédito em dobro.

## Alternativas consideradas

- **Two-Phase Commit (XA):** atômico, mas bloqueia recursos enquanto o coordenador decide e trava tudo se o coordenador cair. Não existe entre instituições financeiras reais.
- **SAGA coreografada** (cada serviço reage aos eventos dos outros): menos acoplamento central, mas o estado de um pagamento fica espalhado e o prazo de aceite não tem um dono. Com um liquidante central no domínio, a orquestração é o modelo natural.
- **Publicar direto no broker depois do commit:** simples, mas perde mensagens quando o processo cai entre o commit e a publicação (o problema do dual write).
- **CDC (Debezium) no lugar do relay:** elimina o polling, mas adiciona Kafka Connect e um container pesado. Pode substituir o relay mais tarde sem mudar o esquema do Outbox.

## Consequências

- **Positivas:** a soma das reservas só muda dentro de uma transação ACID do SPI; queda de qualquer serviço atrasa o pagamento, mas não cria nem destrói dinheiro; o estado de cada pagamento pode ser consultado num só lugar.
- **Negativas:** consistência eventual entre os bancos, porque por alguns segundos o pagador já debitou e o recebedor ainda não creditou; o SPI vira ponto central (mitigado por ser sem estado em memória e poder ter réplicas que concorrem pela mesma ordem com `SELECT ... FOR UPDATE`); o relay do Outbox acrescenta latência de polling.
- **Obrigatório:** todo consumidor precisa ser idempotente; nenhuma mensagem pode ser publicada fora do Outbox.

## Validação

- C1: o auditor compara a soma dos saldos antes e depois de cada cenário (ADR-0005).
- C2: 10.000 reenvios do mesmo EndToEndId geram exatamente uma liquidação.
- C4: com o Toxiproxy cortando a rede SPI↔PSP sob carga, todas as ordens chegam a `CONFIRMADA` ou `REJEITADA` em até 30 s após a reconexão.
