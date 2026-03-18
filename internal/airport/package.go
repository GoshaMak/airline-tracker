package airport

import (
	postgres "airline-tracker/internal/airport/infra/postgres"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(postgres.NewAirportRepository),
	do.Lazy(postgres.NewGateRepository),
)
