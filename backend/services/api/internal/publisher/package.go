package publisher

import (
	"api/internal/publisher/infra/mongo"
	"api/internal/publisher/usecase"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(mongo.NewOutboxRepository),
	// do.Lazy(postgres.NewOutboxRepository),
	do.Lazy(usecase.NewPublisherUsecase),
)
