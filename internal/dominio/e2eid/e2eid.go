// Package e2eid gera e valida o EndToEndId, identificador único de um
// pagamento Pix e chave de idempotência em todas as etapas da SAGA.
//
// Formato (32 caracteres): "E" + ISPB do pagador (8 dígitos) +
// data e hora em UTC (AAAAMMDDHHmm) + 11 caracteres alfanuméricos.
package e2eid

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"
)

// ID é um EndToEndId válido.
type ID string

// Tamanho é o comprimento fixo de um EndToEndId.
const Tamanho = 32

const (
	alfabeto      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	tamanhoSufixo = 11
	layoutHora    = "200601021504"
	// Maior múltiplo de len(alfabeto) que cabe num byte: bytes acima dele
	// são descartados para não enviesar a distribuição.
	limiteSemVies = 256 - 256%len(alfabeto)
)

var (
	// ErrISPB indica ISPB fora do formato de 8 dígitos.
	ErrISPB = errors.New("ISPB deve ter 8 dígitos")
	// ErrFormato indica EndToEndId malformado.
	ErrFormato = errors.New("EndToEndId malformado")

	formatoISPB = regexp.MustCompile(`^\d{8}$`)
	formatoID   = regexp.MustCompile(`^E(\d{8})(\d{12})[a-zA-Z0-9]{11}$`)
)

// Novo gera um EndToEndId para o PSP pagador identificado por ispb.
// Se fonte for nil, usa crypto/rand.
func Novo(ispb string, agora time.Time, fonte io.Reader) (ID, error) {
	if !formatoISPB.MatchString(ispb) {
		return "", fmt.Errorf("%w: %q", ErrISPB, ispb)
	}
	if fonte == nil {
		fonte = rand.Reader
	}
	sufixo, err := sufixoAleatorio(fonte)
	if err != nil {
		return "", fmt.Errorf("gerar sufixo do EndToEndId: %w", err)
	}
	return ID("E" + ispb + agora.UTC().Format(layoutHora) + sufixo), nil
}

func sufixoAleatorio(fonte io.Reader) (string, error) {
	saida := make([]byte, 0, tamanhoSufixo)
	buf := make([]byte, tamanhoSufixo*2)
	for len(saida) < tamanhoSufixo {
		if _, err := io.ReadFull(fonte, buf); err != nil {
			return "", err
		}
		for _, b := range buf {
			if int(b) >= limiteSemVies {
				continue
			}
			saida = append(saida, alfabeto[int(b)%len(alfabeto)])
			if len(saida) == tamanhoSufixo {
				break
			}
		}
	}
	return string(saida), nil
}

// Validar verifica formato e data de um EndToEndId.
func Validar(s string) error {
	m := formatoID.FindStringSubmatch(s)
	if m == nil {
		return fmt.Errorf("%w: %q", ErrFormato, s)
	}
	if _, err := time.Parse(layoutHora, m[2]); err != nil {
		return fmt.Errorf("%w: data inválida em %q", ErrFormato, s)
	}
	return nil
}

// ISPB extrai o ISPB do PSP pagador de um EndToEndId.
func ISPB(s string) (string, error) {
	if err := Validar(s); err != nil {
		return "", err
	}
	return s[1:9], nil
}
