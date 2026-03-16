package service

import (
	"airline-tracker/internal/domain"
	"airline-tracker/internal/domain/repository"
	"context"

	"github.com/samber/do/v2"
)

type AdminService struct {
	flightRepository        repository.FlightRepository
	aircraftRepository      repository.AircraftRepository
	airportRepository       repository.AirportRepository
	gateRepository          repository.GateRepository
	aircraftModelRepository repository.AircraftModelRepository
}

func NewAdminService(i do.Injector) (*AdminService, error) {
	return &AdminService{
		flightRepository:        do.MustInvokeAs[repository.FlightRepository](i),
		aircraftRepository:      do.MustInvokeAs[repository.AircraftRepository](i),
		airportRepository:       do.MustInvokeAs[repository.AirportRepository](i),
		gateRepository:          do.MustInvokeAs[repository.GateRepository](i),
		aircraftModelRepository: do.MustInvokeAs[repository.AircraftModelRepository](i),
	}, nil
}

func (s *AdminService) AddFlight(
	flight *domain.Flight,
	aircraft *domain.Aircraft,
	departureAirport, arrivalAirport *domain.Airport,
	departureGate, arrivalGate *domain.Gate,
) error {
	if err := s.flightRepository.Save(context.Background(),
		flight, aircraft, departureAirport, arrivalAirport,
		departureGate, arrivalGate); err != nil {
		return err
	}
	return nil
}

func (s *AdminService) AddAircraft(a *domain.Aircraft) error {
	if err := s.aircraftRepository.Save(context.Background(), a); err != nil {
		return err
	}
	return nil
}

func (s *AdminService) AddAirport(a *domain.Airport) error {
	if err := s.airportRepository.Save(context.Background(), a); err != nil {
		return err
	}
	return nil
}

func (s *AdminService) AddGate(g *domain.Gate) error {
	if err := s.gateRepository.Save(context.Background(), g); err != nil {
		return err
	}
	return nil
}

func (s *AdminService) AddAircraftModel(m *domain.AircraftModel) error {
	if err := s.aircraftModelRepository.Save(context.Background(), m); err != nil {
		return err
	}
	return nil
}
