package service

import (
	"airline-tracker/internal/airport/command"
	"airline-tracker/internal/airport/domain/repository"
	"context"

	"github.com/samber/do/v2"
)

type AirportService struct {
	repository repository.AirportRepository
}

func NewAirportService(i do.Injector) (*AirportService, error) {
	return &AirportService{
		repository: do.MustInvokeAs[repository.AirportRepository](i),
	}, nil
}

func (s *AirportService) AddAirport(cmd *command.AddAirportCommand) error {
	a, err := command.CommandToAirportDomain(cmd)
	if err != nil {
		return err
	}
	if err := s.repository.Save(context.Background(), a); err != nil {
		return err
	}
	return nil
}
