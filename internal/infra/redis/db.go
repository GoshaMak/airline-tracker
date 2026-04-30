package redis

import (
	"context"
	"errors"
	"fmt"
	"os"

	rds "github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
)

func NewRedisClient(i do.Injector) (*rds.Client, error) {
	cln := rds.NewClient(&rds.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	if cln == nil {
		return nil, errors.New("unable to connect to redis")
	}

	if err := cln.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("unable to ping redis")
	}

	return cln, nil
}

func CloseConnection(rdb *rds.Client) error {
	return rdb.Close()
}
