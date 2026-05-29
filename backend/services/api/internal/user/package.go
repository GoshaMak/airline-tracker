package user

import (
	"api/internal/user/handler"
	"api/internal/user/infra/mongo"
	"api/internal/user/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(handler.NewUserHandler),
	do.Lazy(usecase.NewUserUsecase),
	do.Lazy(mongo.NewUserRepository),
	// do.Lazy(postgres.NewUserRepository),
)
