package sender

import (
	"notifier/internal/sender/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(NewSender),
	do.Lazy(usecase.NewSenderUsecase),
	do.Lazy(usecase.NewEmailSenderUsecase),
)
