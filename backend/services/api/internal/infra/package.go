package infra

import (
	"api/internal/infra/kafka"
	"api/internal/infra/mongo"
	"api/internal/infra/redis"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(mongo.NewMongoDatabase),
	do.Lazy(mongo.NewMongoClient),
	// do.Lazy(postgres.NewPostgresPool),
	do.Lazy(redis.NewRedisClient),
	do.Lazy(kafka.NewNotifySender),
)
