package repository

import (
	"airline-tracker/internal/airport/domain"
	"context"
)

type AirportRepository interface {
	Save(ctx context.Context, a *domain.Airport) error

	Exists(ctx context.Context, a *domain.Airport) (*domain.Airport, error)
}
