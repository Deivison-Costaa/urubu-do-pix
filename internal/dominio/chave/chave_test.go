package chave_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/dominio/chave"
)

func TestValidar(t *testing.T) {
	t.Parallel()
	casos := []struct {
		nome  string
		tipo  chave.Tipo
		valor string
		ok    bool
	}{
		{"cpf valido", chave.CPF, "52998224725", true},
		{"cpf digito errado", chave.CPF, "52998224724", false},
		{"cpf com mascara", chave.CPF, "529.982.247-25", false},
		{"cpf repetido", chave.CPF, "11111111111", false},
		{"cpf curto", chave.CPF, "5299822472", false},
		{"cnpj valido", chave.CNPJ, "11222333000181", true},
		{"cnpj digito errado", chave.CNPJ, "11222333000182", false},
		{"cnpj repetido", chave.CNPJ, "00000000000000", false},
		{"email valido", chave.Email, "urubu@pix.dev", true},
		{"email sem arroba", chave.Email, "urubu.pix.dev", false},
		{"email longo demais", chave.Email, strings.Repeat("a", 70) + "@pix.dev", false},
		{"email com maiuscula", chave.Email, "Urubu@pix.dev", false},
		{"telefone valido", chave.Telefone, "+5585999998888", true},
		{"telefone sem mais", chave.Telefone, "5585999998888", false},
		{"telefone com espaco", chave.Telefone, "+55 85999998888", false},
		{"evp valida", chave.EVP, "123e4567-e89b-42d3-a456-426614174000", true},
		{"evp maiuscula", chave.EVP, "123E4567-E89B-42D3-A456-426614174000", false},
		{"evp sem hifens", chave.EVP, "123e4567e89b42d3a456426614174000", false},
		{"tipo desconhecido", chave.Tipo("BOLETO"), "qualquer", false},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			t.Parallel()
			err := chave.Validar(c.tipo, c.valor)
			if (err == nil) != c.ok {
				t.Fatalf("Validar(%s, %q) = %v, quer ok=%v", c.tipo, c.valor, err, c.ok)
			}
			if err != nil && !errors.Is(err, chave.ErrInvalida) {
				t.Fatalf("erro deveria embrulhar ErrInvalida: %v", err)
			}
		})
	}
}

func TestParseTipo(t *testing.T) {
	t.Parallel()
	for _, s := range []string{"CPF", "CNPJ", "EMAIL", "PHONE", "EVP"} {
		if _, err := chave.ParseTipo(s); err != nil {
			t.Errorf("ParseTipo(%q) = %v", s, err)
		}
	}
	if _, err := chave.ParseTipo("cpf"); err == nil {
		t.Error("ParseTipo deveria ser sensivel a caixa")
	}
}
