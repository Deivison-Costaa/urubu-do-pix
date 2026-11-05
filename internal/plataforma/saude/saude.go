// Package saude implementa as sondas de vivacidade (/healthz) e de
// prontidão (/readyz) usadas pelo Docker Compose e, no Projeto 03, pelo
// Kubernetes.
//
// Vivacidade não consulta dependências: um banco fora do ar não deve fazer
// o orquestrador reiniciar o serviço. Prontidão consulta, para tirar o
// serviço do balanceamento enquanto ele não consegue atender.
package saude

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Verificacao testa uma dependência (banco, broker, cache).
type Verificacao struct {
	Nome      string
	Verificar func(ctx context.Context) error
}

// Relatorio é o corpo JSON das sondas. O motivo das falhas fica no log,
// não na resposta, para não expor detalhes internos.
type Relatorio struct {
	Status       string            `json:"status"`
	Dependencias map[string]string `json:"dependencias,omitempty"`
}

// Vivacidade responde 200 enquanto o processo atende HTTP.
func Vivacidade() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		escrever(w, http.StatusOK, Relatorio{Status: "vivo"})
	})
}

// Prontidao executa as verificações em paralelo, cada uma com timeout, e
// responde 503 se alguma falhar. O motivo da falha vai para log.
func Prontidao(log *slog.Logger, timeout time.Duration, verificacoes ...Verificacao) http.Handler {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		resultados := executar(ctx, verificacoes)
		rel := Relatorio{Status: "pronto", Dependencias: make(map[string]string, len(resultados))}
		code := http.StatusOK
		for nome, err := range resultados {
			if err != nil {
				log.WarnContext(ctx, "dependência indisponível", "dependencia", nome, "erro", err)
				rel.Dependencias[nome] = "falha"
				rel.Status = "indisponivel"
				code = http.StatusServiceUnavailable
				continue
			}
			rel.Dependencias[nome] = "ok"
		}
		escrever(w, code, rel)
	})
}

func executar(ctx context.Context, verificacoes []Verificacao) map[string]error {
	var (
		mu         sync.Mutex
		wg         sync.WaitGroup
		resultados = make(map[string]error, len(verificacoes))
	)
	for _, v := range verificacoes {
		wg.Go(func() {
			err := v.Verificar(ctx)
			mu.Lock()
			resultados[v.Nome] = err
			mu.Unlock()
		})
	}
	wg.Wait()
	return resultados
}

func escrever(w http.ResponseWriter, code int, rel Relatorio) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(rel)
}
