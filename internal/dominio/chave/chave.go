// Package chave valida chaves Pix nos cinco formatos aceitos pelo DICT.
package chave

import (
	"errors"
	"fmt"
	"regexp"
)

// Tipo é o tipo de uma chave Pix, com os nomes usados pela API do DICT.
type Tipo string

// Tipos de chave Pix.
const (
	CPF      Tipo = "CPF"
	CNPJ     Tipo = "CNPJ"
	Email    Tipo = "EMAIL"
	Telefone Tipo = "PHONE"
	EVP      Tipo = "EVP" // chave aleatória
)

// ErrInvalida indica chave fora do formato do seu tipo.
var ErrInvalida = errors.New("chave Pix inválida")

const tamanhoMaxEmail = 77

var (
	soDigitos11 = regexp.MustCompile(`^\d{11}$`)
	soDigitos14 = regexp.MustCompile(`^\d{14}$`)
	formatoTel  = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	formatoMail = regexp.MustCompile(`^[a-z0-9.!#$%&'*+/=?^_{|}~-]+@[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$`)
	formatoEVP  = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// ParseTipo converte o nome do tipo (sensível a caixa) em Tipo.
func ParseTipo(s string) (Tipo, error) {
	switch t := Tipo(s); t {
	case CPF, CNPJ, Email, Telefone, EVP:
		return t, nil
	default:
		return "", fmt.Errorf("%w: tipo desconhecido %q", ErrInvalida, s)
	}
}

// Validar verifica se valor é uma chave válida do tipo informado.
func Validar(tipo Tipo, valor string) error {
	var ok bool
	switch tipo {
	case CPF:
		ok = soDigitos11.MatchString(valor) && digitosVerificadoresOK(valor, pesosCPF)
	case CNPJ:
		ok = soDigitos14.MatchString(valor) && digitosVerificadoresOK(valor, pesosCNPJ)
	case Email:
		ok = len(valor) <= tamanhoMaxEmail && formatoMail.MatchString(valor)
	case Telefone:
		ok = formatoTel.MatchString(valor)
	case EVP:
		ok = formatoEVP.MatchString(valor)
	default:
		return fmt.Errorf("%w: tipo desconhecido %q", ErrInvalida, tipo)
	}
	if !ok {
		return fmt.Errorf("%w: %s %q", ErrInvalida, tipo, valor)
	}
	return nil
}

// Pesos do módulo 11 para cada dígito verificador.
var (
	pesosCPF = [2][]int{
		{10, 9, 8, 7, 6, 5, 4, 3, 2},
		{11, 10, 9, 8, 7, 6, 5, 4, 3, 2},
	}
	pesosCNPJ = [2][]int{
		{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2},
		{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2},
	}
)

// digitosVerificadoresOK confere os dois dígitos finais pelo módulo 11.
// Sequências de um só dígito ("111...") passam no cálculo, mas são inválidas.
func digitosVerificadoresOK(s string, pesos [2][]int) bool {
	if todosIguais(s) {
		return false
	}
	for _, p := range pesos {
		soma := 0
		for j, peso := range p {
			soma += int(s[j]-'0') * peso
		}
		dv := 11 - soma%11
		if dv >= 10 {
			dv = 0
		}
		if int(s[len(p)]-'0') != dv {
			return false
		}
	}
	return true
}

func todosIguais(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}
