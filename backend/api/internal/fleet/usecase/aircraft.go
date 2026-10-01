package usecase

import (
	"api/internal/fleet/command"
	"api/internal/fleet/domain"
	"api/internal/fleet/domain/repository"
	"api/internal/pagination"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type AircraftUsecase struct {
	repo repository.AircraftRepository
}

type paginatedAircraftRepository interface {
	ListPage(context.Context, pagination.Params) ([]domain.Aircraft, error)
}

func (uc *AircraftUsecase) ListAircraftsPage(params pagination.Params) (pagination.Page[domain.Aircraft], error) {
	if repo, ok := uc.repo.(paginatedAircraftRepository); ok {
		aircraft, err := repo.ListPage(context.Background(), params)
		if err != nil {
			return pagination.Page[domain.Aircraft]{}, fmt.Errorf("AircraftUsecase.ListAircraftsPage: %w", err)
		}
		return pagination.FromFetched(aircraft, params, func(a domain.Aircraft) uuid.UUID { return a.Id }), nil
	}
	aircraft, err := uc.ListAircrafts()
	if err != nil {
		return pagination.Page[domain.Aircraft]{}, err
	}
	return pagination.Paginate(aircraft, params, func(a domain.Aircraft) uuid.UUID { return a.Id }), nil
}

func NewAircraftUsecase(i do.Injector) (*AircraftUsecase, error) {
	return &AircraftUsecase{
		repo: do.MustInvoke[repository.AircraftRepository](i),
	}, nil
}

func (uc *AircraftUsecase) CreateAircraft(cmd command.CreateAircraftCommand) error {
	const op = "AircraftUsecase.CreateAircraft"
	a, err := command.ToDomainCreateAircraftCommand(cmd)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := uc.repo.SaveAircraft(context.Background(), a); err != nil {
		if errors.Is(err, repository.ErrAircraftAlreadyExists) {
			return ErrAircraftAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (uc *AircraftUsecase) ListAircrafts() ([]domain.Aircraft, error) {
	const op = "AircraftUsecase.ListAircrafts"
	as, err := uc.repo.List(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return as, nil
}
