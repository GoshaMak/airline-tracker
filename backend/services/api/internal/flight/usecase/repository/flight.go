package repository

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/flight/domain"
	"api/internal/flight/domain/repository"
	"api/internal/flight/infra/postgres"
	"api/internal/flight/infra/redis"
	userDomain "api/internal/user/domain"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type flightRepository struct {
	db *postgres.PostgresDB
	rd *redis.RedisDB
}

func NewFlightRepository(i do.Injector) (repository.FlightRepository, error) {
	return &flightRepository{
		db: do.MustInvoke[*postgres.PostgresDB](i),
		rd: do.MustInvoke[*redis.RedisDB](i),
	}, nil
}
func (r *flightRepository) Save(ctx context.Context, flight domain.Flight) error {
	const op = "FlightRepository.Save"
	if err := r.db.Save(ctx, flight); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := r.rd.SaveFlight(context.Background(), flight); err != nil {
		slog.Warn(op, "err", err)
		return nil
	}

	return nil
}

func (r *flightRepository) Exist(ctx context.Context, fid uuid.UUID) (domain.Flight, error) {
	const op = "FlightRepository.Exist"
	var f domain.Flight
	f, err := r.rd.GetFlightById(ctx, fid)
	if err != nil {
		if errors.Is(err, redis.ErrFlightNotFound) {
			slog.Info(op+": flight not cached", "fid", fid)
		} else {
			slog.Error(op+": redis failed", "err", err)
			if err := r.rd.FlushFlights(ctx); err != nil {
				slog.Error(op+": redis failed while flushing", "err", err)
			}
		}
	}

	f, err = r.db.Exist(ctx, fid)
	if err != nil {
		if errors.Is(err, postgres.ErrFlightNotFound) {
			return domain.Flight{}, repository.ErrFlightNotFound
		}
		return domain.Flight{}, fmt.Errorf("%s: %w", op, err)
	}
	return f, nil
}

func (r *flightRepository) Update(ctx context.Context, ufi domain.UpdateFlightInfo) error {
	const op = "FlightRepository.Update"
	if err := r.db.Update(ctx, ufi); err != nil {
		slog.Warn(op, "err", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := r.rd.UpdateFlight(ctx, ufi); err != nil {
		slog.Warn(op, "err", err)
		if err := r.rd.FlushFlights(ctx); err != nil {
			slog.Error(op+": redis failed while flushing", "err", err)
		}
	}
	return nil
}

func (r *flightRepository) ListFlights(ctx context.Context) ([]domain.Flight, error) {
	const op = "FlightRepository.ListFlights"
	var flights []domain.Flight
	flights, err := r.rd.GetFlights(context.Background())
	if err != nil {
		slog.Error(op+": redis failed", "err", err)
		if err := r.rd.FlushFlights(ctx); err != nil {
			slog.Error(op+": redis failed while flushing", "err", err)
		}
	} else if len(flights) != 0 {
		slog.Info(op + ": flights cached")
		return flights, nil
	}
	slog.Info(op + ": flights not cached")

	flights, err = r.db.ListFlights(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := r.rd.FlushFlights(ctx); err != nil {
		slog.Error(op+": redis failed while flushing", "err", err)
		return flights, nil
	}
	if err := r.rd.SaveFlights(context.Background(), flights); err != nil {
		slog.Error(op+": can't cache flights", "err", err)
		return flights, nil
	}
	slog.Info(op + ": flights cached")

	return flights, nil
}

func (r *flightRepository) GetFlightRoute(ctx context.Context, fid uuid.UUID) (domain.FlightRoute, error) {
	const op = "FlightRepository.GetFlightRoute"
	fr, err := r.db.GetFlightRoute(ctx, fid)
	if err != nil {
		if errors.Is(err, postgres.ErrFlightRouteNotFound) {
			return domain.FlightRoute{}, repository.ErrFlightRouteNotFound
		}
		return domain.FlightRoute{}, fmt.Errorf("%s: %w", op, err)
	}
	return fr, nil
}

func (r *flightRepository) ListSubscribers(ctx context.Context, fid uuid.UUID) ([]userDomain.User, error) {
	const op = "FlightRepository.ListSubscribers"
	subs, err := r.db.ListSubscribers(ctx, fid)
	if err != nil {
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
		return dep, arr, fmt.Errorf("%s: %w", op, err)
	}
	return dep, arr, nil
}
