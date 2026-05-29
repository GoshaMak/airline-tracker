package infra

import (
	"notifier/internal/infra/mongo"
	"notifier/internal/infra/postgres"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(mongo.NewMongoClient),
	do.Lazy(mongo.NewMongoDatabase),
	do.Lazy(postgres.NewPostgresPool),
)
