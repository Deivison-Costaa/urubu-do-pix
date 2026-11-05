package saude_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/saude"
)

func requisitar(t *testing.T, h http.Handler) (int, saude.Relatorio) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	var r saude.Relatorio
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("corpo não é JSON: %v (%s)", err, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q", ct)
	}
	return rec.Code, r
}

func TestProntidaoTudoOK(t *testing.T) {
	t.Parallel()
	ok := func(context.Context) error { return nil }
	code, r := requisitar(t, saude.Prontidao(nil, time.Second,
		saude.Verificacao{Nome: "postgres", Verificar: ok},
		saude.Verificacao{Nome: "kafka", Verificar: ok},
	))
	if code != http.StatusOK || r.Status != "pronto" {
		t.Fatalf("code=%d status=%q", code, r.Status)
	}
	if r.Dependencias["postgres"] != "ok" || r.Dependencias["kafka"] != "ok" {
		t.Fatalf("dependencias = %v", r.Dependencias)
	}
}

func TestProntidaoSemDependencias(t *testing.T) {
	t.Parallel()
	code, r := requisitar(t, saude.Prontidao(nil, time.Second))
	if code != http.StatusOK || r.Status != "pronto" {
		t.Fatalf("code=%d status=%q", code, r.Status)
	}
}

func TestProntidaoFalha(t *testing.T) {
	t.Parallel()
	code, r := requisitar(t, saude.Prontidao(nil, time.Second,
		saude.Verificacao{Nome: "postgres", Verificar: func(context.Context) error { return nil }},
		saude.Verificacao{Nome: "redis", Verificar: func(context.Context) error { return errors.New("conexão recusada") }},
	))
	if code != http.StatusServiceUnavailable || r.Status != "indisponivel" {
		t.Fatalf("code=%d status=%q", code, r.Status)
	}
	if r.Dependencias["redis"] != "falha" || r.Dependencias["postgres"] != "ok" {
		t.Fatalf("dependencias = %v", r.Dependencias)
	}
}

func TestProntidaoRespeitaTimeout(t *testing.T) {
	t.Parallel()
	lenta := func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
	inicio := time.Now()
	code, _ := requisitar(t, saude.Prontidao(nil, 50*time.Millisecond, saude.Verificacao{Nome: "lenta", Verificar: lenta}))
	if code != http.StatusServiceUnavailable {
		t.Fatalf("code = %d", code)
	}
	if d := time.Since(inicio); d > time.Second {
		t.Fatalf("verificação demorou %v; o timeout não foi aplicado", d)
	}
}

func TestVivacidade(t *testing.T) {
	t.Parallel()
	code, r := requisitar(t, saude.Vivacidade())
	if code != http.StatusOK || r.Status != "vivo" {
		t.Fatalf("code=%d status=%q", code, r.Status)
	}
}
