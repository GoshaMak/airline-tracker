package service

import (
	"airline-tracker/internal/airport/command"
	"airline-tracker/internal/airport/domain/repository"
	"context"

	"github.com/samber/do/v2"
)

type GateService struct {
	repository repository.GateRepository
}

func NewGateService(i do.Injector) (*GateService, error) {
	return &GateService{
		repository: do.MustInvokeAs[repository.GateRepository](i),
	}, nil
}

func (s *GateService) AddGate(cmd *command.AddGateCommand) error {
	g, err := command.CommandToGateDomain(cmd)
	if err != nil {
		return err
	}
	if err := s.repository.Save(context.Background(), g); err != nil {
		return err
	}
	return nil
}
