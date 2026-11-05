// Comando antifraude calcula o score de risco de cada pagamento.
// É sem estado (stateless), o que permite escalar réplicas e fazer o
// canary release previsto para o Projeto 03.
package main

import (
	"context"
	"log/slog"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/app"
	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/config"
)

func main() {
	app.Principal(app.Servico{Nome: "antifraude", Montar: montar})
}

func montar(context.Context, *config.Leitor, *slog.Logger) (app.Montagem, error) {
	return app.Montagem{}, nil
}
