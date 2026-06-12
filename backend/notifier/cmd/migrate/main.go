package main

import (
	"log/slog"
	"notifier/internal/db"
	"notifier/internal/infra"
	"notifier/internal/infra/postgres"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

func main() {
	injector := do.New(
		infra.Package,
	)
	m, err := db.NewMigrator(injector)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer postgres.CloseConnection(do.MustInvoke[*pgxpool.Pool](injector))
	defer func() {
		if err := m.Down(); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}()
	if err := m.Up("migrations"); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info("Migrator successfully finished!")
}
