package db

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/samber/do/v2"
)

func NewConnection(i do.Injector) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("Unable to connect to databse", "error", err)
		return nil, nil
	}
	slog.Info("Connected to db")
	return conn, err
}

func CloseConnection(c *pgx.Conn) error {
	return c.Close(context.Background())
}
