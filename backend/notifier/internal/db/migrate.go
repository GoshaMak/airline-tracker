package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/samber/do/v2"
)

//go:embed migrations/*.sql
var fs embed.FS

type Migrator struct {
	db *sql.DB
}

func NewMigrator(i do.Injector) (*Migrator, error) {
	const op = "NewMigrator"
	pool := do.MustInvoke[*pgxpool.Pool](i)
	db := stdlib.OpenDBFromPool(pool)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Migrator{
		db: db,
	}, nil
}

func (m *Migrator) Up(dir string) error {
	const op = "Migrator.Up"
	goose.SetBaseFS(fs)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := goose.Up(m.db, dir); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (m *Migrator) Down() error {
	const op = "Migrator.Down"
	if err := m.db.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
