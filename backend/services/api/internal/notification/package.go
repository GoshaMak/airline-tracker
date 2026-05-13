package notification

import (
	"api/internal/notification/usecase"

	"github.com/samber/do/v2"
)

// server side of notifier
var Package = do.Package(
	do.Lazy(usecase.NewNotificationUsecase),
)
