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
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	connStr := os.Getenv("POSTGRES_URI")
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse postgres config: %w", err)
	}

	if config.ConnConfig.ConnectTimeout == 0 {
		config.ConnConfig.ConnectTimeout = 5 * time.Second
	}
	for name, value := range map[string]string{
		"statement_timeout":                   "10000",
		"lock_timeout":                        "3000",
		"idle_in_transaction_session_timeout": "10000",
	} {
		if _, configured := config.ConnConfig.RuntimeParams[name]; !configured {
			config.ConnConfig.RuntimeParams[name] = value
		}
	}
	for attempt := 1; attempt <= 3; attempt++ {
		var pool *pgxpool.Pool
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			ping, cancel := context.WithTimeout(ctx, 3*time.Second)
			err = pool.Ping(ping)
			cancel()
		}
		if err == nil {
			slog.Info("Connected to PostgreSQL")
			return pool, nil
		}
		if pool != nil {
			pool.Close()
		}
		if attempt == 3 {
			break
		}
		delay := time.Duration(attempt) * 500 * time.Millisecond
		slog.Warn("Postgres not ready", "attempt", attempt, "delay", delay, "err", err)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("unable to connect to postgres: %w", ctx.Err())
		case <-timer.C:
		}
	}

	return nil, fmt.Errorf("unable to ping postgres: %w", err)
}

func CloseConnection(c *pgxpool.Pool) error {
	c.Close()
	return nil
}
