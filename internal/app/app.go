package app

import (
	"airline-tracker/internal/airport"
	airportController "airline-tracker/internal/airport/controller"
	"airline-tracker/internal/auth"
	authController "airline-tracker/internal/auth/controller"
	"airline-tracker/internal/db"
	"airline-tracker/internal/fleet"
	aircraftController "airline-tracker/internal/fleet/controller"
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
		auth.Package,
		airport.Package,
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
	slog.Debug("Auth routes successfully registered")

	{
		airportController.RegisterAirportRoutes(i, r)
		airportController.RegisterGateRoutes(i, r)
		slog.Debug("Airport routes successfully registered")
	}

	{
		aircraftController.RegisterAircraftRoutes(i, r)
		aircraftController.RegisterAircraftModelRoutes(i, r)
		slog.Debug("Fleet routes successfully registered")
	}

	flightController.RegisterRoutes(i, r)
	slog.Debug("Flight routes successfully registered")

	userController.RegisterRoutes(i, r)
	slog.Debug("User routes successfully registered")
}
