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
	"errors"
	"fmt"
	"time"

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
	router         *gin.Engine
	injector       do.Injector
	requestTimeout time.Duration
}

func NewServer(injector *do.RootScope) (*Server, error) {
	requestTimeout := 10 * time.Second
	if value := os.Getenv("API_REQUEST_TIMEOUT"); value != "" {
		var err error
		requestTimeout, err = time.ParseDuration(value)
		if err != nil || requestTimeout <= 0 {
			return nil, fmt.Errorf("API_REQUEST_TIMEOUT must be a positive duration, got %q", value)
		}
	}
	router := gin.New()
	router.Use(
		middleware.Observe(),
		gin.Recovery(),
		middleware.Instance(os.Getenv("INSTANCE_ID")),
	)

	registerRoutes(injector, router)

	return &Server{
		router:         router,
		injector:       injector,
		requestTimeout: requestTimeout,
	}, nil
}

func (s *Server) handler() http.Handler {
	return middleware.Correlate(http.TimeoutHandler(s.router, s.requestTimeout, `{"error":"request timeout"}`))
}

func (s *Server) Run(ctx context.Context) error {
	defer postgres.CloseConnection(do.MustInvoke[*pgxpool.Pool](s.injector))
	defer redis.CloseConnection(do.MustInvoke[*rds.Client](s.injector))

	port := os.Getenv("PORT")
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           s.handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      s.requestTimeout + 5*time.Second,
		IdleTimeout:       60 * time.Second,
	}
	shutdownDone := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		defer close(shutdownDone)
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			slog.Error("API shutdown failed", "err", err)
		}
	})
	defer func() {
		if !stop() {
			<-shutdownDone
		}
	}()
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
