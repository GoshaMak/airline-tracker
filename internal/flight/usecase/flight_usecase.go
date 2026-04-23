package service

import (
	"airline-tracker/internal/flight/command"
	"airline-tracker/internal/flight/domain"
	"airline-tracker/internal/flight/domain/repository"
	"context"

	"github.com/samber/do/v2"
)

type FlightUsecase struct {
	repo repository.FlightRepository
}

func NewFlightUsecase(i do.Injector) (*FlightUsecase, error) {
	return &FlightUsecase{
		repo: do.MustInvokeAs[repository.FlightRepository](i),
	}, nil
}

func (uc *FlightUsecase) ListAllFlights() ([]domain.Flight, error) {
	flights, err := uc.repo.ListAllFlights(context.Background())
	if err != nil {
		return nil, err
	}
	return flights, nil
}

func (uc *FlightUsecase) AddFlight(cmd *command.AddFlightCommand) error {
	f, err := command.CommandToFlightDomain(cmd)
	if err != nil {
		return err
	}
	if err := uc.repo.SaveFlight(context.Background(), f); err != nil {
		return err
	}
	return nil
}
