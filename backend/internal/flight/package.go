package flight

import (
	"airline-tracker/internal/flight/handler"
	"airline-tracker/internal/flight/infra/postgres"
	"airline-tracker/internal/flight/infra/redis"
	"airline-tracker/internal/flight/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(handler.NewFlightHandler),
	do.Lazy(usecase.NewFlightUsecase),
	do.Lazy(postgres.NewFlightRepository),
	do.Lazy(redis.NewFlightCache),
)
