package fleet

import (
	"api/internal/fleet/handler"
	"api/internal/fleet/infra/mysql"
	"api/internal/fleet/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(handler.NewAircraftHandler),
	do.Lazy(usecase.NewAircraftUsecase),
	do.Lazy(mysql.NewAircraftRepository),

	do.Lazy(handler.NewAircraftModelHandler),
	do.Lazy(usecase.NewAircraftModelUsecase),
	do.Lazy(mysql.NewAircraftModelRepository),
)
