package redis

import (
	"api/internal/airport/domain/repository"

	rds "github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
)

type airportCache struct {
	cln *rds.Client
}

func NewAirportCache(i do.Injector) (repository.AirportCache, error) {
	return &airportCache{
		cln: do.MustInvokeAs[*rds.Client](i),
	}, nil
}
