package repository

import (
	"airline-tracker/internal/flight/domain"
	"context"
)

type FlightRepository interface {
	Save(ctx context.Context, flight domain.Flight) error

	Exists(ctx context.Context, id uint32) (domain.Flight, error)

	UpdateById(ctx context.Context, id uint32) error

	ListFlights(ctx context.Context) ([]domain.Flight, error)
}
