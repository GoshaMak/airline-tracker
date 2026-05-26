package repository

import (
	"api/internal/fleet/domain"
	"context"
)

type AircraftRepository interface {
	SaveAircraft(ctx context.Context, a domain.Aircraft) error

	List(ctx context.Context) ([]domain.Aircraft, error)
}
