# Registro de uso de IA

Este arquivo é a base da seção "Ferramentas de IA utilizadas" do README, exigida pelo enunciado. Cada integrante registra aqui o próprio uso, no momento em que acontece: anotar na hora é bem mais fácil do que reconstruir tudo em dezembro.

## Como registrar

Uma entrada por tarefa, da mais recente para a mais antiga:

```markdown
### AAAA-MM-DD: título curto
- **Integrante:**
- **Ferramenta e modelo:**
- **Parte do trabalho:** código, revisão, documentação, testes, análise ou redação
- **Contexto fornecido:** o que foi passado à ferramenta (arquivos, instruções, restrições)
- **Resultado:** o que foi aceito, o que precisou ser corrigido e o que foi descartado, e por quê
```

## Entradas

### 2026-11-13: revisão dos slides 2 e 3 da apresentação do Projeto 02
- **Integrante:** Deivison Costa
- **Ferramenta e modelo:** Claude Code, modelo Claude Fable 5.2
- **Parte do trabalho:** redação (slides da apresentação)
- **Contexto fornecido:** os slides do Projeto 02, a documentação do repositório e o pedido de deixar o slide do problema mais elaborado e as metas dos critérios em palavras simples
- **Resultado:**
  - Aceito: o slide 2 passou a usar um exemplo concreto (Ana, do Alfa, envia R$ 100 para Carla, do Beta) com três falhas possíveis, cada uma resolvida por uma decisão mostrada adiante (Outbox, idempotência, prazo de aceite e estorno); no slide 3, as metas foram reescritas em linguagem simples, com os mesmos números da proposta.
  - Corrigido: o título do slide 3 perdeu o trecho "(proposta aprovada)", por decisão do integrante.
  - Verificação: PDF regerado com o Marp e conferido slide a slide (17 páginas, sem texto cortado).

### 2026-11-04: código-base e documentação do Projeto 02
- **Integrante:** Deivison Costa (pedido e verificação); cada integrante revisou e enviou a parte da própria frente
- **Ferramenta e modelo:** Claude Code, modelo Claude Opus 5.5
- **Parte do trabalho:** código (plataforma comum, tipos de domínio, esqueletos dos cinco serviços), testes unitários, Docker Compose, Envoy, esquemas SQL, pipeline de CI, Dependabot, diagramas C4, ADRs 0001 a 0005, stack, padrões, plano de testes, divisão de responsabilidades e roteiro da apresentação
- **Contexto fornecido:** enunciado, proposta do Projeto 01 e o pedido de fazer tudo o que a entrega exige. Restrições: testes escritos antes da implementação, valores monetários só em inteiros, versões de imagens e actions fixadas, imagens distroless, credenciais fora do repositório, Conventional Commits
- **Resultado:**
  - Aceito: a estrutura do repositório, a plataforma comum, os ADRs e a topologia de tópicos (uma caixa de entrada por participante, como no SPI real).
  - Corrigido durante a verificação: a recuperação de pânico ficava fora da instrumentação e as respostas 500 não entravam nas métricas; o teste de segmentação de rede passava nos dois casos (rede certa e errada) e foi trocado por `pg_isready` com controle positivo; a porta do gateway virou variável porque a 8080 estava ocupada na máquina de desenvolvimento; ajuste de indentação no YAML do Envoy.
  - Corrigido depois de uma revisão de código feita por um segundo agente do Claude Code: o encerramento gracioso cancelava as requisições em andamento (o contexto delas herdava o cancelamento do sinal; agora usa `context.WithoutCancel`, com teste); a interface administrativa do Envoy estava acessível aos PSPs pela rede (agora só no loopback, com um listener à parte só para as métricas); o gateway não tinha healthcheck, o que deixava o teste de fumaça sujeito a corrida no CI; e o CI podia cancelar a publicação na `main` entre o push e a assinatura da imagem. Uma suspeita da revisão (desvio do bloqueio de `/metrics` com caminho codificado) foi testada e não se confirmou.
  - Mudança de desenho em relação à proposta: na proposta, o SPI liquidava e depois devolvia se o recebedor recusasse. No ADR-0001 o SPI pede o aceite antes de liquidar, como no Pix real, e a compensação vira estorno no pagador.
  - Verificação: testes com `-race` (cobertura de 81% a 100% por pacote), golangci-lint sem apontamentos, govulncheck e Trivy sem vulnerabilidades altas, ambiente completo subindo com todos os healthchecks e teste de fumaça passando.
  - Revisão de cada frente pela equipe (cada integrante registra a sua):
    - **Deivison Costa** (liquidação, observabilidade e coordenação): a IA serviu para discutir a ideia do projeto, gerar e revisar o código.
    - **Giancarlo Silveira Cavalcante** (bancos, mensageria e qualidade): *a preencher*
    - **Herlan Alef de Lima Nascimento** (borda, segurança e entrega): *a preencher*

### 2026-10-08: escolha do tema e documentação do Projeto 01
- **Integrante:** Deivison Costa
- **Ferramenta e modelo:** Claude Code, modelo Claude Opus 5.5
- **Parte do trabalho:** análise do enunciado, sugestão de temas, redação da proposta e do README inicial, criação do repositório e identificação dos usuários do GitHub dos integrantes a partir de repositórios em comum
- **Contexto fornecido:** PDF do enunciado e o pedido de ideias mais complexas que as sugestões do professor, seguindo os critérios de avaliação
- **Resultado:**
  - Aceito: o tema de pagamentos instantâneos inspirado no Pix, escolhido entre cinco sugestões. As outras quatro (bolsa de valores com motor de casamento determinístico, central de transplantes, CI/CD com cadeia de suprimentos verificável e rede de recarga de veículos elétricos) foram descartadas.
  - Corrigido: os nomes sugeridos pela IA ("SPI-Sim" e depois "SomaZero") foram trocados por "Urubu do Pix", escolhido pela equipe.
  - Revisão da proposta pela equipe: *a preencher*
