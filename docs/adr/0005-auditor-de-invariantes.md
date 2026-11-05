# ADR-0005: Auditor de invariantes fora dos bancos e observabilidade desde o início

- **Status:** Aceito
- **Data:** 2026-11-04
- **Responsável:** Giancarlo Silveira Cavalcante

## Contexto

Os critérios C1 (conservação), C2 (idempotência) e C4 (resiliência) não podem ser verificados olhando um serviço de cada vez: o dinheiro está espalhado em quatro bancos de dados e passa por estados intermediários. Também precisamos medir latência e erros (C3, C5, C7) com o sistema sob carga e sob falha, e a avaliação final usa dados de testes.

## Decisão

**Um serviço auditor** verifica os invariantes consultando APIs internas de leitura dos PSPs e do SPI, nunca lendo os bancos diretamente:

1. **Conservação global:** a soma das contas de reserva no SPI é constante.
2. **Conservação por banco, em repouso:** para cada PSP, soma dos saldos dos clientes = saldo da reserva no SPI. Os dados iniciais já são montados assim (R$ 5.000,00 por banco).
3. **Conservação em trânsito:** durante a carga, `clientes + pagamentos enviados pendentes = reserva` no pagador, e `clientes + liquidados ainda não creditados = reserva` no recebedor.
4. **Idempotência:** nenhum EndToEndId aparece mais de uma vez como liquidação no SPI nem tem dois lançamentos do mesmo tipo num PSP.
5. **Finalização:** nenhuma ordem fica em estado não final por mais que o prazo da SAGA mais 30 s (C4).

O auditor roda ao fim de cada cenário de teste e devolve um relatório JSON com o resultado de cada invariante. O cenário falha se qualquer invariante falhar.

**Observabilidade desde o primeiro commit:**

- todo serviço expõe `/metrics` (Prometheus) com contagem e histograma de latência por rota. O rótulo usa o padrão da rota (`GET /v1/chaves/{chave}`), não o caminho bruto, para não explodir a cardinalidade;
- sondas separadas: `/healthz` (o processo está vivo; não consulta dependências) e `/readyz` (as dependências respondem). Um banco fora do ar tira o serviço do balanceamento, mas não faz o orquestrador reiniciá-lo em loop;
- logs JSON estruturados (`log/slog`) com serviço e versão em cada linha; o gateway gera `x-request-id`;
- Prometheus e Grafana no perfil `observabilidade` do Compose. Jaeger com OpenTelemetry entra no P03 para rastrear um pagamento de ponta a ponta.

## Alternativas consideradas

- **Auditor lendo os bancos direto:** mais simples, mas acopla o auditor aos esquemas de todos os serviços e quebra o database-per-service. A API de leitura também é o que um regulador real usaria.
- **Verificar invariantes dentro de cada serviço:** um serviço não enxerga o dinheiro dos outros; a conservação é propriedade do sistema inteiro.
- **Event sourcing do razão inteiro para auditar pelo log:** dá auditoria completa, mas muda o modelo de dados de todos os serviços. Os lançamentos imutáveis nas tabelas já dão rastreabilidade suficiente.
- **Só logs, sem métricas:** não dá p99 nem taxa de erro sob carga, que são critérios de sucesso.

## Consequências

- **Positivas:** os critérios de sucesso viram testes automáticos e repetíveis; qualquer cenário de caos termina com um veredito objetivo; as métricas existem antes da primeira linha de regra de negócio.
- **Negativas:** PSPs e SPI precisam expor APIs internas de leitura (que não podem ser expostas na borda, ADR-0003); o invariante em trânsito exige consistência de leitura entre consultas, por isso a verificação forte é feita em repouso, depois de drenar as filas.
- **Obrigatório:** toda rota nova tem métrica automática; todo cenário de teste termina chamando o auditor.

## Validação

- O teste de fumaça já verifica o invariante 2 nos dados iniciais.
- P03: cenários de carga e de caos (Toxiproxy, queda de containers) terminam com o relatório do auditor anexado aos resultados.
