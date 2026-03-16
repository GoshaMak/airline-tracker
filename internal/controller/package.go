package controller

import "github.com/samber/do/v2"

var Package = do.Package(
	do.Lazy(NewUserController),
	do.Lazy(NewFlightController),
	do.Lazy(NewAdminController),
	do.Lazy(NewAuthController),
)
