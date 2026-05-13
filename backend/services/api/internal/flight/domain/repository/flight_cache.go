package repository

import (
	"api/internal/flight/domain"
	"context"

	"github.com/google/uuid"
)

type FlightCache interface {
	Save(ctx context.Context, f domain.Flight) error

	GetById(ctx context.Context, id uuid.UUID) (domain.Flight, error)

	DeleteById(ctx context.Context, id uuid.UUID) error

	SaveFlights(ctx context.Context, flights []domain.Flight) error

	GetFlights(ctx context.Context) ([]domain.Flight, error)
}
