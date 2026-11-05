package servidor_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"log/slog"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/saude"
	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/servidor"
)

func get(t *testing.T, h http.Handler, caminho string) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, caminho, nil))
	return rec.Code, rec.Body.String()
}

func TestRotasComuns(t *testing.T) {
	t.Parallel()
	h := servidor.NovoHandler(servidor.Opcoes{
		Servico: "spi",
		Versao:  "teste",
		Verificacoes: []saude.Verificacao{
			{Nome: "postgres", Verificar: func(context.Context) error { return errors.New("fora") }},
		},
		Rotas: func(mux *http.ServeMux) {
			mux.HandleFunc("GET /v1/ping", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "pong") })
			mux.HandleFunc("GET /v1/panico", func(http.ResponseWriter, *http.Request) { panic("bug") })
		},
	})

	if code, _ := get(t, h, "/healthz"); code != http.StatusOK {
		t.Errorf("/healthz = %d", code)
	}
	if code, _ := get(t, h, "/readyz"); code != http.StatusServiceUnavailable {
		t.Errorf("/readyz com dependência fora = %d", code)
	}
	if code, body := get(t, h, "/"); code != http.StatusOK || !strings.Contains(body, `"servico":"spi"`) {
		t.Errorf("/ = %d %s", code, body)
	}
	if code, body := get(t, h, "/v1/ping"); code != http.StatusOK || body != "pong" {
		t.Errorf("/v1/ping = %d %s", code, body)
	}
	if code, _ := get(t, h, "/v1/panico"); code != http.StatusInternalServerError {
		t.Errorf("/v1/panico = %d, quer 500", code)
	}
	if code, _ := get(t, h, "/nao-existe"); code != http.StatusNotFound {
		t.Errorf("/nao-existe = %d", code)
	}

	code, body := get(t, h, "/metrics")
	if code != http.StatusOK {
		t.Fatalf("/metrics = %d", code)
	}
	for _, quer := range []string{
		`http_requisicoes_total{metodo="GET",rota="GET /v1/ping",servico="spi",status="200",versao="teste"} 1`,
		`rota="desconhecida"`,
		`status="500"`,
		"go_goroutines",
	} {
		if !strings.Contains(body, quer) {
			t.Errorf("/metrics não contém %q", quer)
		}
	}
}

func TestExecutarEncerraGraciosamente(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	endereco := ln.Addr().String()
	_ = ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	fim := make(chan error, 1)
	go func() {
		fim <- servidor.Executar(ctx, servidor.Opcoes{Servico: "teste", Endereco: endereco, Log: slog.New(slog.DiscardHandler)})
	}()

	if err := esperarPronto("http://" + endereco + "/healthz"); err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case err := <-fim:
		if err != nil {
			t.Fatalf("Executar = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Executar não encerrou após o cancelamento")
	}
}

func TestExecutarEnderecoOcupado(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	err = servidor.Executar(context.Background(), servidor.Opcoes{Endereco: ln.Addr().String(), Log: slog.New(slog.DiscardHandler)})
	if err == nil {
		t.Fatal("Executar em porta ocupada deveria falhar")
	}
}

func esperarPronto(url string) error {
	limite := time.Now().Add(3 * time.Second)
	for time.Now().Before(limite) {
		resp, err := http.Get(url) //nolint:noctx // teste local
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return errors.New("servidor não respondeu a tempo")
}

// Ao receber o sinal de parada, requisições em andamento terminam normalmente:
// o contexto delas não pode ser cancelado junto com o do processo.
func TestEncerramentoEsperaRequisicaoEmAndamento(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	endereco := ln.Addr().String()
	_ = ln.Close()

	começou := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	fim := make(chan error, 1)
	go func() {
		fim <- servidor.Executar(ctx, servidor.Opcoes{
			Servico:  "teste",
			Endereco: endereco,
			Rotas: func(mux *http.ServeMux) {
				mux.HandleFunc("GET /lenta", func(w http.ResponseWriter, r *http.Request) {
					close(começou)
					select {
					case <-time.After(300 * time.Millisecond):
						_, _ = io.WriteString(w, "concluida")
					case <-r.Context().Done():
						http.Error(w, "cancelada", http.StatusServiceUnavailable)
					}
				})
			},
		})
	}()
	if err := esperarPronto("http://" + endereco + "/healthz"); err != nil {
		t.Fatal(err)
	}

	resposta := make(chan string, 1)
	go func() {
		resp, err := http.Get("http://" + endereco + "/lenta") //nolint:noctx // teste local
		if err != nil {
			resposta <- "erro: " + err.Error()
			return
		}
		defer resp.Body.Close()
		corpo, _ := io.ReadAll(resp.Body)
		resposta <- strings.TrimSpace(string(corpo))
	}()
	<-começou
	cancel()

	if got := <-resposta; got != "concluida" {
		t.Fatalf("requisição em andamento terminou com %q, quer %q", got, "concluida")
	}
	if err := <-fim; err != nil {
		t.Fatalf("Executar = %v", err)
	}
}
