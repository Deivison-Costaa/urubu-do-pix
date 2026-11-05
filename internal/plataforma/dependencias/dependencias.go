// Package dependencias conecta os serviços a Postgres, Redis e Redpanda e
// expõe uma verificação de prontidão para cada conexão.
package dependencias

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/Deivison-Costaa/urubu-do-pix/internal/plataforma/saude"
)

// Postgres cria o pool de conexões. A conexão é preguiçosa: o serviço sobe
// mesmo com o banco fora do ar e fica "indisponível" no /readyz até ele voltar.
func Postgres(ctx context.Context, url string) (*pgxpool.Pool, saude.Verificacao, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, saude.Verificacao{}, fmt.Errorf("DATABASE_URL inválida: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, saude.Verificacao{}, fmt.Errorf("criar pool do Postgres: %w", err)
	}
	return pool, saude.Verificacao{Nome: "postgres", Verificar: pool.Ping}, nil
}

// Redis cria o cliente a partir de uma URL redis://.
func Redis(url string) (*redis.Client, saude.Verificacao, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, saude.Verificacao{}, fmt.Errorf("REDIS_URL inválida: %w", err)
	}
	cli := redis.NewClient(opt)
	return cli, saude.Verificacao{Nome: "redis", Verificar: func(ctx context.Context) error {
		return cli.Ping(ctx).Err()
	}}, nil
}

// Kafka cria o cliente da API Kafka (Redpanda).
func Kafka(brokers []string, clientID string) (*kgo.Client, saude.Verificacao, error) {
	if len(brokers) == 0 {
		return nil, saude.Verificacao{}, fmt.Errorf("KAFKA_BROKERS vazio")
	}
	cli, err := kgo.NewClient(kgo.SeedBrokers(brokers...), kgo.ClientID(clientID))
	if err != nil {
		return nil, saude.Verificacao{}, fmt.Errorf("criar cliente Kafka: %w", err)
	}
	return cli, saude.Verificacao{Nome: "kafka", Verificar: cli.Ping}, nil
}
