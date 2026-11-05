# ADR-0004: Pipeline DevSecOps com imagens distroless, SBOM e assinatura

- **Status:** Aceito
- **Data:** 2026-11-04
- **Responsável:** Herlan Alef de Lima Nascimento

## Contexto

O projeto tem cinco serviços, três pessoas commitando e uma entrega final que exige código funcional e testado. Ataques à cadeia de suprimentos (dependência comprometida, action do GitHub sequestrada, imagem trocada no registro) são hoje um dos vetores mais comuns, e o enunciado valoriza DevSecOps e supply chain security. O critério C7 prevê canary do antifraude com rollback automático, o que exige imagens versionadas e confiáveis.

## Decisão

**GitHub Actions** em todo push e pull request para a `main` (`.github/workflows/ci.yml`):

| Job | O que faz |
|---|---|
| Testes Go | `gofmt`, `go mod tidy -diff`, `go vet`, testes com `-race` e relatório de cobertura |
| Lint | golangci-lint com `gosec`, `errorlint`, `bodyclose`, `noctx` e outros |
| Segurança | `govulncheck` (vulnerabilidades alcançáveis no código) e Trivy no repositório (dependências, segredos, Dockerfile) |
| Imagens | build de cada serviço (descoberto pelas pastas de `cmd/`), Trivy na imagem e SBOM SPDX com Syft |
| Fumaça | sobe o Docker Compose inteiro e roda `scripts/fumaca.sh` |
| Publicar | só na `main` do repositório oficial: push no GHCR, SBOM e proveniência SLSA anexados, assinatura keyless com Cosign |

**Imagem única e distroless:** um só Dockerfile, com `--build-arg SERVICO`, gera todas as imagens. O binário é estático (`CGO_ENABLED=0`) e a base é `distroless/static:nonroot`: sem shell nem gerenciador de pacotes, com poucos MB e quase nada para o scanner achar. O healthcheck é um subcomando do próprio binário (`/servico healthcheck`), porque não existe curl na imagem.

**Cadeia de suprimentos:**

- actions fixadas por SHA de commit, não por tag (tags podem ser movidas por quem comprometer o repositório da action);
- Dependabot semanal para módulos Go, imagens base e actions;
- Go sempre no patch mais recente da versão (`1.26.x`), não na versão mínima do `go.mod`: a primeira execução do pipeline foi barrada pelo govulncheck justamente por isso (28 vulnerabilidades da biblioteca padrão da 1.26.0);
- permissões mínimas: `contents: read` por padrão; `packages: write` e `id-token: write` só no job de publicação;
- assinatura keyless (Sigstore): sem chave privada para guardar; a identidade é o workflow do repositório, registrada no log de transparência Rekor.

**No Projeto 03:** Argo Rollouts no kind faz o canary do antifraude, com análise automática da taxa de erro no Prometheus (C7), e uma política de admissão só aceita imagens assinadas por este workflow.

## Alternativas consideradas

- **GitLab CI ou Jenkins:** o repositório está no GitHub, e Actions não exige infraestrutura própria.
- **Imagem Alpine ou Debian slim:** traz shell e pacotes, o que aumenta a superfície de ataque e os alertas de CVE sem benefício para um binário estático.
- **Cosign com par de chaves:** a chave privada viraria um segredo a proteger e rotacionar; o keyless usa a identidade OIDC do próprio workflow.
- **Actions por tag (`@v4`):** mais legível, mas confia que a tag não será movida para um commit malicioso.

## Consequências

- **Positivas:** todo PR mostra testes, lint, CVEs e se o ambiente sobe; vulnerabilidade alta ou crítica com correção disponível quebra o build; qualquer pessoa pode verificar a origem de uma imagem com `cosign verify`.
- **Negativas:** o pipeline completo leva alguns minutos; o scanner pode barrar o build por uma CVE nova numa dependência que não mudou (é proposital, e a correção vem pelo Dependabot).
- **Obrigatório:** nada vai para a `main` com o CI vermelho; serviço novo em `cmd/` ganha imagem automaticamente.

## Validação

- O próprio pipeline é a evidência: cada execução fica registrada no GitHub Actions.
- `cosign verify ghcr.io/deivison-costaa/urubu-do-pix/<serviço> --certificate-identity-regexp 'https://github.com/Deivison-Costaa/urubu-do-pix/.*' --certificate-oidc-issuer https://token.actions.githubusercontent.com`
- P03: um deploy de versão com erro induzido no antifraude volta sozinho para a versão anterior (C7).
