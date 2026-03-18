package admin

import (
	"airline-tracker/internal/admin/controller"
	"airline-tracker/internal/admin/service"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(controller.NewAdminController),
	do.Lazy(service.NewAdminService),
)
