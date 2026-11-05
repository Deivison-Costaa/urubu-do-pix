// Comando spi é o Sistema de Pagamentos Instantâneos: liquida as ordens de
// pagamento entre as contas de reserva dos PSPs e orquestra a SAGA.
package main

import (
	"context"
	"log/slog"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/app"
	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/config"
	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/dependencias"
)

func main() {
	app.Principal(app.Servico{Nome: "spi", Montar: montar})
}

func montar(ctx context.Context, cfg *config.Leitor, _ *slog.Logger) (app.Montagem, error) {
	var m app.Montagem
	dbURL := cfg.Obrigatorio("DATABASE_URL")
	brokers := cfg.Lista("KAFKA_BROKERS")
	if err := cfg.Erro(); err != nil {
		return m, err
	}

	pool, vpg, err := dependencias.Postgres(ctx, dbURL)
	if err != nil {
		return m, err
	}
	m.Fechar = append(m.Fechar, pool.Close)

	kafka, vk, err := dependencias.Kafka(brokers, "spi")
	if err != nil {
		return m, err
	}
	m.Fechar = append(m.Fechar, kafka.Close)

	m.Verificacoes = append(m.Verificacoes, vpg, vk)
	return m, nil
}
