package usecase

import (
	"airline-tracker/internal/airport/command"
	"airline-tracker/internal/airport/domain/repository"
	"airline-tracker/internal/airport/infra"
	"context"
	"errors"

	"github.com/samber/do/v2"
)

type AirportUsecase struct {
	repo repository.AirportRepository
}

func NewAirportUsecase(i do.Injector) (*AirportUsecase, error) {
	return &AirportUsecase{
		repo: do.MustInvokeAs[repository.AirportRepository](i),
	}, nil
}

func (uc *AirportUsecase) CreateAirport(cmd *command.CreateAirportCommand) error {
	a, err := command.CommandToAirportDomain(cmd)
	if err != nil {
		return err
	}

	if err := uc.repo.Save(context.Background(), a); err != nil {
		if errors.Is(err, infra.ErrAirportAlreadyExists) {
			return ErrAirportAlreadyExists
		}
		return ErrUnexpected
	}

	return nil
}
