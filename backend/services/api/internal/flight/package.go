package flight

import (
	"api/internal/flight/handler"
	"api/internal/flight/infra/postgres"
	"api/internal/flight/infra/redis"
	"api/internal/flight/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(handler.NewFlightHandler),
	do.Lazy(usecase.NewFlightUsecase),
	do.Lazy(postgres.NewFlightRepository),
	do.Lazy(redis.NewFlightCache),
)
