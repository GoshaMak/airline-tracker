package repository

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/flight/domain"
	"api/internal/flight/infra/postgres"
	"api/internal/flight/infra/redis"
	infraPostgres "api/internal/infra/postgres"
	userDomain "api/internal/user/domain"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

var (
	ErrFlightNotFound        = errors.New("flight not found")
	ErrFlightRouteNotFound   = errors.New("flight route not found")
	ErrOperationNotPermitted = errors.New("repository operation not permitted")
)

type FlightDB interface {
	Save(ctx context.Context, flight domain.Flight) error
	Exist(ctx context.Context, fid uuid.UUID) (domain.Flight, error)
	Update(ctx context.Context, ufi domain.UpdateFlightInfo) error
	ListFlights(ctx context.Context) ([]domain.Flight, error)
	GetFlightRoute(ctx context.Context, fid uuid.UUID) (domain.FlightRoute, error)
	ListSubscribers(ctx context.Context, fid uuid.UUID) ([]userDomain.User, error)
	GetFlightAirports(
		ctx context.Context,
		fid uuid.UUID,
	) (dep airportDomain.Airport, arr airportDomain.Airport, err error)
}

type FlightCache interface {
	SaveFlight(context.Context, domain.Flight) error
	SaveFlights(context.Context, []domain.Flight) error
	GetFlightById(context.Context, uuid.UUID) (domain.Flight, error)
	GetFlights(context.Context) ([]domain.Flight, error)
	UpdateFlight(context.Context, domain.UpdateFlightInfo) error
	FlushFlights(context.Context) error
}

type flightRepository struct {
	db    FlightDB
	cache FlightCache
}

func NewFlightRepository(i do.Injector) (FlightDB, error) {
	return &flightRepository{
		db:    do.MustInvoke[*postgres.PostgresClient](i),
		cache: do.MustInvoke[*redis.RedisClient](i),
	}, nil
}

func (r *flightRepository) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	if tx, ok := r.db.(interface {
		WithTransaction(context.Context, func(context.Context) error) error
	}); ok {
		err := tx.WithTransaction(ctx, fn)
		if err == nil {
			if flushErr := r.cache.FlushFlights(ctx); flushErr != nil {
				slog.WarnContext(ctx, "FlightRepository.WithTransaction: cache flush failed", "err", flushErr)
			}
		}
		return err
	}
	return fn(ctx)
}

func (r *flightRepository) Save(ctx context.Context, flight domain.Flight) error {
	const op = "FlightRepository.Save"
	if err := r.db.Save(ctx, flight); err != nil {
		if errors.Is(err, postgres.ErrOperationNotPermitted) {
			return ErrOperationNotPermitted
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := r.cache.SaveFlight(ctx, flight); err != nil {
		slog.WarnContext(ctx, op, "err", err)
		return nil
	}

	return nil
}

func (r *flightRepository) Exist(ctx context.Context, fid uuid.UUID) (domain.Flight, error) {
	const op = "FlightRepository.Exist"
	var f domain.Flight
	f, err := r.cache.GetFlightById(ctx, fid)
	if err != nil {
		if errors.Is(err, redis.ErrFlightNotFound) {
			slog.InfoContext(ctx, op+": flight not cached", "fid", fid)
		} else {
			slog.ErrorContext(ctx, op+": redis failed", "err", err)
			if err := r.cache.FlushFlights(ctx); err != nil {
				slog.ErrorContext(ctx, op+": redis failed while flushing", "err", err)
			}
		}
	}

	f, err = r.db.Exist(ctx, fid)
	if err != nil {
		if errors.Is(err, postgres.ErrFlightNotFound) {
			return domain.Flight{}, ErrFlightNotFound
		}
		if errors.Is(err, postgres.ErrOperationNotPermitted) {
			return domain.Flight{}, ErrOperationNotPermitted
		}
		return domain.Flight{}, fmt.Errorf("%s: %w", op, err)
	}
	return f, nil
}

func (r *flightRepository) Update(ctx context.Context, ufi domain.UpdateFlightInfo) error {
	const op = "FlightRepository.Update"
	if err := r.db.Update(ctx, ufi); err != nil {
		slog.WarnContext(ctx, op, "err", err)
		if errors.Is(err, postgres.ErrFlightNotFound) {
			return ErrFlightNotFound
		}
		if errors.Is(err, postgres.ErrOperationNotPermitted) {
			return ErrOperationNotPermitted
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if infraPostgres.TxFromContext(ctx) != nil {
		return nil
	}
	if err := r.cache.UpdateFlight(ctx, ufi); err != nil {
		slog.WarnContext(ctx, op, "err", err)
		if err := r.cache.FlushFlights(ctx); err != nil {
			slog.ErrorContext(ctx, op+": redis failed while flushing", "err", err)
		}
	}
	return nil
}

func (r *flightRepository) UpdateIfChanged(ctx context.Context, ufi domain.UpdateFlightInfo) (bool, error) {
	const op = "FlightRepository.UpdateIfChanged"
	updater, ok := r.db.(interface {
		UpdateIfChanged(context.Context, domain.UpdateFlightInfo) (bool, error)
	})
	if !ok {
		if err := r.db.Update(ctx, ufi); err != nil {
			return false, err
		}
		return true, nil
	}
	changed, err := updater.UpdateIfChanged(ctx, ufi)
	if err != nil {
		slog.WarnContext(ctx, op, "err", err)
		if errors.Is(err, postgres.ErrFlightNotFound) {
			return false, ErrFlightNotFound
		}
		if errors.Is(err, postgres.ErrOperationNotPermitted) {
			return false, ErrOperationNotPermitted
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if changed && infraPostgres.TxFromContext(ctx) == nil {
		if cacheErr := r.cache.UpdateFlight(ctx, ufi); cacheErr != nil {
			slog.WarnContext(ctx, op, "err", cacheErr)
			_ = r.cache.FlushFlights(ctx)
		}
	}
	return changed, nil
}

func (r *flightRepository) ListFlights(ctx context.Context) ([]domain.Flight, error) {
	const op = "FlightRepository.ListFlights"
	var flights []domain.Flight
	flights, err := r.cache.GetFlights(ctx)
	if err != nil {
		slog.ErrorContext(ctx, op+": redis failed", "err", err)
		if err := r.cache.FlushFlights(ctx); err != nil {
			slog.ErrorContext(ctx, op+": redis failed while flushing", "err", err)
		}
	} else if len(flights) != 0 {
		slog.InfoContext(ctx, op+": flights cached")
		return flights, nil
	}
	slog.InfoContext(ctx, op+": flights not cached")

	flights, err = r.db.ListFlights(ctx)
	if err != nil {
		if errors.Is(err, postgres.ErrOperationNotPermitted) {
			return nil, ErrOperationNotPermitted
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := r.cache.FlushFlights(ctx); err != nil {
		slog.ErrorContext(ctx, op+": redis failed while flushing", "err", err)
		return flights, nil
	}
	if err := r.cache.SaveFlights(ctx, flights); err != nil {
		slog.ErrorContext(ctx, op+": can't cache flights", "err", err)
		return flights, nil
	}
	slog.InfoContext(ctx, op+": flights cached")

	return flights, nil
}

func (r *flightRepository) GetFlightRoute(ctx context.Context, fid uuid.UUID) (domain.FlightRoute, error) {
	const op = "FlightRepository.GetFlightRoute"
	fr, err := r.db.GetFlightRoute(ctx, fid)
	if err != nil {
		if errors.Is(err, postgres.ErrFlightRouteNotFound) {
			return domain.FlightRoute{}, ErrFlightRouteNotFound
		}
		if errors.Is(err, postgres.ErrOperationNotPermitted) {
			return domain.FlightRoute{}, ErrOperationNotPermitted
		}
		return domain.FlightRoute{}, fmt.Errorf("%s: %w", op, err)
	}
	return fr, nil
}

func (r *flightRepository) ListSubscribers(ctx context.Context, fid uuid.UUID) ([]userDomain.User, error) {
	const op = "FlightRepository.ListSubscribers"
	subs, err := r.db.ListSubscribers(ctx, fid)
	if err != nil {
		if errors.Is(err, postgres.ErrOperationNotPermitted) {
			return nil, ErrOperationNotPermitted
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return subs, nil
}

func (r *flightRepository) GetFlightAirports(
	ctx context.Context,
	fid uuid.UUID,
) (dep airportDomain.Airport, arr airportDomain.Airport, err error) {
	const op = "FlightRepository.GetFlightAirports"
	dep, arr, err = r.db.GetFlightAirports(ctx, fid)
	if err != nil {
		if errors.Is(err, postgres.ErrOperationNotPermitted) {
			return dep, arr, ErrOperationNotPermitted
		}
		return dep, arr, fmt.Errorf("%s: %w", op, err)
	}
	return dep, arr, nil
}
