package usecase

import (
	"api/internal/airport/command"
	"api/internal/airport/domain"
	"api/internal/airport/domain/repository"
	"api/internal/pagination"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type GateUsecase struct {
	repo repository.GateRepository
}

type paginatedGateRepository interface {
	ListPage(context.Context, pagination.Params) ([]domain.Gate, error)
}

func (uc *GateUsecase) ListGatesPage(params pagination.Params) (pagination.Page[domain.Gate], error) {
	if repo, ok := uc.repo.(paginatedGateRepository); ok {
		gates, err := repo.ListPage(context.Background(), params)
		if err != nil {
			return pagination.Page[domain.Gate]{}, fmt.Errorf("GateUsecase.ListGatesPage: %w", err)
		}
		return pagination.FromFetched(gates, params, func(g domain.Gate) uuid.UUID { return g.Id }), nil
	}
	gates, err := uc.ListGates()
	if err != nil {
		return pagination.Page[domain.Gate]{}, err
	}
	return pagination.Paginate(gates, params, func(g domain.Gate) uuid.UUID { return g.Id }), nil
}

func NewGateUsecase(i do.Injector) (*GateUsecase, error) {
	return &GateUsecase{
		repo: do.MustInvoke[repository.GateRepository](i),
	}, nil
}

func (uc *GateUsecase) CreateGate(cmd *command.CreateGateCommand) error {
	const op = "GateUsecase.CreateGate"
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

func (uc *GateUsecase) ListGates() ([]domain.Gate, error) {
	const op = "GateUsecase.ListGates"
	gs, err := uc.repo.List(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return gs, nil
}
