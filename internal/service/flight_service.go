package service

import (
	"airline-tracker/internal/domain/repository"

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
