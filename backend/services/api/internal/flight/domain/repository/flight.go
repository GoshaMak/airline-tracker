package repository

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/flight/domain"
	userDomain "api/internal/user/domain"
	"context"

	"github.com/google/uuid"
)

type FlightRepository interface {
	Save(ctx context.Context, flight domain.Flight) error

	Exist(ctx context.Context, fid uuid.UUID) (domain.Flight, error)

	Update(ctx context.Context, ufi domain.UpdateFlightInfo) error

	ListFlights(ctx context.Context) ([]domain.Flight, error)

	GetFlightRoute(ctx context.Context, fid uuid.UUID) (domain.FlightRoute, error)

	ListSubscribers(ctx context.Context, fid uuid.UUID) ([]userDomain.User, error)

	GetFlightAirports(
		ctx context.Context,
		fid uuid.UUID,
	) (dep airportDomain.Airport, arr airportDomain.Airport, err error)
}
