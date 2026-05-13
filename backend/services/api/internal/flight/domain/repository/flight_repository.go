package repository

import (
	"api/internal/flight/domain"
	"context"

	"github.com/google/uuid"
)

type FlightRepository interface {
	Save(ctx context.Context, flight domain.Flight) error

	Exist(ctx context.Context, fid uuid.UUID) (domain.Flight, error)

	UpdateById(ctx context.Context, fid uuid.UUID) error

	ListFlights(ctx context.Context) ([]domain.Flight, error)

	GetFlightRoute(ctx context.Context, fid uuid.UUID) (domain.FlightRoute, error)
}
