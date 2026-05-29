package flight

import (
	"api/internal/flight/handler"
	"api/internal/flight/infra/mongo"
	"api/internal/flight/infra/redis"
	"api/internal/flight/usecase"
	"api/internal/flight/usecase/repository"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(handler.NewFlightHandler),
	do.Lazy(usecase.NewFlightUsecase),

	do.Lazy(repository.NewFlightRepository),
	do.Lazy(mongo.NewMongoDB),
	// do.Lazy(postgres.NewPostgresDB),
	do.Lazy(redis.NewRedisDB),
)
