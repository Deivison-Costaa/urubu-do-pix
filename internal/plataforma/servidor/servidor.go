// Package servidor monta o servidor HTTP comum a todos os serviços:
// sondas de saúde, métricas Prometheus, recuperação de pânico e
// encerramento gracioso.
package servidor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/saude"
)

// Tempos do servidor. ReadHeaderTimeout evita conexões lentas (Slowloris).
const (
	tempoCabecalho      = 5 * time.Second
	tempoLeitura        = 15 * time.Second
	tempoEscrita        = 15 * time.Second
	tempoOcioso         = 60 * time.Second
	tempoProntidao      = 2 * time.Second
	tempoEncerramento   = 10 * time.Second
	tamanhoMaxCabecalho = 1 << 20
)

// Opcoes configura o servidor de um serviço.
type Opcoes struct {
	Servico      string
	Versao       string
	Endereco     string
	Log          *slog.Logger
	Verificacoes []saude.Verificacao
	// Rotas registra as rotas de negócio do serviço (opcional).
	Rotas func(mux *http.ServeMux)
}

// NovoHandler monta o roteador com as rotas comuns e as do serviço.
func NovoHandler(op Opcoes) http.Handler {
	log := op.Log
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	registro := prometheus.NewRegistry()
	registro.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	met := novasMetricas(registro, op.Servico, op.Versao)

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", saude.Vivacidade())
	mux.Handle("GET /readyz", saude.Prontidao(log, tempoProntidao, op.Verificacoes...))
	mux.Handle("GET /metrics", promhttp.HandlerFor(registro, promhttp.HandlerOpts{Registry: registro}))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"servico": op.Servico, "versao": op.Versao})
	})
	if op.Rotas != nil {
		op.Rotas(mux)
	}
	return met.instrumentar(recuperar(log, mux))
}

// Executar atende HTTP até ctx ser cancelado e então encerra de forma
// graciosa, esperando as requisições em andamento.
func Executar(ctx context.Context, op Opcoes) error {
	if op.Log == nil {
		op.Log = slog.New(slog.DiscardHandler)
	}
	srv := &http.Server{
		Addr:              op.Endereco,
		Handler:           NovoHandler(op),
		ReadHeaderTimeout: tempoCabecalho,
		ReadTimeout:       tempoLeitura,
		WriteTimeout:      tempoEscrita,
		IdleTimeout:       tempoOcioso,
		MaxHeaderBytes:    tamanhoMaxCabecalho,
		// Sem o cancelamento: o sinal de parada não pode abortar as
		// requisições em andamento, que o Shutdown espera terminar.
		BaseContext: func(net.Listener) context.Context { return context.WithoutCancel(ctx) },
	}
	erros := make(chan error, 1)
	go func() {
		op.Log.Info("servidor HTTP iniciado", "endereco", op.Endereco)
		erros <- srv.ListenAndServe()
	}()

	select {
	case err := <-erros:
		return fmt.Errorf("servidor HTTP: %w", err)
	case <-ctx.Done():
	}
	op.Log.Info("encerrando servidor HTTP")
	fim, cancel := context.WithTimeout(context.Background(), tempoEncerramento)
	defer cancel()
	if err := srv.Shutdown(fim); err != nil {
		return fmt.Errorf("encerrar servidor HTTP: %w", err)
	}
	if err := <-erros; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("servidor HTTP: %w", err)
	}
	return nil
}

func recuperar(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				if err, ok := p.(error); ok && errors.Is(err, http.ErrAbortHandler) {
					panic(p)
				}
				log.ErrorContext(r.Context(), "pânico ao atender requisição", "rota", r.Pattern, "panico", p)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type metricas struct {
	requisicoes *prometheus.CounterVec
	duracao     *prometheus.HistogramVec
}

func novasMetricas(reg prometheus.Registerer, servico, versao string) metricas {
	rotulos := prometheus.Labels{"servico": servico, "versao": versao}
	m := metricas{
		requisicoes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_requisicoes_total",
			Help:        "Requisições HTTP atendidas, por rota, método e status.",
			ConstLabels: rotulos,
		}, []string{"rota", "metodo", "status"}),
		duracao: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:        "http_duracao_segundos",
			Help:        "Duração das requisições HTTP, por rota e método.",
			ConstLabels: rotulos,
			Buckets:     []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2, 5},
		}, []string{"rota", "metodo"}),
	}
	reg.MustRegister(m.requisicoes, m.duracao)
	return m
}

// instrumentar mede cada requisição usando o padrão da rota (r.Pattern),
// e não o caminho bruto, para não explodir a cardinalidade das métricas.
func (m metricas) instrumentar(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		gw := &gravadorStatus{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(gw, r)
		rota := r.Pattern
		if rota == "" {
			rota = "desconhecida"
		}
		m.requisicoes.WithLabelValues(rota, r.Method, strconv.Itoa(gw.status)).Inc()
		m.duracao.WithLabelValues(rota, r.Method).Observe(time.Since(inicio).Seconds())
	})
}

type gravadorStatus struct {
	http.ResponseWriter
	status int
}

func (g *gravadorStatus) WriteHeader(code int) {
	g.status = code
	g.ResponseWriter.WriteHeader(code)
}

func (g *gravadorStatus) Unwrap() http.ResponseWriter { return g.ResponseWriter }
