package service

import (
	"airline-tracker/internal/fleet/command"
	"airline-tracker/internal/fleet/domain/repository"
	"context"

	"github.com/samber/do/v2"
)

type AircraftService struct {
	repository repository.AircraftRepository
}

func NewAircraftService(i do.Injector) (*AircraftService, error) {
	return &AircraftService{
		repository: do.MustInvokeAs[repository.AircraftRepository](i),
	}, nil
}

func (s *AircraftService) AddAircraft(cmd *command.CreateAircraftCommand) error {
	a, err := command.ToDomainCreateAircraftCommand(cmd)
	if err != nil {
		return err
	}
	if err := s.repository.Save(context.Background(), a); err != nil {
		return err
	}
	return nil
}
