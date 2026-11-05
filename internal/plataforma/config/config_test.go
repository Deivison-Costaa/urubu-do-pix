package config_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/config"
)

func fonte(m map[string]string) config.Busca {
	return func(n string) (string, bool) { v, ok := m[n]; return v, ok }
}

func TestLeitor(t *testing.T) {
	t.Parallel()
	l := config.NovoLeitor(fonte(map[string]string{
		"A":     "valor",
		"VAZIO": "  ",
		"LISTA": "x:1, y:2 ,,",
	}))
	if got := l.Obrigatorio("A"); got != "valor" {
		t.Errorf("Obrigatorio(A) = %q", got)
	}
	if got := l.Padrao("NAO_EXISTE", "padrao"); got != "padrao" {
		t.Errorf("Padrao = %q", got)
	}
	if got := l.Padrao("A", "padrao"); got != "valor" {
		t.Errorf("Padrao com valor = %q", got)
	}
	if got := l.Lista("LISTA"); !reflect.DeepEqual(got, []string{"x:1", "y:2"}) {
		t.Errorf("Lista = %#v", got)
	}
	if err := l.Erro(); err != nil {
		t.Fatalf("Erro() = %v, quer nil", err)
	}
	l.Obrigatorio("FALTA")
	l.Obrigatorio("VAZIO")
	l.Lista("OUTRA")
	err := l.Erro()
	if !errors.Is(err, config.ErrFaltando) {
		t.Fatalf("Erro() = %v, quer ErrFaltando", err)
	}
	if want := "variáveis de ambiente obrigatórias ausentes: FALTA, VAZIO, OUTRA"; err.Error() != want {
		t.Fatalf("mensagem = %q, quer %q", err.Error(), want)
	}
}
