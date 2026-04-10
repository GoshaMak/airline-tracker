package usecase

import (
	"airline-tracker/internal/airport/command"
	"airline-tracker/internal/airport/domain/repository"
	"airline-tracker/internal/airport/infra"
	"context"
	"errors"

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

func (uc *GateUsecase) CreateGate(cmd *command.CreateGateCommand) error {
	g, err := command.CommandToGateDomain(cmd)
	if err != nil {
		return err
	}

	if err := uc.repo.Save(context.Background(), g); err != nil {
		if errors.Is(err, infra.ErrGateAlreadyExists) {
			return ErrGateAlreadyExists
		}
		return ErrUnexpected
	}

	return nil
}
