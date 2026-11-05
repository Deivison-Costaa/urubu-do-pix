# syntax=docker/dockerfile:1
# Imagem única para todos os serviços: SERVICO escolhe qual binário compilar.
# docker build -f deploy/docker/Dockerfile --build-arg SERVICO=spi .

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY cmd ./cmd
COPY internal ./internal
ARG SERVICO
ARG VERSAO=dev
RUN test -n "$SERVICO" || (echo "defina --build-arg SERVICO" && exit 1)
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w -X github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/app.Versao=${VERSAO}" \
      -o /out/servico ./cmd/${SERVICO}

# Distroless: sem shell nem gerenciador de pacotes, roda como usuário sem privilégio.
FROM gcr.io/distroless/static-debian13:nonroot
ARG SERVICO
LABEL org.opencontainers.image.source="https://github.com/Deivison-Costaa/urubu-do-pix" \
      org.opencontainers.image.title="urubu-do-pix-${SERVICO}"
COPY --from=build /out/servico /servico
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=5 CMD ["/servico", "healthcheck"]
ENTRYPOINT ["/servico"]
