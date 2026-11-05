#!/bin/sh
# Contas de exemplo. O total de cada banco é igual ao saldo inicial da sua
# conta de reserva no SPI (deploy/postgres/spi/02-dados.sql); o auditor
# usa essa igualdade para verificar a conservação do dinheiro.
set -eu
case "${PSP_ISPB:?}" in
  10000001) contas="('0001','10001-1','Ana Souza','52998224725',150000),('0001','10002-2','Bruno Lima','11144477735',350000)" ;;
  20000002) contas="('0001','20001-1','Carla Dias','39053344705',200000),('0001','20002-2','Davi Rocha','71428793860',300000)" ;;
  *) echo "sem dados de exemplo para o ISPB $PSP_ISPB"; exit 0 ;;
esac
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<SQL
INSERT INTO contas (agencia, numero, titular_nome, titular_documento, saldo_centavos) VALUES $contas;
SQL
