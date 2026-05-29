package server

import (
	"api/cmd/docs"
	airportHandler "api/internal/airport/handler"
	authHandler "api/internal/auth/handler"
	fleetHandler "api/internal/fleet/handler"
	flightHandler "api/internal/flight/handler"
	"api/internal/infra/kafka"
	"api/internal/infra/mysql"
	"api/internal/infra/redis"
	userHandler "api/internal/user/handler"
	"context"
	"database/sql"

	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	defer mysql.CloseConnection(do.MustInvoke[*sql.DB](s.injector))
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
