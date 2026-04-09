package airport

import (
	"airline-tracker/internal/airport/handler"
	"airline-tracker/internal/airport/infra/postgres"
	"airline-tracker/internal/airport/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(handler.NewAirportHandler),
	do.Lazy(usecase.NewAirportUsecase),

	do.Lazy(handler.NewGateHandler),
	do.Lazy(usecase.NewGateUsecase),

	do.Lazy(postgres.NewAirportRepository),
	do.Lazy(postgres.NewGateRepository),
)
