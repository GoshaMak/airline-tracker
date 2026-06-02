package infra

import (
	"notifier/internal/infra/mysql"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(mysql.NewMySQLDB),
)
