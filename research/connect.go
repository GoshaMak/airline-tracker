package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool() (*pgxpool.Pool, error) {
	return newPool(context.Background())
}

func newPool(ctx context.Context) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("POSTGRE_USER"),
		os.Getenv("POSTGRE_PASSWORD"),
		os.Getenv("POSTGRE_HOST"),
		os.Getenv("POSTGRE_PORT"),
		os.Getenv("POSTGRE_NAME"),
	)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping postgres")
	}

	return pool, nil
}

func CloseConnection(c *pgxpool.Pool) error {
	c.Close()
	return nil
}
