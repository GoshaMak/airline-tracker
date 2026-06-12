package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

func NewPostgresPool(i do.Injector) (*pgxpool.Pool, error) {
	return newPool(context.Background())
}

func newPool(ctx context.Context) (*pgxpool.Pool, error) {
	connStr := os.Getenv("POSTGRES_URI")
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse postgres config: %w", err)
	}

	attempts := 0
	delay := 1 * time.Second
	for {
		pool, err := pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			err = pool.Ping(ctx)
		}
		if err == nil {
			slog.Info("Connected to PostgreSQL")
			return pool, nil
		}
		attempts++
		if attempts > 5 {
			break
		}
		slog.Warn("Postgres not ready", "attempt", attempts, "delay %v", delay, "err", err)
		time.Sleep(delay)
		delay *= 2
	}

	return nil, fmt.Errorf("unable to ping postgres")
}

func CloseConnection(c *pgxpool.Pool) error {
	c.Close()
	return nil
}
