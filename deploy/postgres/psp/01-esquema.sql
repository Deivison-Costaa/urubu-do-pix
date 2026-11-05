-- Esquema do PSP (banco simulado). Uma cópia por banco: database-per-service.
-- Valores sempre em centavos (BIGINT), nunca em ponto flutuante.

CREATE TABLE contas (
    id                 BIGSERIAL PRIMARY KEY,
    agencia            CHAR(4)     NOT NULL,
    numero             VARCHAR(20) NOT NULL,
    titular_nome       TEXT        NOT NULL,
    titular_documento  VARCHAR(14) NOT NULL,
    saldo_centavos     BIGINT      NOT NULL CHECK (saldo_centavos >= 0),
    criada_em          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (agencia, numero)
);

-- Estado de cada pagamento visto por este PSP. O EndToEndId é a chave de
-- idempotência: reenvio do mesmo pagamento não cria um segundo registro.
CREATE TABLE pagamentos (
    end_to_end_id   CHAR(32)    NOT NULL,
    direcao         TEXT        NOT NULL CHECK (direcao IN ('ENVIADO', 'RECEBIDO')),
    conta_id        BIGINT      NOT NULL REFERENCES contas (id),
    valor_centavos  BIGINT      NOT NULL CHECK (valor_centavos > 0),
    estado          TEXT        NOT NULL CHECK (estado IN ('PENDENTE', 'LIQUIDADO', 'DEVOLVIDO', 'REJEITADO')),
    criado_em       TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (end_to_end_id, direcao)
);

-- Livro-razão imutável: o saldo é consequência dos lançamentos.
-- A restrição UNIQUE impede debitar ou creditar duas vezes o mesmo pagamento.
CREATE TABLE lancamentos (
    id              BIGSERIAL   PRIMARY KEY,
    conta_id        BIGINT      NOT NULL REFERENCES contas (id),
    end_to_end_id   CHAR(32)    NOT NULL,
    tipo            TEXT        NOT NULL CHECK (tipo IN ('DEBITO', 'CREDITO', 'ESTORNO')),
    valor_centavos  BIGINT      NOT NULL CHECK (valor_centavos > 0),
    criado_em       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (end_to_end_id, tipo)
);

-- Transactional Outbox: a mensagem é gravada na mesma transação do débito
-- e publicada depois por um relay, então débito e mensagem nunca divergem.
CREATE TABLE outbox (
    id            BIGSERIAL   PRIMARY KEY,
    topico        TEXT        NOT NULL,
    chave         TEXT        NOT NULL,
    payload       JSONB       NOT NULL,
    criada_em     TIMESTAMPTZ NOT NULL DEFAULT now(),
    publicada_em  TIMESTAMPTZ
);
CREATE INDEX outbox_pendentes ON outbox (id) WHERE publicada_em IS NULL;

-- Inbox idempotente: mensagens já processadas são ignoradas na reentrega.
CREATE TABLE mensagens_processadas (
    id_mensagem    TEXT        PRIMARY KEY,
    processada_em  TIMESTAMPTZ NOT NULL DEFAULT now()
);
