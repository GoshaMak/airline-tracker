package fleet

import (
	"api/internal/fleet/handler"
	"api/internal/fleet/infra/mongo"
	"api/internal/fleet/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(handler.NewAircraftHandler),
	do.Lazy(usecase.NewAircraftUsecase),
	do.Lazy(mongo.NewAircraftRepository),
	// do.Lazy(postgres.NewAircraftRepository),

	do.Lazy(handler.NewAircraftModelHandler),
	do.Lazy(usecase.NewAircraftModelUsecase),
	do.Lazy(mongo.NewAircraftModelRepository),
	// do.Lazy(postgres.NewAircraftModelRepository),
)
