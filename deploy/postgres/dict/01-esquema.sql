-- Esquema do DICT: vínculo entre chave Pix e conta.
CREATE TABLE chaves (
    chave              VARCHAR(77) PRIMARY KEY,
    tipo               TEXT        NOT NULL CHECK (tipo IN ('CPF', 'CNPJ', 'EMAIL', 'PHONE', 'EVP')),
    ispb               CHAR(8)     NOT NULL,
    agencia            CHAR(4)     NOT NULL,
    conta              VARCHAR(20) NOT NULL,
    titular_nome       TEXT        NOT NULL,
    titular_documento  VARCHAR(14) NOT NULL,
    criada_em          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX chaves_por_conta ON chaves (ispb, agencia, conta);

INSERT INTO chaves (chave, tipo, ispb, agencia, conta, titular_nome, titular_documento) VALUES
    ('52998224725',     'CPF',   '10000001', '0001', '10001-1', 'Ana Souza',  '52998224725'),
    ('bruno@alfa.dev',  'EMAIL', '10000001', '0001', '10002-2', 'Bruno Lima', '11144477735'),
    ('+5585999990001',  'PHONE', '20000002', '0001', '20001-1', 'Carla Dias', '39053344705'),
    ('71428793860',     'CPF',   '20000002', '0001', '20002-2', 'Davi Rocha', '71428793860');
