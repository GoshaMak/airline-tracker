package service

import (
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
