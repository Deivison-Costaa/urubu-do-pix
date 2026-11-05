# Architecture Decision Records

Cada ADR registra uma decisão: o contexto, o que foi decidido, as alternativas descartadas e as consequências. Um ADR aceito não é editado. Se a decisão mudar, um ADR novo o substitui e o antigo passa a "Substituído por ADR-XXXX".

| ADR | Decisão | Status | Responsável |
|---|---|---|---|
| [0001](0001-saga-orquestrada-pelo-spi.md) | SAGA orquestrada pelo SPI, com Transactional Outbox e Inbox | Aceito | Deivison Costa |
| [0002](0002-mensageria-redpanda.md) | Mensageria com Redpanda, uma caixa de entrada por participante e entrega pelo menos uma vez | Aceito | Giancarlo Silveira Cavalcante |
| [0003](0003-zero-trust-borda-e-rede.md) | Zero Trust: gateway Envoy, segmentação de rede e, no P03, JWT, OPA e mTLS | Aceito | Herlan Alef de Lima Nascimento |
| [0004](0004-pipeline-devsecops.md) | Pipeline DevSecOps com imagens distroless, SBOM e assinatura | Aceito | Herlan Alef de Lima Nascimento |
| [0005](0005-auditor-de-invariantes.md) | Auditor de invariantes fora dos bancos e observabilidade desde o início | Aceito | Giancarlo Silveira Cavalcante |

## Modelo

```markdown
# ADR-XXXX: título no infinitivo ou como decisão

- **Status:** Proposto | Aceito | Substituído por ADR-YYYY
- **Data:** AAAA-MM-DD
- **Responsável:** nome

## Contexto
O problema e as forças em jogo (requisitos, restrições, critérios de sucesso afetados).

## Decisão
O que foi decidido, de forma direta.

## Alternativas consideradas
Cada alternativa e por que não foi escolhida.

## Consequências
O que melhora, o que piora e o que passa a ser obrigatório.

## Validação
Como os testes vão mostrar que a decisão funciona.
```
