package repository

import (
	airportDomain "airline-tracker/internal/airport/domain"
	fleetDomain "airline-tracker/internal/fleet/domain"
	flightDomain "airline-tracker/internal/flight/domain"
	"context"
)

type FlightRepository interface {
	Save(
		ctx context.Context,
		flight *flightDomain.Flight,
		aircraft *fleetDomain.Aircraft,
		departureAirport, arrivalAirport *airportDomain.Airport,
		departureGate, arrivalGate *airportDomain.Gate,
	) error

	Exists(ctx context.Context, id uint32) (*flightDomain.Flight, error)

	UpdateByID(ctx context.Context, id uint32) error

	ListAllFlights(ctx context.Context) ([]flightDomain.Flight, error)
}
