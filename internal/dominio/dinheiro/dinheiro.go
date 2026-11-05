// Package dinheiro representa valores monetários em centavos inteiros.
//
// Nunca usamos ponto flutuante para dinheiro: a soma de float64 perde
// centavos, e o invariante central do projeto é que nenhum centavo seja
// criado ou perdido.
package dinheiro

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Centavos é um valor monetário em centavos de real.
type Centavos int64

var (
	// ErrFormato indica texto fora do formato da API Pix ("123.45").
	ErrFormato = errors.New("valor fora do formato 0.00")
	// ErrValorNaoPositivo indica valor zero, que não é aceito em pagamento.
	ErrValorNaoPositivo = errors.New("valor deve ser maior que zero")
	// ErrOverflow indica soma fora do intervalo de int64.
	ErrOverflow = errors.New("soma de valores excede o limite")
)

// Mesmo padrão do campo "valor" da API Pix do Banco Central:
// até 10 dígitos inteiros e exatamente 2 casas decimais.
var formatoValor = regexp.MustCompile(`^\d{1,10}\.\d{2}$`)

// Parse converte o texto de um valor de pagamento ("123.45") em centavos.
func Parse(s string) (Centavos, error) {
	if !formatoValor.MatchString(s) {
		return 0, fmt.Errorf("%w: %q", ErrFormato, s)
	}
	inteiro, fracao, _ := strings.Cut(s, ".")
	n, err := strconv.ParseInt(inteiro+fracao, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %q", ErrFormato, s)
	}
	if n == 0 {
		return 0, ErrValorNaoPositivo
	}
	return Centavos(n), nil
}

// String formata o valor com duas casas decimais, como na API Pix.
func (c Centavos) String() string {
	sinal := ""
	v := int64(c)
	if v < 0 {
		sinal = "-"
		v = -v
	}
	return fmt.Sprintf("%s%d.%02d", sinal, v/100, v%100)
}

// Somar soma os valores e falha em vez de dar a volta no int64.
func Somar(valores ...Centavos) (Centavos, error) {
	var total int64
	for _, v := range valores {
		x := int64(v)
		if (x > 0 && total > math.MaxInt64-x) || (x < 0 && total < math.MinInt64-x) {
			return 0, ErrOverflow
		}
		total += x
	}
	return Centavos(total), nil
}
