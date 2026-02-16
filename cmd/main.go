package main

import (
	"airline-ticketing-svc/internal/controller"
	"airline-ticketing-svc/internal/db"
	"airline-ticketing-svc/internal/postgres"
	routes "airline-ticketing-svc/internal/route"
	"airline-ticketing-svc/internal/service"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do/v2"
)

func main() {
	injector := do.New(
		postgres.Package,
		controller.Package,
		service.Package,
		db.Package,
	)

	defer db.CloseConnection(do.MustInvoke[*pgx.Conn](injector))

	setupLogger()

	r := gin.Default()
	routes.RegisterUserRoutes(injector, r)

	if err := r.Run(); err != nil {
		slog.Error("Failed to run server", "error", err)
		os.Exit(1)
	}
	slog.Info("Program finished")
}

func setupLogger() {
	logger := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	))
	if logger == nil {
		os.Exit(1)
	}
	slog.SetDefault(logger)
	slog.Info("Logger inited")
}
