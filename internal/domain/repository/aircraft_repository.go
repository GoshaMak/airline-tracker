package repository

import (
	"airline-tracker/internal/domain"
	"context"
)

type AircraftRepository interface {
	Save(ctx context.Context, a *domain.Aircraft) error

	Exists(ctx context.Context, a *domain.Aircraft) (*domain.Aircraft, error)
}
