package auth

import (
	"airline-tracker/internal/auth/handler"
	"airline-tracker/internal/auth/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(handler.NewAuthHandler),
	do.Lazy(usecase.NewAuthUsecase),
)
