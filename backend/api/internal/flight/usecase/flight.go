package usecase

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/flight/command"
	"api/internal/flight/domain"
	"api/internal/flight/usecase/repository"
	"api/internal/pagination"
	publisherDomain "api/internal/publisher/domain"
	outboxRepository "api/internal/publisher/infra/postgres"
	userDomain "api/internal/user/domain"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"shared/common"
	"shared/slicex"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type FlightClient interface {
	CreateFlight(ctx context.Context, cmd command.CreateFlightCommand) error
	FlightById(ctx context.Context, fid uuid.UUID) (domain.Flight, error)
	UpdateFlight(ctx context.Context, cmd command.UpdateFlightCommand) error
	ListFlights(ctx context.Context) ([]domain.Flight, error)
	ListFlightsPage(ctx context.Context, params pagination.Params) (pagination.Page[domain.Flight], error)
}

type FlightUsecase struct {
	repo       repository.FlightDB
	outboxRepo outboxRepository.OutboxRepository
}

func NewFlightUsecase(i do.Injector) (*FlightUsecase, error) {
	return &FlightUsecase{
		repo:       do.MustInvoke[repository.FlightDB](i),
		outboxRepo: do.MustInvoke[outboxRepository.OutboxRepository](i),
	}, nil
}

func (uc *FlightUsecase) CreateFlight(ctx context.Context, cmd command.CreateFlightCommand) error {
	const op = "FlightUsecase.CreateFlight"
	f, err := command.CommandToFlightDomain(cmd)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := uc.repo.Save(ctx, f); err != nil {
		if errors.Is(err, repository.ErrOperationNotPermitted) {
			return ErrOperationNotPermitted
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (uc *FlightUsecase) FlightById(ctx context.Context, fid uuid.UUID) (domain.Flight, error) {
	const op = "FlightUsecase.FlightById"
	f, err := uc.repo.Exist(ctx, fid)
	if err != nil {
		if errors.Is(err, repository.ErrFlightNotFound) {
			return domain.Flight{}, ErrFlightNotFound
		}
		if errors.Is(err, repository.ErrOperationNotPermitted) {
			return domain.Flight{}, ErrOperationNotPermitted
		}
		return domain.Flight{}, fmt.Errorf("%s: %w", op, err)
	}
	return f, nil
}

type flightUpdater interface {
	UpdateIfChanged(context.Context, domain.UpdateFlightInfo) (bool, error)
}

type transactionRunner interface {
	WithTransaction(context.Context, func(context.Context) error) error
}

func (uc *FlightUsecase) UpdateFlight(ctx context.Context, cmd command.UpdateFlightCommand) error {
	const op = "FlightUsecase.UpdateFlight"
	ufi, err := command.UpdateFlightCommandToDomain(cmd)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	work := func(txCtx context.Context) error {
		updater, ok := uc.repo.(flightUpdater)
		if !ok {
			return fmt.Errorf("%s: flight repo doesn't implement flightUpdater", op)
		}
		changed, err := updater.UpdateIfChanged(txCtx, ufi)
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}
		subs, err := uc.repo.ListSubscribers(txCtx, ufi.FlightId)
		if err != nil {
			return err
		}
		if len(subs) == 0 {
			return nil
		}
		dep, arr, err := uc.repo.GetFlightAirports(txCtx, ufi.FlightId)
		if err != nil {
			return err
		}
		payload, err := newFlightUpdatedPayload(ufi, dep, arr, subs)
		if err != nil {
			return err
		}
		ob, err := publisherDomain.NewOutbox(os.Getenv("FLIGHT_UPDATED_TOPIC"), &payload)
		if err != nil {
			return err
		}
		return uc.outboxRepo.Save(txCtx, ob)
	}
	tx, ok := uc.repo.(transactionRunner)
	if !ok {
		return fmt.Errorf("%s: flight repo doesn't implement transactionRunner", op)
	}
	err = tx.WithTransaction(ctx, work)
	if errors.Is(err, repository.ErrFlightNotFound) {
		return ErrFlightNotFound
	}
	if errors.Is(err, repository.ErrOperationNotPermitted) {
		return ErrOperationNotPermitted
	}
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	slog.DebugContext(ctx, op+": flight updated")
	return nil
}

func newFlightUpdatedPayload(
	ufi domain.UpdateFlightInfo,
	dep,
	arr airportDomain.Airport,
	users []userDomain.User,
) (domain.FlightUpdatedPayload, error) {
	emails := slicex.Map(users, func(u userDomain.User) common.Email { return u.Email })
	return domain.NewFlightUpdatedPayload(
		ufi.FlightId,
		emails,
		dep.Title,
		arr.Title,
		ufi.ScheduledDeparture,
		ufi.ActualDeparture,
		ufi.ScheduledArrival,
		ufi.ActualArrival,
		ufi.Status,
		ufi.Plan,
	)
}

func (uc *FlightUsecase) ListFlights(ctx context.Context) ([]domain.Flight, error) {
	const op = "FlightUsecase.ListFlights"
	flights, err := uc.repo.ListFlights(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrOperationNotPermitted) {
			return nil, ErrOperationNotPermitted
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return flights, nil
}

type paginatedFlightRepository interface {
	ListFlightsPage(context.Context, pagination.Params) ([]domain.Flight, error)
}

func (uc *FlightUsecase) ListFlightsPage(ctx context.Context, params pagination.Params) (pagination.Page[domain.Flight], error) {
	if repo, ok := uc.repo.(paginatedFlightRepository); ok {
		flights, err := repo.ListFlightsPage(ctx, params)
		if err != nil {
			if errors.Is(err, repository.ErrOperationNotPermitted) {
				return pagination.Page[domain.Flight]{}, ErrOperationNotPermitted
			}
			return pagination.Page[domain.Flight]{}, fmt.Errorf("FlightUsecase.ListFlightsPage: %w", err)
		}
		return pagination.FromFetched(flights, params, func(f domain.Flight) uuid.UUID { return f.Id }), nil
	}
	flights, err := uc.ListFlights(ctx)
	if err != nil {
		return pagination.Page[domain.Flight]{}, err
	}
	return pagination.Paginate(flights, params, func(f domain.Flight) uuid.UUID { return f.Id }), nil
}
