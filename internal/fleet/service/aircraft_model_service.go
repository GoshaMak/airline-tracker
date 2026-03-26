package service

import (
	"airline-tracker/internal/fleet/command"
	"airline-tracker/internal/fleet/domain/repository"
	"context"

	"github.com/samber/do/v2"
)

type AircraftModelService struct {
	repository repository.AircraftModelRepository
}

func NewAircraftModelService(i do.Injector) (*AircraftModelService, error) {
	return &AircraftModelService{
		repository: do.MustInvokeAs[repository.AircraftModelRepository](i),
	}, nil
}

func (s *AircraftModelService) AddAircraftModel(cmd *command.CreateAircraftModelCommand) error {
	am, err := command.ToDomainCreateAircraftModelCommand(cmd)
	if err != nil {
		return err
	}
	if err := s.repository.Save(context.Background(), am); err != nil {
		return err
	}
	return nil
}
