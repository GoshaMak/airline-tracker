package fleet

import (
	"airline-tracker/internal/fleet/handler"
	"airline-tracker/internal/fleet/infra/postgres"
	"airline-tracker/internal/fleet/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(handler.NewAircraftHandler),
	do.Lazy(usecase.NewAircraftUsecase),
	do.Lazy(postgres.NewAircraftRepository),

	do.Lazy(handler.NewAircraftModelHandler),
	do.Lazy(usecase.NewAircraftModelUsecase),
	do.Lazy(postgres.NewAircraftModelRepository),
)
