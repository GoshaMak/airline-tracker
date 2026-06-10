package repository

import (
	"api/internal/fleet/domain"
	"context"

	"github.com/google/uuid"
)

type AircraftModelRepository interface {
	SaveAircraftModel(ctx context.Context, am domain.AircraftModel) error

	GetAircraftModelById(ctx context.Context, id uuid.UUID) (domain.AircraftModel, error)
}
