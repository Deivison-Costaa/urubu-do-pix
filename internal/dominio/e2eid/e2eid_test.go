package e2eid_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/dominio/e2eid"
)

var instante = time.Date(2026, 11, 4, 13, 7, 59, 0, time.FixedZone("BRT", -3*3600))

func TestNovoFormato(t *testing.T) {
	t.Parallel()
	id, err := e2eid.Novo("10000001", instante, bytes.NewReader(bytes.Repeat([]byte{0}, 64)))
	if err != nil {
		t.Fatalf("Novo: %v", err)
	}
	// E + ISPB + AAAAMMDDHHmm em UTC + 11 caracteres alfanuméricos.
	quer := "E10000001202611041607" + "aaaaaaaaaaa"
	if string(id) != quer {
		t.Fatalf("Novo = %q, quer %q", id, quer)
	}
	if len(id) != e2eid.Tamanho {
		t.Fatalf("tamanho = %d, quer %d", len(id), e2eid.Tamanho)
	}
	if err := e2eid.Validar(string(id)); err != nil {
		t.Fatalf("Validar(Novo()) = %v", err)
	}
}

func TestNovoGeraIdsDistintos(t *testing.T) {
	t.Parallel()
	vistos := make(map[e2eid.ID]struct{})
	for range 1000 {
		id, err := e2eid.Novo("10000001", instante, nil)
		if err != nil {
			t.Fatalf("Novo: %v", err)
		}
		if _, repetido := vistos[id]; repetido {
			t.Fatalf("id repetido: %s", id)
		}
		vistos[id] = struct{}{}
	}
}

func TestNovoISPBInvalido(t *testing.T) {
	t.Parallel()
	for _, ispb := range []string{"", "1234567", "123456789", "1234567a"} {
		if _, err := e2eid.Novo(ispb, instante, nil); !errors.Is(err, e2eid.ErrISPB) {
			t.Errorf("Novo(%q) erro = %v, quer ErrISPB", ispb, err)
		}
	}
}

func TestNovoFonteAleatoriaFalha(t *testing.T) {
	t.Parallel()
	if _, err := e2eid.Novo("10000001", instante, strings.NewReader("")); err == nil {
		t.Fatal("Novo com fonte vazia deveria falhar")
	}
}

func TestValidar(t *testing.T) {
	t.Parallel()
	casos := []struct {
		nome string
		id   string
		ok   bool
	}{
		{"valido", "E10000001202611041607abcDEF12345", true},
		{"prefixo errado", "D10000001202611041607abcDEF12345", false},
		{"curto", "E10000001202611041607abc", false},
		{"longo", "E10000001202611041607abcDEF123456", false},
		{"ISPB com letra", "E1000000a202611041607abcDEF12345", false},
		{"mes 13", "E10000001202613041607abcDEF12345", false},
		{"sequencial com simbolo", "E10000001202611041607abc-EF12345", false},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			t.Parallel()
			err := e2eid.Validar(c.id)
			if (err == nil) != c.ok {
				t.Fatalf("Validar(%q) = %v, quer ok=%v", c.id, err, c.ok)
			}
		})
	}
}

func TestISPBDoID(t *testing.T) {
	t.Parallel()
	ispb, err := e2eid.ISPB("E10000001202611041607abcDEF12345")
	if err != nil || ispb != "10000001" {
		t.Fatalf("ISPB = %q, %v; quer 10000001, nil", ispb, err)
	}
	if _, err := e2eid.ISPB("invalido"); err == nil {
		t.Fatal("ISPB de id invalido deveria falhar")
	}
}
