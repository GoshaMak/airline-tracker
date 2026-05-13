package usecase

import (
	"api/internal/fleet/command"
	"api/internal/fleet/domain/repository"
	"context"

	"github.com/samber/do/v2"
)

type AircraftUsecase struct {
	repo repository.AircraftRepository
}

func NewAircraftUsecase(i do.Injector) (*AircraftUsecase, error) {
	return &AircraftUsecase{
		repo: do.MustInvokeAs[repository.AircraftRepository](i),
	}, nil
}

func (uc *AircraftUsecase) AddAircraft(cmd command.CreateAircraftCommand) error {
	a, err := command.ToDomainCreateAircraftCommand(cmd)
	if err != nil {
		return err
	}
	if err := uc.repo.SaveAircraft(context.Background(), a); err != nil {
		return err
	}
	return nil
}
