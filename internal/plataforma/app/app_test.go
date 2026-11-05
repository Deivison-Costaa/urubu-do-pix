package app_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/app"
	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/config"
)

func env(m map[string]string) config.Busca {
	return func(n string) (string, bool) { v, ok := m[n]; return v, ok }
}

func portaLivre(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().String()
}

func TestConfigFaltandoFalhaNaPartida(t *testing.T) {
	t.Parallel()
	fechou := false
	s := app.Servico{Nome: "spi", Montar: func(_ context.Context, cfg *config.Leitor, _ *slog.Logger) (app.Montagem, error) {
		cfg.Obrigatorio("DATABASE_URL")
		return app.Montagem{Fechar: []func(){func() { fechou = true }}}, nil
	}}
	var saida bytes.Buffer
	if code := app.Executar(context.Background(), nil, env(nil), &saida, s); code != 1 {
		t.Fatalf("código = %d, quer 1", code)
	}
	if !strings.Contains(saida.String(), "DATABASE_URL") {
		t.Fatalf("log não cita a variável ausente: %s", saida.String())
	}
	if !fechou {
		t.Fatal("recursos não foram fechados")
	}
}

func TestErroAoMontar(t *testing.T) {
	t.Parallel()
	s := app.Servico{Nome: "dict", Montar: func(context.Context, *config.Leitor, *slog.Logger) (app.Montagem, error) {
		return app.Montagem{}, errors.New("sem banco")
	}}
	var saida bytes.Buffer
	if code := app.Executar(context.Background(), nil, env(nil), &saida, s); code != 1 {
		t.Fatalf("código = %d, quer 1", code)
	}
}

func TestExecutaEHealthcheck(t *testing.T) {
	t.Parallel()
	endereco := portaLivre(t)
	vars := env(map[string]string{"HTTP_ADDR": endereco, "LOG_LEVEL": "debug"})
	s := app.Servico{Nome: "antifraude", Montar: func(context.Context, *config.Leitor, *slog.Logger) (app.Montagem, error) {
		return app.Montagem{}, nil
	}}

	ctx, cancel := context.WithCancel(context.Background())
	fim := make(chan int, 1)
	var saida bytes.Buffer
	go func() { fim <- app.Executar(ctx, nil, vars, &saida, s) }()

	limite := time.Now().Add(3 * time.Second)
	for app.Healthcheck(context.Background(), endereco) != nil {
		if time.Now().After(limite) {
			t.Fatal("serviço não ficou saudável")
		}
		time.Sleep(20 * time.Millisecond)
	}
	var descarte bytes.Buffer
	if code := app.Executar(context.Background(), []string{app.SubcomandoHealthcheck}, vars, &descarte, s); code != 0 {
		t.Fatalf("subcomando healthcheck = %d, quer 0", code)
	}
	cancel()
	if code := <-fim; code != 0 {
		t.Fatalf("código de saída = %d, quer 0", code)
	}
}

func TestHealthcheckSemServidor(t *testing.T) {
	t.Parallel()
	vars := env(map[string]string{"HTTP_ADDR": portaLivre(t)})
	var saida bytes.Buffer
	if code := app.Executar(context.Background(), []string{app.SubcomandoHealthcheck}, vars, &saida, app.Servico{Nome: "x"}); code != 1 {
		t.Fatalf("healthcheck sem servidor = %d, quer 1", code)
	}
	if err := app.Healthcheck(context.Background(), "sem-porta"); err == nil {
		t.Fatal("endereço inválido deveria falhar")
	}
}

func TestHealthcheckStatusRuim(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}), ReadHeaderTimeout: time.Second}
	go func() { _ = srv.Serve(ln) }()
	defer srv.Close()
	if err := app.Healthcheck(context.Background(), ln.Addr().String()); err == nil {
		t.Fatal("healthz 503 deveria falhar")
	}
}
