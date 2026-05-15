package usecase

import (
	"api/internal/airport/command"
	"api/internal/airport/domain/repository"
	"api/internal/airport/query"
	"context"
	"errors"
	"fmt"

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
