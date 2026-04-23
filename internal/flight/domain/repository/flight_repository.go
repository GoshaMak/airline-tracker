package repository

import (
	"airline-tracker/internal/flight/domain"
	"context"
)

type FlightRepository interface {
	SaveFlight(ctx context.Context, flight domain.Flight) error

	Exists(ctx context.Context, id uint32) (domain.Flight, error)

	UpdateByID(ctx context.Context, id uint32) error

	ListAllFlights(ctx context.Context) ([]domain.Flight, error)
}
