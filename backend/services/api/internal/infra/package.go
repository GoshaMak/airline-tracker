package infra

import (
	"api/internal/infra/kafka"
	"api/internal/infra/mysql"
	"api/internal/infra/redis"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(mysql.NewMySQLDB),
	do.Lazy(redis.NewRedisClient),
	do.Lazy(kafka.NewNotifySender),
)
