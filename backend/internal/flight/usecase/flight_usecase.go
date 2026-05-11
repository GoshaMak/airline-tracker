package usecase

import (
	"airline-tracker/internal/flight/command"
	"airline-tracker/internal/flight/domain"
	"airline-tracker/internal/flight/domain/repository"
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type FlightUsecase struct {
	repo  repository.FlightRepository
	cache repository.FlightCache
}

func NewFlightUsecase(i do.Injector) (*FlightUsecase, error) {
	return &FlightUsecase{
		repo:  do.MustInvokeAs[repository.FlightRepository](i),
		cache: do.MustInvokeAs[repository.FlightCache](i),
	}, nil
}

func (uc *FlightUsecase) ListFlights() ([]domain.Flight, error) {
	op := "FlightUsecase.ListFlights"
	var flights []domain.Flight
	flights, err := uc.cache.GetFlights(context.Background())
	if err != nil {
		if err != repository.ErrCacheEmpty {
			slog.Warn("error in cache", "err", err)
		}
	} else {
		slog.Info("list flights cached")
		return flights, nil
	}

	flights, err = uc.repo.ListFlights(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// mb do it inside goroutine
	if err := uc.cache.SaveFlights(context.Background(), flights); err != nil {
		slog.Error("error in cache", "err", err)
		return flights, ErrCacheSave
	}

	return flights, nil
}

func (uc *FlightUsecase) CreateFlight(cmd command.CreateFlightCommand) error {
	op := "FlightUsecase.CreateFlight"
	f, err := command.CommandToFlightDomain(cmd)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := uc.repo.Save(context.Background(), f); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := uc.cache.Save(context.Background(), f); err != nil {
		slog.Debug(op, "err", err)
		return ErrCacheSave
	}
	return nil
}

func (uc *FlightUsecase) GetFlightById(fid uuid.UUID) (domain.Flight, error) {
	f := domain.Flight{}
	return f, nil
}
