package repository

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/flight/domain"
	"api/internal/flight/infra/postgres"
	"api/internal/flight/infra/redis"
	infraPostgres "api/internal/infra/postgres"
	"api/internal/pagination"
	userDomain "api/internal/user/domain"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"shared/observability"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	GetFlightPage(context.Context, string) ([]domain.Flight, error)
	SaveFlightPage(context.Context, string, []domain.Flight) error
	BumpFlightPageVersion(context.Context) error
	FlightPageVersion(context.Context) (string, error)
}

type flightRepository struct {
	db    FlightDB
	cache FlightCache
}

func startFlightSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return observability.Start(ctx, "repository.flight."+name, trace.WithSpanKind(trace.SpanKindClient))
}

func endFlightSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Flight repository operation failed")
	}
	span.End()
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
		dbCtx, span := startFlightSpan(ctx, "transaction.database")
		err := tx.WithTransaction(dbCtx, fn)
		endFlightSpan(span, err)
		if err == nil {
			cacheCtx, cacheSpan := startFlightSpan(ctx, "transaction.flush_cache")
			flushErr := r.cache.FlushFlights(cacheCtx)
			endFlightSpan(cacheSpan, flushErr)
			if flushErr != nil {
				slog.WarnContext(ctx, "FlightRepository.WithTransaction: cache flush failed", "err", flushErr)
			}
		}
		return err
	}
	return fn(ctx)
}

func (r *flightRepository) Save(ctx context.Context, flight domain.Flight) error {
	const op = "FlightRepository.Save"
	dbCtx, dbSpan := startFlightSpan(ctx, "save.database")
	err := r.db.Save(dbCtx, flight)
	endFlightSpan(dbSpan, err)
	if err != nil {
		if errors.Is(err, postgres.ErrOperationNotPermitted) {
			return ErrOperationNotPermitted
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := r.cache.BumpFlightPageVersion(ctx); err != nil {
		slog.WarnContext(ctx, op+": page cache version bump failed", "err", err)
	}

	cacheCtx, cacheSpan := startFlightSpan(ctx, "save.cache")
	err = r.cache.SaveFlight(cacheCtx, flight)
	endFlightSpan(cacheSpan, err)
	if err != nil {
		slog.WarnContext(ctx, op, "err", err)
		return nil
	}

	return nil
}

func (r *flightRepository) Exist(ctx context.Context, fid uuid.UUID) (domain.Flight, error) {
	const op = "FlightRepository.Exist"
	var f domain.Flight
	cacheCtx, cacheSpan := startFlightSpan(ctx, "exist.cache_read")
	f, err := r.cache.GetFlightById(cacheCtx, fid)
	if errors.Is(err, redis.ErrFlightNotFound) {
		endFlightSpan(cacheSpan, nil)
	} else {
		endFlightSpan(cacheSpan, err)
	}
	if err != nil {
		if errors.Is(err, redis.ErrFlightNotFound) {
			slog.InfoContext(ctx, op+": flight not cached", "fid", fid)
		} else {
			slog.ErrorContext(ctx, op+": redis failed", "err", err)
			flushCtx, flushSpan := startFlightSpan(ctx, "exist.flush_cache")
			flushErr := r.cache.FlushFlights(flushCtx)
			endFlightSpan(flushSpan, flushErr)
			if flushErr != nil {
				slog.ErrorContext(ctx, op+": redis failed while flushing", "err", flushErr)
			}
		}
	}

	dbCtx, dbSpan := startFlightSpan(ctx, "exist.database")
	f, err = r.db.Exist(dbCtx, fid)
	endFlightSpan(dbSpan, err)
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
	dbCtx, dbSpan := startFlightSpan(ctx, "update.database")
	err := r.db.Update(dbCtx, ufi)
	endFlightSpan(dbSpan, err)
	if err != nil {
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
	if err := r.cache.BumpFlightPageVersion(ctx); err != nil {
		slog.WarnContext(ctx, op+": page cache version bump failed", "err", err)
	}
	cacheCtx, cacheSpan := startFlightSpan(ctx, "update.cache")
	err = r.cache.UpdateFlight(cacheCtx, ufi)
	endFlightSpan(cacheSpan, err)
	if err != nil {
		slog.WarnContext(ctx, op, "err", err)
		flushCtx, flushSpan := startFlightSpan(ctx, "update.flush_cache")
		flushErr := r.cache.FlushFlights(flushCtx)
		endFlightSpan(flushSpan, flushErr)
		if flushErr != nil {
			slog.ErrorContext(ctx, op+": redis failed while flushing", "err", flushErr)
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
		dbCtx, dbSpan := startFlightSpan(ctx, "update_if_changed.database")
		err := r.db.Update(dbCtx, ufi)
		endFlightSpan(dbSpan, err)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	dbCtx, dbSpan := startFlightSpan(ctx, "update_if_changed.database")
	changed, err := updater.UpdateIfChanged(dbCtx, ufi)
	endFlightSpan(dbSpan, err)
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
		if err := r.cache.BumpFlightPageVersion(ctx); err != nil {
			slog.WarnContext(ctx, op+": page cache version bump failed", "err", err)
		}
		cacheCtx, cacheSpan := startFlightSpan(ctx, "update_if_changed.cache")
		cacheErr := r.cache.UpdateFlight(cacheCtx, ufi)
		endFlightSpan(cacheSpan, cacheErr)
		if cacheErr != nil {
			slog.WarnContext(ctx, op, "err", cacheErr)
			flushCtx, flushSpan := startFlightSpan(ctx, "update_if_changed.flush_cache")
			flushErr := r.cache.FlushFlights(flushCtx)
			endFlightSpan(flushSpan, flushErr)
		}
	}
	return changed, nil
}

func (r *flightRepository) ListFlights(ctx context.Context) ([]domain.Flight, error) {
	const op = "FlightRepository.ListFlights"
	var flights []domain.Flight
	cacheCtx, cacheSpan := startFlightSpan(ctx, "list.cache_read")
	flights, err := r.cache.GetFlights(cacheCtx)
	endFlightSpan(cacheSpan, err)
	if err != nil {
		slog.ErrorContext(ctx, op+": redis failed", "err", err)
		flushCtx, flushSpan := startFlightSpan(ctx, "list.flush_cache_after_read_error")
		flushErr := r.cache.FlushFlights(flushCtx)
		endFlightSpan(flushSpan, flushErr)
		if flushErr != nil {
			slog.ErrorContext(ctx, op+": redis failed while flushing", "err", flushErr)
		}
	} else if len(flights) != 0 {
		slog.InfoContext(ctx, op+": flights cached")
		return flights, nil
	}
	slog.DebugContext(ctx, op+": flights not cached")

	dbCtx, dbSpan := startFlightSpan(ctx, "list.database")
	flights, err = r.db.ListFlights(dbCtx)
	endFlightSpan(dbSpan, err)
	if err != nil {
		if errors.Is(err, postgres.ErrOperationNotPermitted) {
			return nil, ErrOperationNotPermitted
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	flushCtx, flushSpan := startFlightSpan(ctx, "list.flush_cache")
	flushErr := r.cache.FlushFlights(flushCtx)
	endFlightSpan(flushSpan, flushErr)
	if flushErr != nil {
		slog.ErrorContext(ctx, op+": redis failed while flushing", "err", flushErr)
		return flights, nil
	}
	cacheSaveCtx, cacheSaveSpan := startFlightSpan(ctx, "list.cache_write")
	cacheSaveErr := r.cache.SaveFlights(cacheSaveCtx, flights)
	endFlightSpan(cacheSaveSpan, cacheSaveErr)
	if cacheSaveErr != nil {
		slog.ErrorContext(ctx, op+": can't cache flights", "err", cacheSaveErr)
		return flights, nil
	}
	slog.InfoContext(ctx, op+": flights cached")

	return flights, nil
}

func (r *flightRepository) ListFlightsPage(
	ctx context.Context,
	params pagination.Params,
) ([]domain.Flight, error) {
	version, err := r.cache.FlightPageVersion(ctx)
	if err != nil {
		return r.listFlightsPageDB(ctx, params)
	}
	cursor := "first"
	if params.Cursor != nil {
		cursor = params.Cursor.String()
	}
	key := fmt.Sprintf("flight:pages:%s:%s:%d", version, cursor, params.Limit)
	if flights, err := r.cache.GetFlightPage(ctx, key); err == nil {
		return flights, nil
	}
	flights, err := r.listFlightsPageDB(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := r.cache.SaveFlightPage(ctx, key, flights); err != nil {
		slog.WarnContext(ctx, "FlightRepository.ListFlightsPage: cache page write failed", "err", err)
	}
	return flights, nil
}

func (r *flightRepository) listFlightsPageDB(ctx context.Context, params pagination.Params) ([]domain.Flight, error) {
	db, ok := r.db.(interface {
		ListFlightsPage(context.Context, pagination.Params) ([]domain.Flight, error)
	})
	if !ok {
		return nil, errors.New("flight database does not support pagination")
	}
	return db.ListFlightsPage(ctx, params)
}

func (r *flightRepository) GetFlightRoute(ctx context.Context, fid uuid.UUID) (domain.FlightRoute, error) {
	const op = "FlightRepository.GetFlightRoute"
	dbCtx, span := startFlightSpan(ctx, "get_route.database")
	fr, err := r.db.GetFlightRoute(dbCtx, fid)
	endFlightSpan(span, err)
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
	dbCtx, span := startFlightSpan(ctx, "list_subscribers.database")
	subs, err := r.db.ListSubscribers(dbCtx, fid)
	endFlightSpan(span, err)
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
	dbCtx, span := startFlightSpan(ctx, "get_airports.database")
	dep, arr, err = r.db.GetFlightAirports(dbCtx, fid)
	endFlightSpan(span, err)
	if err != nil {
		if errors.Is(err, postgres.ErrOperationNotPermitted) {
			return dep, arr, ErrOperationNotPermitted
		}
		return dep, arr, fmt.Errorf("%s: %w", op, err)
	}
	return dep, arr, nil
}
