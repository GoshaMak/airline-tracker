package repository

import (
	"airline-tracker/internal/fleet/domain"
	"context"
)

type AircraftModelRepository interface {
	SaveAircraftModel(ctx context.Context, a domain.AircraftModel) error

	Exists(ctx context.Context, a domain.AircraftModel) (domain.AircraftModel, error)
}
