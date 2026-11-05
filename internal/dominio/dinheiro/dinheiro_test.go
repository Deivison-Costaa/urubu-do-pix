package dinheiro_test

import (
	"errors"
	"math"
	"testing"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/dominio/dinheiro"
)

func TestParse(t *testing.T) {
	t.Parallel()
	casos := []struct {
		nome    string
		entrada string
		quer    dinheiro.Centavos
		erro    error
	}{
		{"valor com centavos", "123.45", 12345, nil},
		{"um centavo", "0.01", 1, nil},
		{"maximo de 10 digitos inteiros", "9999999999.99", 999999999999, nil},
		{"zero e invalido para pagamento", "0.00", 0, dinheiro.ErrValorNaoPositivo},
		{"sem casas decimais", "10", 0, dinheiro.ErrFormato},
		{"uma casa decimal", "10.5", 0, dinheiro.ErrFormato},
		{"virgula como separador", "10,50", 0, dinheiro.ErrFormato},
		{"negativo", "-1.00", 0, dinheiro.ErrFormato},
		{"vazio", "", 0, dinheiro.ErrFormato},
		{"mais de 10 digitos inteiros", "12345678901.00", 0, dinheiro.ErrFormato},
		{"espacos", " 1.00", 0, dinheiro.ErrFormato},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			t.Parallel()
			got, err := dinheiro.Parse(c.entrada)
			if !errors.Is(err, c.erro) {
				t.Fatalf("Parse(%q) erro = %v, quer %v", c.entrada, err, c.erro)
			}
			if got != c.quer {
				t.Fatalf("Parse(%q) = %d, quer %d", c.entrada, got, c.quer)
			}
		})
	}
}

func TestString(t *testing.T) {
	t.Parallel()
	casos := map[dinheiro.Centavos]string{
		1:      "0.01",
		10:     "0.10",
		100:    "1.00",
		12345:  "123.45",
		-12345: "-123.45",
		0:      "0.00",
	}
	for valor, quer := range casos {
		if got := valor.String(); got != quer {
			t.Errorf("Centavos(%d).String() = %q, quer %q", int64(valor), got, quer)
		}
	}
}

func TestIdaEVolta(t *testing.T) {
	t.Parallel()
	for _, s := range []string{"0.01", "1.00", "123.45", "9999999999.99"} {
		v, err := dinheiro.Parse(s)
		if err != nil {
			t.Fatalf("Parse(%q): %v", s, err)
		}
		if v.String() != s {
			t.Errorf("ida e volta de %q virou %q", s, v.String())
		}
	}
}

func TestSomar(t *testing.T) {
	t.Parallel()
	got, err := dinheiro.Somar(150, 250, -100)
	if err != nil || got != 300 {
		t.Fatalf("Somar = %d, %v; quer 300, nil", got, err)
	}
	if _, err := dinheiro.Somar(math.MaxInt64, 1); !errors.Is(err, dinheiro.ErrOverflow) {
		t.Fatalf("Somar com overflow positivo: erro = %v, quer ErrOverflow", err)
	}
	if _, err := dinheiro.Somar(math.MinInt64, -1); !errors.Is(err, dinheiro.ErrOverflow) {
		t.Fatalf("Somar com overflow negativo: erro = %v, quer ErrOverflow", err)
	}
	if got, err := dinheiro.Somar(); err != nil || got != 0 {
		t.Fatalf("Somar() = %d, %v; quer 0, nil", got, err)
	}
}
