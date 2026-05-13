package server

import (
	"api/cmd/docs"
	"api/internal/airport"
	airportHandler "api/internal/airport/handler"
	"api/internal/auth"
	authHandler "api/internal/auth/handler"
	"api/internal/fleet"
	fleetHandler "api/internal/fleet/handler"
	"api/internal/flight"
	flightHandler "api/internal/flight/handler"
	"api/internal/infra"
	"api/internal/infra/kafka"
	"api/internal/infra/postgres"
	"api/internal/infra/redis"
	"api/internal/notification"
	"api/internal/user"
	userHandler "api/internal/user/handler"
	"io"
	"shared/logger"

	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

type Server struct {
	router   *gin.Engine
	injector *do.RootScope
}

func NewServer() (*Server, error) {
	// TODO: add SIGINT handler
	godotenv.Load()

	var w io.Writer
	switch os.Getenv("MODE") {
	case "DEBUG":
		w = os.Stdout
	default:
		logFileName := os.Getenv("LOG_FILE")
		logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return nil, err
		}
		defer logFile.Close()
		w = logFile
	}

	if err := logger.SetupLogger(w); err != nil {
		return nil, err
	}
	slog.Info("Logger initialized")

	injector := do.New(
		auth.Package,
		airport.Package,
		fleet.Package,
		flight.Package,
		user.Package,
		infra.Package,
		notification.Package,
	)

	router := gin.Default()

	registerRoutes(injector, router)

	return &Server{
		router:   router,
		injector: injector,
	}, nil
}

func (s *Server) Run() error {
	defer postgres.CloseConnection(do.MustInvoke[*pgxpool.Pool](s.injector))
	defer redis.CloseConnection(do.MustInvoke[*rds.Client](s.injector))
	defer kafka.CloseConnection(do.MustInvoke[*kafka.NotifySender](s.injector))

	port := os.Getenv("PORT")
	if err := s.router.Run(":" + port); err != nil {
		slog.Error("Failed to run server", "err", err)
		return err
	}
	slog.Info("Program finished")
	return nil
}

// @Summary status example
// @Description check status
// @Tags Utils
// @Accept json
// @Produce json
// @Success 200 "OK"
// @Router /status [get]
func status(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "OK")
}

// @Summary create uuid
// @Description creates new uuid
// @Tags Utils
// @Accept json
// @Produce json
// @Success 200 "uuid"
// @Router /create_uuid [get]
func createUUID(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, uuid.NewString())
}

func registerRoutes(i *do.RootScope, r *gin.Engine) {
	r.GET("/status", status)
	r.GET("/create_uuid", createUUID)

	authHandler.RegisterAuthRoutes(i, r)
	slog.Debug("Auth routes successfully registered")

	{
		airportHandler.RegisterAirportRoutes(i, r)
		airportHandler.RegisterGateRoutes(i, r)
		slog.Debug("Airport routes successfully registered")
	}

	{
		fleetHandler.RegisterAircraftRoutes(i, r)
		fleetHandler.RegisterAircraftModelRoutes(i, r)
		slog.Debug("Fleet routes successfully registered")
	}

	flightHandler.RegisterRoutes(i, r)
	slog.Debug("Flight routes successfully registered")

	userHandler.RegisterRoutes(i, r)
	slog.Debug("User routes successfully registered")

	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	slog.Debug("Swagger routes successfully registered")
}
