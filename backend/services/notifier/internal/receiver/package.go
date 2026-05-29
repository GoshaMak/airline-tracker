package receiver

import (
	"notifier/internal/receiver/infra/mongo"
	"notifier/internal/receiver/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(mongo.NewNotificationRepository),
	// do.Lazy(postgres.NewNotificationRepository),
	do.Lazy(usecase.NewNotifierUsecase),
)
