package auth

import (
	"airline-tracker/internal/auth/controller"
	"airline-tracker/internal/auth/service"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(controller.NewAuthController),
	do.Lazy(service.NewAuthService),
)
