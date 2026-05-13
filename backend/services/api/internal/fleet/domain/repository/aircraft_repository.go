package repository

import (
	"api/internal/fleet/domain"
	"context"
)

type AircraftRepository interface {
	SaveAircraft(ctx context.Context, a domain.Aircraft) error

	Exists(ctx context.Context, a domain.Aircraft) (domain.Aircraft, error)
}
