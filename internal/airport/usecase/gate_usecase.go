package usecase

import (
	"airline-tracker/internal/airport/command"
	"airline-tracker/internal/airport/domain/repository"
	"context"

	"github.com/samber/do/v2"
)

type GateUsecase struct {
	repo repository.GateRepository
}

func NewGateUsecase(i do.Injector) (*GateUsecase, error) {
	return &GateUsecase{
		repo: do.MustInvokeAs[repository.GateRepository](i),
	}, nil
}

func (uc *GateUsecase) AddGate(cmd *command.AddGateCommand) error {
	g, err := command.CommandToGateDomain(cmd)
	if err != nil {
		return err
	}
	if err := uc.repo.Save(context.Background(), g); err != nil {
		return err
	}
	return nil
}
