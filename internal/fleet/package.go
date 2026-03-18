package fleet

import (
	"airline-tracker/internal/fleet/infra/postgres"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(postgres.NewAircraftRepository),
	do.Lazy(postgres.NewAircraftModelRepository),
)
