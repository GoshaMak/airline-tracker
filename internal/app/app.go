package app

import (
	"airline-tracker/cmd/docs"
	"airline-tracker/internal/airport"
	airportController "airline-tracker/internal/airport/handler"
	"airline-tracker/internal/auth"
	authController "airline-tracker/internal/auth/handler"
	"airline-tracker/internal/fleet"
	fleetController "airline-tracker/internal/fleet/handler"
	"airline-tracker/internal/flight"
	flightController "airline-tracker/internal/flight/handler"
	"airline-tracker/internal/infra"
	"airline-tracker/internal/infra/postgres"
	"airline-tracker/internal/infra/redis"
	"airline-tracker/internal/user"
	userController "airline-tracker/internal/user/handler"

	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	rds "github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title my api

// @host localhost:8080

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

type App struct {
	router   *gin.Engine
	injector *do.RootScope
}

func NewApp() *App {
	// TODO: add SIGINT handler
	godotenv.Load()

	setupLogger()

	injector := do.New(
		auth.Package,
		airport.Package,
		fleet.Package,
		flight.Package,
		user.Package,
		infra.Package,
	)

	router := gin.Default()

	registerRoutes(injector, router)

	return &App{
		router:   router,
		injector: injector,
	}
}

func (a *App) Run() {
	defer postgres.CloseConnection(do.MustInvoke[*pgxpool.Pool](a.injector))
	defer redis.CloseConnection(do.MustInvoke[*rds.Client](a.injector))

	port := os.Getenv("PORT")
	if err := a.router.Run(":" + port); err != nil {
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

// @Summary status example
// @Description check status
// @Tags Health check
// @Accept json
// @Produce json
// @Success 200 "OK"
// @Router /status [get]
func status(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "OK")
}

func registerRoutes(i *do.RootScope, r *gin.Engine) {
	r.GET("/status", status)

	authController.RegisterAuthRoutes(i, r)
	slog.Debug("Auth routes successfully registered")

	{
		airportController.RegisterAirportRoutes(i, r)
		airportController.RegisterGateRoutes(i, r)
		slog.Debug("Airport routes successfully registered")
	}

	{
		fleetController.RegisterAircraftRoutes(i, r)
		fleetController.RegisterAircraftModelRoutes(i, r)
		slog.Debug("Fleet routes successfully registered")
	}

	flightController.RegisterRoutes(i, r)
	slog.Debug("Flight routes successfully registered")

	userController.RegisterRoutes(i, r)
	slog.Debug("User routes successfully registered")

	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	slog.Debug("Swagger routes successfully registered")
}
