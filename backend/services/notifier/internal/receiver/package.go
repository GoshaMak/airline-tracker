package receiver

import (
	"notifier/internal/receiver/infra/postgres"
	"notifier/internal/receiver/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(postgres.NewNotificationRepository),
	do.Lazy(usecase.NewNotifierUsecase),
)
