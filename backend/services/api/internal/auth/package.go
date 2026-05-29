package auth

import (
	"api/internal/auth/handler"
	"api/internal/auth/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(handler.NewAuthHandler),
	do.Lazy(usecase.NewAuthUsecase),
)
