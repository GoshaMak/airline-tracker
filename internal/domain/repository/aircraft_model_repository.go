package repository

import (
	"airline-tracker/internal/domain"
	"context"
)

type AircraftModelRepository interface {
	Save(ctx context.Context, a *domain.AircraftModel) error

	Exists(ctx context.Context, a *domain.AircraftModel) (*domain.AircraftModel, error)
}
