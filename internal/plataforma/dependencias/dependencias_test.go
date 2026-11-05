package dependencias_test

import (
	"context"
	"testing"
	"time"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/dependencias"
)

// Os testes com bancos de verdade entram no Projeto 03, com Testcontainers.
// Aqui validamos a configuração e que a verificação detecta dependência fora.

func TestURLsInvalidas(t *testing.T) {
	t.Parallel()
	if _, _, err := dependencias.Postgres(context.Background(), "::nao-e-url"); err == nil {
		t.Error("Postgres com URL inválida deveria falhar")
	}
	if _, _, err := dependencias.Redis("http://errado"); err == nil {
		t.Error("Redis com esquema errado deveria falhar")
	}
	if _, _, err := dependencias.Kafka(nil, "teste"); err == nil {
		t.Error("Kafka sem brokers deveria falhar")
	}
}

func TestVerificacoesDetectamDependenciaFora(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, vpg, err := dependencias.Postgres(ctx, "postgres://u:s@127.0.0.1:1/db?connect_timeout=1")
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if vpg.Nome != "postgres" || vpg.Verificar(ctx) == nil {
		t.Error("verificação do Postgres deveria falhar com a porta fechada")
	}

	rcli, vr, err := dependencias.Redis("redis://127.0.0.1:1/0")
	if err != nil {
		t.Fatal(err)
	}
	defer rcli.Close()
	if vr.Nome != "redis" || vr.Verificar(ctx) == nil {
		t.Error("verificação do Redis deveria falhar com a porta fechada")
	}

	kcli, vk, err := dependencias.Kafka([]string{"127.0.0.1:1"}, "teste")
	if err != nil {
		t.Fatal(err)
	}
	defer kcli.Close()
	if vk.Nome != "kafka" || vk.Verificar(ctx) == nil {
		t.Error("verificação do Kafka deveria falhar com a porta fechada")
	}
}
