package repository

import (
	flightDomain "airline-tracker/internal/flight/domain"
	"context"
)

type FlightRepository interface {
	Save(ctx context.Context, flight *flightDomain.Flight) error

	Exists(ctx context.Context, id uint32) (*flightDomain.Flight, error)

	UpdateByID(ctx context.Context, id uint32) error

	ListAllFlights(ctx context.Context) ([]flightDomain.Flight, error)
}
