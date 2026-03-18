package user

import (
	"airline-tracker/internal/user/controller"
	"airline-tracker/internal/user/infra/postgres"
	"airline-tracker/internal/user/service"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(controller.NewUserController),
	do.Lazy(service.NewUserService),
	do.Lazy(postgres.NewUserRepository),
)
