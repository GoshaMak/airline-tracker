package service

import (
	"airline-tracker/internal/fleet/domain"
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

func (s *AircraftModelService) AddAircraftModel(m *domain.AircraftModel) error {
	if err := s.repository.Save(context.Background(), m); err != nil {
		return err
	}
	return nil
}
