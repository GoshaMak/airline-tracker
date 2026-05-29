package receiver

import (
	"notifier/internal/receiver/infra/mysql"
	"notifier/internal/receiver/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(mysql.NewNotificationRepository),
	do.Lazy(usecase.NewNotifierUsecase),
)
