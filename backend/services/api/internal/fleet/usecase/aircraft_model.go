package usecase

import (
	"api/internal/fleet/command"
	"api/internal/fleet/domain"
	"api/internal/fleet/domain/repository"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
	const op = "AircraftModelUsecase.AddAircraftModel"
	am, err := command.ToDomainCreateAircraftModelCommand(cmd)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := uc.repo.SaveAircraftModel(context.Background(), am); err != nil {
		if errors.Is(err, repository.ErrAircraftModelAlreadyExists) {
			return ErrAircraftModelAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (uc *AircraftModelUsecase) GetById(amId uuid.UUID) (domain.AircraftModel, error) {
	const op = "AircraftModelUsecase.GetById"
	amd, err := uc.repo.GetAircraftModelById(context.Background(), amId)
	if err != nil {
		if errors.Is(err, repository.ErrAircraftModelNotFound) {
			return domain.AircraftModel{}, ErrAircraftModelNotFound
		}
		return domain.AircraftModel{}, fmt.Errorf("%s: %w", op, err)
	}
	return amd, nil
}
