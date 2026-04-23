package usecase

import (
	"airline-tracker/internal/airport/command"
	"airline-tracker/internal/airport/domain/repository"
	"context"
	"errors"
	"fmt"

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
	op := "GateUsecase.CreateGate"
	g, err := command.CommandToGateDomain(cmd)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := uc.repo.Save(context.Background(), g); err != nil {
		if errors.Is(err, repository.ErrGateAlreadyExists) {
			return ErrGateAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
