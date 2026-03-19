package airport

import (
	"airline-tracker/internal/airport/controller"
	"airline-tracker/internal/airport/infra/postgres"
	"airline-tracker/internal/airport/service"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(controller.NewAirportController),
	do.Lazy(service.NewAirportService),

	do.Lazy(controller.NewGateController),
	do.Lazy(service.NewGateService),

	do.Lazy(postgres.NewAirportRepository),
	do.Lazy(postgres.NewGateRepository),
)
