package airport

import (
	"api/internal/airport/handler"
	"api/internal/airport/infra/postgres"
	"api/internal/airport/usecase"

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
