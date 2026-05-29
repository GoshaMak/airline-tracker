package repository

import (
	"api/internal/airport/domain"
	"context"
)

type AirportRepository interface {
	Save(ctx context.Context, a domain.Airport) error

	ListAirports(ctx context.Context) ([]domain.Airport, error)
}
