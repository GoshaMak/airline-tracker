package service

import (
	airportDomain "airline-tracker/internal/airport/domain"
	airportRepository "airline-tracker/internal/airport/domain/repository"
	fleetDomain "airline-tracker/internal/fleet/domain"
	fleetRepository "airline-tracker/internal/fleet/domain/repository"
	flightDomain "airline-tracker/internal/flight/domain"
	flightRepository "airline-tracker/internal/flight/domain/repository"
	"context"

	"github.com/samber/do/v2"
)

type AdminService struct {
	flightRepository        flightRepository.FlightRepository
	aircraftRepository      fleetRepository.AircraftRepository
	aircraftModelRepository fleetRepository.AircraftModelRepository
	airportRepository       airportRepository.AirportRepository
	gateRepository          airportRepository.GateRepository
}

func NewAdminService(i do.Injector) (*AdminService, error) {
	return &AdminService{
		flightRepository:        do.MustInvokeAs[flightRepository.FlightRepository](i),
		aircraftRepository:      do.MustInvokeAs[fleetRepository.AircraftRepository](i),
		aircraftModelRepository: do.MustInvokeAs[fleetRepository.AircraftModelRepository](i),
		airportRepository:       do.MustInvokeAs[airportRepository.AirportRepository](i),
		gateRepository:          do.MustInvokeAs[airportRepository.GateRepository](i),
	}, nil
}

func (s *AdminService) AddFlight(
	flight *flightDomain.Flight,
	aircraft *fleetDomain.Aircraft,
	departureAirport, arrivalAirport *airportDomain.Airport,
	departureGate, arrivalGate *airportDomain.Gate,
) error {
	if err := s.flightRepository.Save(context.Background(),
		flight, aircraft, departureAirport, arrivalAirport,
		departureGate, arrivalGate); err != nil {
		return err
	}
	return nil
}

func (s *AdminService) AddAircraft(a *fleetDomain.Aircraft) error {
	if err := s.aircraftRepository.Save(context.Background(), a); err != nil {
		return err
	}
	return nil
}

func (s *AdminService) AddAirport(a *airportDomain.Airport) error {
	if err := s.airportRepository.Save(context.Background(), a); err != nil {
		return err
	}
	return nil
}

func (s *AdminService) AddGate(g *airportDomain.Gate) error {
	if err := s.gateRepository.Save(context.Background(), g); err != nil {
		return err
	}
	return nil
}

func (s *AdminService) AddAircraftModel(m *fleetDomain.AircraftModel) error {
	if err := s.aircraftModelRepository.Save(context.Background(), m); err != nil {
		return err
	}
	return nil
}
