package publisher

import (
	"api/internal/publisher/infra/postgres"
	"api/internal/publisher/usecase"

	"github.com/samber/do/v2"
)

var OutboxPackage = do.Package(
	do.Lazy(postgres.NewOutboxRepository),
)

var Package = do.Package(
	do.Lazy(postgres.NewOutboxRepository),
	do.Lazy(usecase.NewPublisherUsecase),
)
