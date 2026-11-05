// Package config lê variáveis de ambiente e acumula as obrigatórias que
// faltam, para o serviço falhar na partida com uma única mensagem clara.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ErrFaltando indica variáveis obrigatórias ausentes ou vazias.
var ErrFaltando = errors.New("variáveis de ambiente obrigatórias ausentes")

// Busca devolve o valor de uma variável e se ela existe (como os.LookupEnv).
type Busca func(nome string) (string, bool)

// Leitor lê variáveis de uma Busca e registra as obrigatórias ausentes.
type Leitor struct {
	busca    Busca
	faltando []string
}

// NovoLeitor cria um Leitor. Com busca nil, usa os.LookupEnv.
func NovoLeitor(busca Busca) *Leitor {
	if busca == nil {
		busca = os.LookupEnv
	}
	return &Leitor{busca: busca}
}

func (l *Leitor) valor(nome string) string {
	v, _ := l.busca(nome)
	return strings.TrimSpace(v)
}

// Obrigatorio devolve a variável ou registra a falta dela.
func (l *Leitor) Obrigatorio(nome string) string {
	v := l.valor(nome)
	if v == "" {
		l.faltando = append(l.faltando, nome)
	}
	return v
}

// Padrao devolve a variável ou o valor padrão se ela estiver ausente.
func (l *Leitor) Padrao(nome, padrao string) string {
	if v := l.valor(nome); v != "" {
		return v
	}
	return padrao
}

// Lista lê uma variável obrigatória com itens separados por vírgula.
func (l *Leitor) Lista(nome string) []string {
	var itens []string
	for _, p := range strings.Split(l.Obrigatorio(nome), ",") {
		if p = strings.TrimSpace(p); p != "" {
			itens = append(itens, p)
		}
	}
	return itens
}

// Erro devolve ErrFaltando com os nomes ausentes, ou nil.
func (l *Leitor) Erro() error {
	if len(l.faltando) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrFaltando, strings.Join(l.faltando, ", "))
}
