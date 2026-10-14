package server

import (
	"api/cmd/docs"
	airportHandler "api/internal/airport/handler"
	authHandler "api/internal/auth/handler"
	fleetHandler "api/internal/fleet/handler"
	flightHandler "api/internal/flight/handler"
	"api/internal/infra/postgres"
	"api/internal/infra/redis"
	"api/internal/middleware"
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
)

type Server struct {
	router   *gin.Engine
	injector do.Injector
}

func NewServer(injector *do.RootScope) (*Server, error) {
	router := gin.Default()
	router.Use(middleware.Instance(os.Getenv("INSTANCE_ID")))

	registerRoutes(injector, router)

	return &Server{
		router:   router,
		injector: injector,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	defer postgres.CloseConnection(do.MustInvoke[*pgxpool.Pool](s.injector))
	defer redis.CloseConnection(do.MustInvoke[*rds.Client](s.injector))

	port := os.Getenv("PORT")
	if err := s.router.Run(":" + port); err != nil {
		slog.Error("Failed to run server", "err", err)
		return err
	}
	slog.Info("Program finished")
	return nil
}

func status(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "OK")
}

func createUUID(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, uuid.NewString())
}

func registerRoutes(i *do.RootScope, r *gin.Engine) {
	api := r.Group("/api/v1")

	{
		api.GET("/status", status)
		api.POST("/uuids", createUUID)
	}

	registerDocumentationRoutes(api)

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

	r.GET("/swagger/*any", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusMovedPermanently, "/api/v1/docs/")
	})
}

func registerDocumentationRoutes(api *gin.RouterGroup) {
	serveOpenAPI := func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "application/yaml; charset=utf-8", docs.OpenAPI)
	}

	api.GET("/openapi.yml", serveOpenAPI)
	api.GET("/openapi.yaml", serveOpenAPI)
	api.GET("/docs", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusMovedPermanently, "/api/v1/docs/")
	})

	assets := http.StripPrefix("/api/v1/docs", http.FileServer(swaggerFiles.HTTP))
	api.GET("/docs/*any", func(ctx *gin.Context) {
		switch ctx.Param("any") {
		case "/", "/index.html":
			ctx.Data(http.StatusOK, "text/html; charset=utf-8", docs.UI)
		case "/swagger-ui.css", "/swagger-ui-bundle.js", "/oauth2-redirect.html":
			assets.ServeHTTP(ctx.Writer, ctx.Request)
		default:
			ctx.Status(http.StatusNotFound)
		}
	})
}
