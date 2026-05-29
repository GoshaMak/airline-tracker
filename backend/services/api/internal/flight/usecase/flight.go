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
	outboxRepo outboxRepository.OutboxRepository
}

func NewFlightUsecase(i do.Injector) (*FlightUsecase, error) {
	return &FlightUsecase{
		repo:       do.MustInvoke[repository.FlightRepository](i),
		outboxRepo: do.MustInvoke[outboxRepository.OutboxRepository](i),
	}, nil
}

func (uc *FlightUsecase) ListFlights() ([]domain.Flight, error) {
	const op = "FlightUsecase.ListFlights"
	flights, err := uc.repo.ListFlights(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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
	slog.Debug(op + ": flight updated")

	// TODO: figure how to do it more generally (DRY)
	go func() {
		subs, err := uc.repo.ListSubscribers(context.Background(), ufi.FlightId)
		if err != nil {
			slog.Error(op, "err", err)
			return
		}
		if len(subs) == 0 {
			return
		}

		dep, arr, err := uc.repo.GetFlightAirports(context.Background(), ufi.FlightId)
		if err != nil {
			slog.Error(op, "err", err)
			return
		}
		topic := os.Getenv("FLIGHT_UPDATED_TOPIC") // TODO: get it from config?
		payload, err := uc.formPayload(ufi, dep, arr)
		if err != nil {
			slog.Error(op, "err", err)
			return
		}
		slog.Debug(op+": payload formed", "payload", payload)
		ob, err := publisherDomain.NewOutbox(topic, &payload)
		if err != nil {
			slog.Error(op, "err", err)
			return
		}
		if err := uc.outboxRepo.Save(context.Background(), ob); err != nil {
			slog.Error(op, "err", err)
			return
		}
		slog.Debug(op+": saved to outbox", "outbox", ob)
	}()

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
