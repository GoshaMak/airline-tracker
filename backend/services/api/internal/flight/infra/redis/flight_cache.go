package redis

import (
	"api/internal/flight/domain"
	"api/internal/flight/domain/repository"
	"api/internal/flight/infra/redis/model"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	"github.com/google/uuid"
	rds "github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
)

type flightCache struct {
	cln       *rds.Client
	mu        sync.RWMutex
	flightIds []string // INFO: from here i'll take ids while listing flights
}

func NewFlightCache(i do.Injector) (repository.FlightCache, error) {
	return &flightCache{
		cln:       do.MustInvoke[*rds.Client](i),
		mu:        sync.RWMutex{},
		flightIds: make([]string, 0),
	}, nil
}

func formKey(id string) string {
	return "flight:" + id
}

func (c *flightCache) Save(
	ctx context.Context,
	flight domain.Flight,
) error {
	op := "FlightCache.Save"
	fm, err := model.FlightDomainToModel(flight)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	data, err := json.Marshal(fm)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	idStr := flight.Id.String()
	key := formKey(idStr)
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.cln.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	c.flightIds = append(c.flightIds, idStr)

	return nil
}

func (c *flightCache) GetById(
	ctx context.Context,
	id uuid.UUID,
) (domain.Flight, error) {
	op := "FlightCache.GetByID"
	idStr := id.String()
	key := formKey(idStr)

	c.mu.RLock()
	data, err := c.cln.Get(ctx, key).Result()
	c.mu.RUnlock()

	if err == rds.Nil {
		return domain.Flight{}, repository.ErrCacheEmpty
	} else if err != nil {
		return domain.Flight{}, fmt.Errorf("%s: %w", op, err)
	}

	var fm model.Flight
	if err := json.Unmarshal([]byte(data), &fm); err != nil {
		return domain.Flight{}, fmt.Errorf("%s: %w", op, err)
	}
	f, err := model.FlightModelToDomain(fm)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("%s: %w", op, err)
	}

	return f, nil
}

func (c *flightCache) DeleteById(
	ctx context.Context,
	id uuid.UUID,
) error {
	op := "FlightCache.Delete"
	idStr := id.String()
	key := formKey(idStr)

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.cln.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if index := slices.Index(c.flightIds, idStr); index != -1 {
		c.flightIds = slices.Delete(c.flightIds, index, index+1)
	}

	return nil
}

func (c *flightCache) SaveFlights(ctx context.Context, flights []domain.Flight) error {
	op := "FlightCache.SaveFlights"
	var (
		err    error
		errInd int
	)
	for i, flight := range flights {
		if err = c.Save(ctx, flight); err != nil {
			err = fmt.Errorf("%s: %w", op, err)
			errInd = i
			break
		}
	}
	if err != nil {
		for i := range errInd + 1 {
			c.DeleteById(ctx, flights[i].Id)
		}
		return err
	}
	return nil
}

func (c *flightCache) GetFlights(ctx context.Context) ([]domain.Flight, error) {
	op := "FlightCache.GetFlights"
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.flightIds) == 0 {
		return nil, repository.ErrCacheEmpty
	}

	flights := make([]domain.Flight, len(c.flightIds))
	var fm model.Flight
	for i, idStr := range c.flightIds {
		key := formKey(idStr)
		data, err := c.cln.Get(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if err := json.Unmarshal([]byte(data), &fm); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		flight, err := model.FlightModelToDomain(fm)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		flights[i] = flight
	}

	return flights, nil
}
