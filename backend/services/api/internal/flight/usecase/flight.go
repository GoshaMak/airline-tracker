package usecase

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/flight/command"
	"api/internal/flight/domain"
	"api/internal/flight/domain/repository"
	publisherDomain "api/internal/publisher/domain"
	outboxRepository "api/internal/publisher/domain/repository"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"shared/common"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type FlightUsecase struct {
	repo       repository.FlightRepository
	cache      repository.FlightCache
	outboxRepo outboxRepository.OutboxRepository
}

func NewFlightUsecase(i do.Injector) (*FlightUsecase, error) {
	return &FlightUsecase{
		repo:       do.MustInvokeAs[repository.FlightRepository](i),
		cache:      do.MustInvokeAs[repository.FlightCache](i),
		outboxRepo: do.MustInvokeAs[outboxRepository.OutboxRepository](i),
	}, nil
}

func (uc *FlightUsecase) ListFlights() ([]domain.Flight, error) {
	const op = "FlightUsecase.ListFlights"
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
	const op = "FlightUsecase.CreateFlight"
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

func (uc *FlightUsecase) FlightById(fid uuid.UUID) (domain.Flight, error) {
	const op = "FlightUsecase.FlightById"
	f, err := uc.repo.Exist(context.Background(), fid)
	if err != nil {
		if errors.Is(err, repository.ErrFlightNotFound) {
			return domain.Flight{}, ErrFlightNotFound
		}
		return domain.Flight{}, fmt.Errorf("%s: %w", op, err)
	}
	return f, nil
}

func (uc *FlightUsecase) UpdateFlight(cmd command.UpdateFlightCommand) error {
	const op = "FlightUsecase.UpdateFlight"
	ufi, err := command.UpdateFlightCommandToDomain(cmd)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := uc.repo.Update(context.Background(), ufi); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	slog.Debug(op + ": user subscribed")

	// TODO: figure how to do it more generally (DRY)
	dep, arr, err := uc.repo.GetFlightAirports(context.Background(), ufi.FlightId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	topic := os.Getenv("FLIGHT_UPDATED_TOPIC") // TODO: get it from config?
	payload, err := uc.formPayload(ufi, dep, arr)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	slog.Debug(op+": payload formed", "payload", payload)
	ob, err := publisherDomain.NewOutbox(topic, &payload)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := uc.outboxRepo.Save(context.Background(), ob); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	slog.Debug(op+": saved to outbox", "outbox", ob)

	return nil
}

func (uc *FlightUsecase) formPayload(
	ufi domain.UpdateFlightInfo,
	departrueAirport,
	arrivalAirport airportDomain.Airport,
) (domain.FlightUpdatedPayload, error) {
	const op = "FlightUsecase.formPayload"
	users, err := uc.repo.ListSubscribers(context.Background(), ufi.FlightId)
	if err != nil {
		return domain.FlightUpdatedPayload{}, fmt.Errorf("%s: %w", op, err)
	}

	emails := make([]common.Email, len(users))
	for i, u := range users {
		emails[i] = u.Email
	}

	payload, err := domain.NewFlightUpdatedPayload(
		ufi.FlightId,
		emails,
		departrueAirport.Title,
		arrivalAirport.Title,
		ufi.ScheduledDeparture,
		ufi.ActualDeparture,
		ufi.ScheduledArrival,
		ufi.ActualArrival,
		ufi.Status,
		ufi.Plan,
	)
	if err != nil {
		return domain.FlightUpdatedPayload{}, fmt.Errorf("%s: %w", op, err)
	}
	return payload, nil
}
