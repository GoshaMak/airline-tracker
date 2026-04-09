package repository

import (
	"airline-tracker/internal/airport/domain"
	"context"
)

type AirportRepository interface {
	Save(ctx context.Context, a *domain.Airport) error
}
