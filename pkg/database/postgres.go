package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Pool wraps pgxpool.Pool for database operations.
type Pool struct {
	*pgxpool.Pool
}

// NewPool creates a new PostgreSQL connection pool.
func NewPool(dsn string, maxOpen, maxIdle int, maxLifetime time.Duration) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	cfg.MaxConns = int32(maxOpen)
	cfg.MinConns = int32(maxIdle)
	cfg.MaxConnLifetime = maxLifetime

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Info().Str("dsn", maskDSN(dsn)).Msg("database connected")

	return &Pool{Pool: pool}, nil
}

// HealthCheck returns nil if the database is reachable.
func (p *Pool) HealthCheck(ctx context.Context) error {
	return p.Ping(ctx)
}

// DSN builds a PostgreSQL DSN from components.
func DSN(host string, port int, user, password, dbname string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user, password, host, port, dbname)
}

func maskDSN(dsn string) string {
	// simple masking: hide password
	raw := []byte(dsn)
	for i := 0; i < len(raw); i++ {
		if raw[i] == ':' {
			for j := i + 1; j < len(raw); j++ {
				if raw[j] == '@' {
					copy(raw[i+1:j], "****")
					break
				}
			}
			break
		}
	}
	return string(raw)
}
