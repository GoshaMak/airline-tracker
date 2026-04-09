package fleet

import (
	"airline-tracker/internal/fleet/controller"
	"airline-tracker/internal/fleet/infra/postgres"
	"airline-tracker/internal/fleet/service"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(controller.NewAircraftController),
	do.Lazy(service.NewAircraftService),
	do.Lazy(postgres.NewAircraftRepository),

	do.Lazy(controller.NewAircraftModelController),
	do.Lazy(service.NewAircraftModelService),
	do.Lazy(postgres.NewAircraftModelRepository),
)
