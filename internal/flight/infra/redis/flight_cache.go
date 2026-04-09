package redis

import (
	"airline-tracker/internal/flight/domain"
	"airline-tracker/internal/flight/domain/repository"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	rds "github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
)

type flightCache struct {
	cl *rds.Client
}

func NewFlightCache(i do.Injector) (repository.FlightCache, error) {
	return &flightCache{
		cl: do.MustInvoke[*rds.Client](i),
	}, nil
}

func (c *flightCache) Set(
	ctx context.Context,
	f *domain.Flight,
) error {
	key := "flight:" + f.ID.String()
	data, err := json.Marshal(*f)
	if err != nil {
		return err
	}
	if err := c.cl.Set(ctx, key, data, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (c *flightCache) GetByID(
	ctx context.Context,
	id *uuid.UUID,
) (*domain.Flight, error) {
	key := "flight:" + id.String()
	data, err := c.cl.Get(ctx, key).Result()
	if err == rds.Nil {
		return nil, fmt.Errorf("key doesn't exist")
	} else if err != nil {
		return nil, fmt.Errorf("unexpected error")
	}
	var f domain.Flight
	if err := json.Unmarshal([]byte(data), &f); err != nil {
		return nil, err
	}
	return &f, nil
}

func (c *flightCache) Delete(
	ctx context.Context,
	id *uuid.UUID,
) error {
	return nil
}
