package publisher

import (
	"api/internal/publisher/infra/mysql"
	"api/internal/publisher/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(mysql.NewOutboxRepository),
	do.Lazy(usecase.NewPublisherUsecase),
)
