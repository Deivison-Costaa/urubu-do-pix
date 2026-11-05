// Comando dict é o Diretório de Identificadores de Contas Transacionais:
// resolve chaves Pix em contas, com cache-aside no Redis.
package main

import (
	"context"
	"log/slog"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/app"
	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/config"
	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/dependencias"
)

func main() {
	app.Principal(app.Servico{Nome: "dict", Montar: montar})
}

func montar(ctx context.Context, cfg *config.Leitor, _ *slog.Logger) (app.Montagem, error) {
	var m app.Montagem
	dbURL := cfg.Obrigatorio("DATABASE_URL")
	redisURL := cfg.Obrigatorio("REDIS_URL")
	if err := cfg.Erro(); err != nil {
		return m, err
	}

	pool, vpg, err := dependencias.Postgres(ctx, dbURL)
	if err != nil {
		return m, err
	}
	m.Fechar = append(m.Fechar, pool.Close)

	cache, vr, err := dependencias.Redis(redisURL)
	if err != nil {
		return m, err
	}
	m.Fechar = append(m.Fechar, func() { _ = cache.Close() })

	m.Verificacoes = append(m.Verificacoes, vpg, vr)
	return m, nil
}
