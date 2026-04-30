package usecase

import (
	"airline-tracker/internal/fleet/command"
	"airline-tracker/internal/fleet/domain/repository"
	"context"

	"github.com/samber/do/v2"
)

type AircraftModelUsecase struct {
	repo repository.AircraftModelRepository
}

func NewAircraftModelUsecase(i do.Injector) (*AircraftModelUsecase, error) {
	return &AircraftModelUsecase{
		repo: do.MustInvokeAs[repository.AircraftModelRepository](i),
	}, nil
}

func (uc *AircraftModelUsecase) AddAircraftModel(cmd command.CreateAircraftModelCommand) error {
	am, err := command.ToDomainCreateAircraftModelCommand(cmd)
	if err != nil {
		return err
	}
	if err := uc.repo.SaveAircraftModel(context.Background(), am); err != nil {
		return err
	}
	return nil
}
