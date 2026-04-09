package redis

import (
	"log/slog"
	"os"

	rds "github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
)

// TODO: make custom type?
func NewRedisConnection(i do.Injector) (*rds.Client, error) {
	rdb := rds.NewClient(&rds.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})
	if rdb == nil {
		slog.Error("Unable to connect to redis")
		return nil, nil
	}
	slog.Info("Connected to redis")
	return rdb, nil
}

func CloseConnection(rdb *rds.Client) error {
	return rdb.Close()
}
