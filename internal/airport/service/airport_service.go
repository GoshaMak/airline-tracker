package service

import (
	"airline-tracker/internal/airport/domain"
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

func (s *AirportService) AddAirport(a *domain.Airport) error {
	if err := s.repository.Save(context.Background(), a); err != nil {
		return err
	}
	return nil
}
