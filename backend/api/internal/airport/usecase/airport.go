package usecase

import (
	"api/internal/airport/command"
	"api/internal/airport/domain"
	"api/internal/airport/domain/repository"
	"api/internal/airport/query"
	"api/internal/pagination"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type AirportUsecase struct {
	repo repository.AirportRepository
}

type paginatedAirportRepository interface {
	ListAirportsPage(context.Context, pagination.Params) ([]domain.Airport, error)
}

func (uc *AirportUsecase) ListAirportsPage(params pagination.Params) (pagination.Page[domain.Airport], error) {
	if repo, ok := uc.repo.(paginatedAirportRepository); ok {
		airports, err := repo.ListAirportsPage(context.Background(), params)
		if err != nil {
			return pagination.Page[domain.Airport]{}, fmt.Errorf("AirportUsecase.ListAirportsPage: %w", err)
		}
		return pagination.FromFetched(airports, params, func(a domain.Airport) uuid.UUID { return a.ID }), nil
	}
	queryResult, err := uc.ListAirports()
	if err != nil {
		return pagination.Page[domain.Airport]{}, err
	}
	return pagination.Paginate(queryResult.Airports, params, func(a domain.Airport) uuid.UUID { return a.ID }), nil
}

func NewAirportUsecase(i do.Injector) (*AirportUsecase, error) {
	return &AirportUsecase{
		repo: do.MustInvoke[repository.AirportRepository](i),
	}, nil
}

func (uc *AirportUsecase) CreateAirport(cmd *command.CreateAirportCommand) error {
	const op = "AirportUsecase.CreateAirport"
	a, err := command.CommandToAirportDomain(cmd)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := uc.repo.Save(context.Background(), a); err != nil {
		if errors.Is(err, repository.ErrAirportAlreadyExists) {
			return ErrAirportAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (uc *AirportUsecase) ListAirports() (query.ListAirportsQuery, error) {
	const op = "AirportUsecase.ListAirports"
	airports, err := uc.repo.ListAirports(context.Background())
	if err != nil {
		return query.ListAirportsQuery{}, fmt.Errorf("%s: %w", op, err)
	}

	q := query.ListAirportsQuery{}
	for _, a := range airports {
		q.Airports = append(q.Airports, a)
	}

	return q, nil
}
