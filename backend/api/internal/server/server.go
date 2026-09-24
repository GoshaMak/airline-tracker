package server

import (
	"api/cmd/docs"
	airportHandler "api/internal/airport/handler"
	authHandler "api/internal/auth/handler"
	fleetHandler "api/internal/fleet/handler"
	flightHandler "api/internal/flight/handler"
	"api/internal/infra/kafka"
	"api/internal/infra/postgres"
	"api/internal/infra/redis"
	userHandler "api/internal/user/handler"
	"context"

	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	rds "github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Airline Tracker
// @version 1.0.0
// @description HTTP API for browsing flights and managing airline tracker data.

// @host localhost:8080

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

type Server struct {
	router   *gin.Engine
	injector do.Injector
}

func NewServer(injector *do.RootScope) (*Server, error) {
	router := gin.Default()

	registerRoutes(injector, router)

	return &Server{
		router:   router,
		injector: injector,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
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
// @Router /api/v1/status [get]
func status(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "OK")
}

// @Summary create UUID
// @Description creates a new UUID
// @Tags Utils
// @Accept json
// @Produce json
// @Success 200 "uuid"
// @Router /api/v1/uuids [post]
func createUUID(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, uuid.NewString())
}

func registerRoutes(i *do.RootScope, r *gin.Engine) {
	api := r.Group("/api/v1")
	api.GET("/status", status)
	api.POST("/uuids", createUUID)
	api.GET("/openapi.yaml", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "application/yaml; charset=utf-8", docs.OpenAPI)
	})

	authHandler.RegisterAuthRoutes(i, api)
	slog.Debug("Auth routes successfully registered")

	{
		airportHandler.RegisterAirportRoutes(i, api)
		airportHandler.RegisterGateRoutes(i, api)
		slog.Debug("Airport routes successfully registered")
	}

	{
		fleetHandler.RegisterAircraftRoutes(i, api)
		fleetHandler.RegisterAircraftModelRoutes(i, api)
		slog.Debug("Fleet routes successfully registered")
	}

	flightHandler.RegisterRoutes(i, api)
	slog.Debug("Flight routes successfully registered")

	userHandler.RegisterRoutes(i, api)
	slog.Debug("User routes successfully registered")

	r.GET(
		"/swagger/*any",
		ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.URL("/api/v1/openapi.yaml"),
		),
	)
	slog.Debug("Swagger routes successfully registered")
}
