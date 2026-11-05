// Comando auditor verifica os invariantes do sistema (conservação de saldo
// e idempotência) consultando as APIs internas dos PSPs e do SPI.
// Não lê os bancos dos outros serviços, para respeitar database-per-service.
package main

import (
	"context"
	"log/slog"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/app"
	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/config"
)

func main() {
	app.Principal(app.Servico{Nome: "auditor", Montar: montar})
}

func montar(_ context.Context, cfg *config.Leitor, log *slog.Logger) (app.Montagem, error) {
	alvos := cfg.Lista("AUDITOR_ALVOS")
	if err := cfg.Erro(); err != nil {
		return app.Montagem{}, err
	}
	log.Info("auditor configurado", "alvos", alvos)
	return app.Montagem{}, nil
}
