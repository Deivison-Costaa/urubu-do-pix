-- Esquema do SPI: contas de reserva dos PSPs e ordens de pagamento.

CREATE TABLE contas_reserva (
    ispb            CHAR(8)     PRIMARY KEY,
    nome            TEXT        NOT NULL,
    saldo_centavos  BIGINT      NOT NULL CHECK (saldo_centavos >= 0),
    atualizada_em   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Uma linha por pagamento: o EndToEndId como chave primária garante que
-- cada ordem é liquidada no máximo uma vez.
CREATE TABLE ordens (
    end_to_end_id    CHAR(32)    PRIMARY KEY,
    ispb_pagador     CHAR(8)     NOT NULL REFERENCES contas_reserva (ispb),
    ispb_recebedor   CHAR(8)     NOT NULL REFERENCES contas_reserva (ispb),
    valor_centavos   BIGINT      NOT NULL CHECK (valor_centavos > 0),
    estado           TEXT        NOT NULL CHECK (estado IN ('RECEBIDA', 'LIQUIDADA', 'CONFIRMADA', 'DEVOLVIDA', 'REJEITADA')),
    recebida_em      TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizada_em    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (ispb_pagador <> ispb_recebedor)
);

-- Lançamentos nas contas de reserva. Cada liquidação gera um débito e um
-- crédito do mesmo valor na mesma transação ACID: a soma não muda.
CREATE TABLE lancamentos_reserva (
    id              BIGSERIAL   PRIMARY KEY,
    end_to_end_id   CHAR(32)    NOT NULL REFERENCES ordens (end_to_end_id),
    ispb            CHAR(8)     NOT NULL REFERENCES contas_reserva (ispb),
    tipo            TEXT        NOT NULL CHECK (tipo IN ('DEBITO', 'CREDITO')),
    motivo          TEXT        NOT NULL CHECK (motivo IN ('LIQUIDACAO', 'DEVOLUCAO')),
    valor_centavos  BIGINT      NOT NULL CHECK (valor_centavos > 0),
    criado_em       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (end_to_end_id, ispb, tipo, motivo)
);

CREATE TABLE outbox (
    id            BIGSERIAL   PRIMARY KEY,
    topico        TEXT        NOT NULL,
    chave         TEXT        NOT NULL,
    payload       JSONB       NOT NULL,
    criada_em     TIMESTAMPTZ NOT NULL DEFAULT now(),
    publicada_em  TIMESTAMPTZ
);
CREATE INDEX outbox_pendentes ON outbox (id) WHERE publicada_em IS NULL;

CREATE TABLE mensagens_processadas (
    id_mensagem    TEXT        PRIMARY KEY,
    processada_em  TIMESTAMPTZ NOT NULL DEFAULT now()
);
