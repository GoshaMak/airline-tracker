package flight

import (
	"airline-tracker/internal/flight/controller"
	"airline-tracker/internal/flight/infra/postgres"
	"airline-tracker/internal/flight/service"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(controller.NewFlightController),
	do.Lazy(service.NewFlightService),
	do.Lazy(postgres.NewFlightRepository),
)
