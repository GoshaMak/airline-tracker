package infra

import (
	"airline-tracker/internal/infra/postgres"
	"airline-tracker/internal/infra/redis"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(postgres.NewPostgresPool),
	do.Lazy(redis.NewRedisConnection),
)
