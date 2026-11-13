package auth

import (
	"api/internal/auth/handler"
	authRedis "api/internal/auth/infra/redis"
	"api/internal/auth/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(authRedis.NewAuthStateStore),
	do.Lazy(handler.NewAuthHandler),
	do.Lazy(usecase.NewAuthUsecase),
)
