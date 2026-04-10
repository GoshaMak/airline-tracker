package user

import (
	"airline-tracker/internal/user/handler"
	"airline-tracker/internal/user/infra/postgres"
	"airline-tracker/internal/user/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(handler.NewUserHandler),
	do.Lazy(usecase.NewUserUsecase),
	do.Lazy(postgres.NewUserRepository),
)
