package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	rds "github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
)

func NewRedisClient(i do.Injector) (*rds.Client, error) {
	cln := rds.NewClient(&rds.Options{
		Addr:                  os.Getenv("REDIS_ADDR"),
		Password:              os.Getenv("REDIS_PASSWORD"),
		DB:                    0,
		DialTimeout:           3 * time.Second,
		ReadTimeout:           3 * time.Second,
		WriteTimeout:          3 * time.Second,
		PoolTimeout:           3 * time.Second,
		ContextTimeoutEnabled: true,
		MaxRetries:            2,
		MinRetryBackoff:       100 * time.Millisecond,
		MaxRetryBackoff:       300 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cln.Ping(ctx).Err(); err != nil {
		cln.Close()
		return nil, fmt.Errorf("unable to ping redis: %w", err)
	}

	return cln, nil
}

func CloseConnection(rdb *rds.Client) error {
	return rdb.Close()
}
