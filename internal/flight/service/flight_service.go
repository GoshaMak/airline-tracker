package service

import (
	airportDomain "airline-tracker/internal/airport/domain"
	fleetDomain "airline-tracker/internal/fleet/domain"
	"airline-tracker/internal/flight/domain"
	"airline-tracker/internal/flight/domain/repository"
	"context"

	"github.com/samber/do/v2"
)

type FlightService struct {
	repository repository.FlightRepository
}

func NewFlightService(i do.Injector) (*FlightService, error) {
	return &FlightService{
		repository: do.MustInvokeAs[repository.FlightRepository](i),
	}, nil
}

func (s *FlightService) ListAllFlights() ([]domain.Flight, error) {
	flights, err := s.repository.ListAllFlights(context.Background())
	if err != nil {
		return nil, err
	}
	return flights, nil
}

func (s *FlightService) AddFlight(
	flight *domain.Flight,
	aircraft *fleetDomain.Aircraft,
	departureAirport, arrivalAirport *airportDomain.Airport,
	departureGate, arrivalGate *airportDomain.Gate,
) error {
	if err := s.repository.Save(context.Background(),
		flight, aircraft, departureAirport, arrivalAirport,
		departureGate, arrivalGate); err != nil {
		return err
	}
	return nil
}
