package main

import (
	"airline-tracker/internal/controller"
	"airline-tracker/internal/db"
	"airline-tracker/internal/postgres"
	"airline-tracker/internal/service"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/samber/do/v2"
)

func main() {
	godotenv.Load()

	injector := do.New(
		controller.Package,
		service.Package,
		db.Package,
		postgres.Package,
	)

	defer db.CloseConnection(do.MustInvoke[*pgx.Conn](injector))

	setupLogger()

	r := gin.Default()
	registerRoutes(injector, r)

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

func registerRoutes(i *do.RootScope, r *gin.Engine) {
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"msg": "pong"})
	})
	controller.RegisterAuthRoutes(i, r)
	controller.RegisterAdminRoutes(i, r)
	controller.RegisterUserRoutes(i, r)
	controller.RegisterFlightRoutes(i, r)
}
