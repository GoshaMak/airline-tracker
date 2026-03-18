package app

import (
	"airline-tracker/internal/admin"
	adminController "airline-tracker/internal/admin/controller"
	"airline-tracker/internal/airport"
	"airline-tracker/internal/auth"
	authController "airline-tracker/internal/auth/controller"
	"airline-tracker/internal/db"
	"airline-tracker/internal/fleet"
	"airline-tracker/internal/flight"
	flightController "airline-tracker/internal/flight/controller"
	"airline-tracker/internal/user"
	userController "airline-tracker/internal/user/controller"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/samber/do/v2"
)

type App struct {
	router   *gin.Engine
	injector *do.RootScope
}

func NewApp() *App {
	godotenv.Load()

	setupLogger()

	injector := do.New(
		admin.Package,
		airport.Package,
		auth.Package,
		fleet.Package,
		flight.Package,
		user.Package,
		db.Package,
	)

	router := gin.Default()

	registerRoutes(injector, router)

	return &App{
		router:   router,
		injector: injector,
	}
}

func (a *App) Run() {
	defer db.CloseConnection(do.MustInvoke[*pgx.Conn](a.injector))

	if err := a.router.Run(); err != nil {
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
	r.GET("/status", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"msg": "everything is workind"})
	})
	authController.RegisterAuthRoutes(i, r)
	adminController.RegisterAdminRoutes(i, r)
	userController.RegisterUserRoutes(i, r)
	flightController.RegisterFlightRoutes(i, r)
}
