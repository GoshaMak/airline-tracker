package service

import "github.com/samber/do/v2"

var Package = do.Package(
	do.Lazy(NewUserService),
	do.Lazy(NewFlightService),
	do.Lazy(NewAdminService),
	do.Lazy(NewAuthService),
)
