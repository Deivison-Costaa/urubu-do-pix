// Package app é o ponto de entrada comum dos serviços: lê a configuração,
// monta as dependências, atende HTTP e encerra de forma graciosa ao
// receber SIGINT ou SIGTERM.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/config"
	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/saude"
	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/servidor"
)

// Versao é preenchida no build com -ldflags "-X .../app.Versao=<commit>".
var Versao = "dev"

const (
	enderecoPadrao   = ":8080"
	tempoHealthcheck = 2 * time.Second
	// SubcomandoHealthcheck permite ao Docker testar o serviço sem curl,
	// que não existe na imagem distroless.
	SubcomandoHealthcheck = "healthcheck"
)

// Montagem é o que cada serviço entrega ao executor.
type Montagem struct {
	Verificacoes []saude.Verificacao
	Rotas        func(mux *http.ServeMux)
	// Fechar libera conexões; é chamada em ordem inversa no encerramento.
	Fechar []func()
}

// Servico descreve um serviço executável.
type Servico struct {
	Nome   string
	Montar func(ctx context.Context, cfg *config.Leitor, log *slog.Logger) (Montagem, error)
}

// Principal executa o serviço e encerra o processo com o código de saída.
func Principal(s Servico) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// Depois do primeiro sinal, devolve o tratamento padrão: um segundo
	// Ctrl+C encerra na hora, sem esperar o fim do encerramento gracioso.
	go func() { <-ctx.Done(); cancel() }()
	code := Executar(ctx, os.Args[1:], os.LookupEnv, os.Stdout, s)
	cancel()
	os.Exit(code)
}

// Executar roda o serviço até ctx ser cancelado e devolve o código de saída.
func Executar(ctx context.Context, args []string, busca config.Busca, saida io.Writer, s Servico) int {
	cfg := config.NovoLeitor(busca)
	endereco := cfg.Padrao("HTTP_ADDR", enderecoPadrao)
	log := slog.New(slog.NewJSONHandler(saida, &slog.HandlerOptions{Level: nivelLog(cfg.Padrao("LOG_LEVEL", "info"))})).
		With("servico", s.Nome, "versao", Versao)

	if len(args) > 0 && args[0] == SubcomandoHealthcheck {
		if err := Healthcheck(ctx, endereco); err != nil {
			log.Error("healthcheck falhou", "erro", err)
			return 1
		}
		return 0
	}

	if err := rodar(ctx, cfg, endereco, log, s); err != nil {
		log.Error("serviço encerrado com erro", "erro", err)
		return 1
	}
	log.Info("serviço encerrado")
	return 0
}

func rodar(ctx context.Context, cfg *config.Leitor, endereco string, log *slog.Logger, s Servico) error {
	m, err := s.Montar(ctx, cfg, log)
	defer func() {
		for i := len(m.Fechar) - 1; i >= 0; i-- {
			m.Fechar[i]()
		}
	}()
	if err != nil {
		return fmt.Errorf("montar %s: %w", s.Nome, err)
	}
	if err := cfg.Erro(); err != nil {
		return err
	}
	return servidor.Executar(ctx, servidor.Opcoes{
		Servico:      s.Nome,
		Versao:       Versao,
		Endereco:     endereco,
		Log:          log,
		Verificacoes: m.Verificacoes,
		Rotas:        m.Rotas,
	})
}

// Healthcheck consulta /healthz do próprio serviço em 127.0.0.1.
func Healthcheck(ctx context.Context, endereco string) error {
	_, porta, err := net.SplitHostPort(endereco)
	if err != nil {
		return fmt.Errorf("HTTP_ADDR inválido %q: %w", endereco, err)
	}
	ctx, cancel := context.WithTimeout(ctx, tempoHealthcheck)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+net.JoinHostPort("127.0.0.1", porta)+"/healthz", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("healthz respondeu " + resp.Status)
	}
	return nil
}

func nivelLog(s string) slog.Level {
	var n slog.Level
	if err := n.UnmarshalText([]byte(s)); err != nil {
		return slog.LevelInfo
	}
	return n
}
