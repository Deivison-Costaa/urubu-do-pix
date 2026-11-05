// Comando psp é o banco simulado (Provedor de Serviços de Pagamento).
// O mesmo binário roda uma instância por banco, cada uma com o seu ISPB e o
// seu próprio Postgres (database-per-service; implantação single-tenant).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/app"
	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/config"
	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/dependencias"
)

var formatoISPB = regexp.MustCompile(`^\d{8}$`)

func main() {
	app.Principal(app.Servico{Nome: "psp", Montar: montar})
}

func montar(ctx context.Context, cfg *config.Leitor, log *slog.Logger) (app.Montagem, error) {
	var m app.Montagem
	ispb := cfg.Obrigatorio("PSP_ISPB")
	nome := cfg.Obrigatorio("PSP_NOME")
	dbURL := cfg.Obrigatorio("DATABASE_URL")
	brokers := cfg.Lista("KAFKA_BROKERS")
	if err := cfg.Erro(); err != nil {
		return m, err
	}
	if !formatoISPB.MatchString(ispb) {
		return m, fmt.Errorf("PSP_ISPB deve ter 8 dígitos: %q", ispb)
	}

	pool, vpg, err := dependencias.Postgres(ctx, dbURL)
	if err != nil {
		return m, err
	}
	m.Fechar = append(m.Fechar, pool.Close)

	kafka, vk, err := dependencias.Kafka(brokers, "psp-"+ispb)
	if err != nil {
		return m, err
	}
	m.Fechar = append(m.Fechar, kafka.Close)

	m.Verificacoes = append(m.Verificacoes, vpg, vk)
	log.Info("PSP configurado", "ispb", ispb, "nome", nome)
	return m, nil
}
