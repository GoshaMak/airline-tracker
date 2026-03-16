package repository

import (
	"airline-tracker/internal/domain"
	"context"
)

type FlightRepository interface {
	Save(
		ctx context.Context,
		flight *domain.Flight,
		aircraft *domain.Aircraft,
		departureAirport, arrivalAirport *domain.Airport,
		departureGate, arrivalGate *domain.Gate,
	) error

	Exists(ctx context.Context, id uint32) (*domain.Flight, error)

	UpdateByID(ctx context.Context, id uint32) error

	ListAllFlights(ctx context.Context) ([]domain.Flight, error)
}
