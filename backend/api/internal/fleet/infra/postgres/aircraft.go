package postgres

import (
	"api/internal/fleet/domain"
	"api/internal/fleet/domain/repository"
	"api/internal/fleet/infra/postgres/model"
	"api/internal/pagination"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

type aircraftRepository struct {
	conn aircraftDB
}

type aircraftDB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func NewAircraftRepository(i do.Injector) (repository.AircraftRepository, error) {
	return &aircraftRepository{
		conn: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

func (r *aircraftRepository) SaveAircraft(
	ctx context.Context,
	a domain.Aircraft,
) error {
	const op = "AircraftRepository.SaveAircraft"
	query := `
	insert into
		aircraft(registration_number, aircraft_model_id, serial_number, mileage)
		values ($1, $2, $3, $4)
	`
	_, err := r.conn.Exec(ctx, query,
		a.RegistrationNumber.String(), a.AircraftModelId.String(),
		a.SerialNumber.String(), a.Mileage.String(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrAircraftAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *aircraftRepository) List(ctx context.Context) ([]domain.Aircraft, error) {
	const op = "AircraftRepository.List"
	query := `
	select *
	from aircraft
	`
	rows, err := r.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	ams, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.AircraftModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	ads := make([]domain.Aircraft, len(ams))
	for i, am := range ams {
		ad, err := model.AircraftModelToDomain(am)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		ads[i] = ad
	}
	return ads, nil
}

func (r *aircraftRepository) ListPage(
	ctx context.Context,
	params pagination.Params,
) ([]domain.Aircraft, error) {
	const op = "AircraftRepository.ListPage"
	query := `
	select *
	from aircraft
	where ($1::uuid is null or id > $1)
	order by id
	limit $2
	`
	rows, err := r.conn.Query(ctx, query, params.CursorValue(), params.FetchLimit())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.AircraftModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	aircraft := make([]domain.Aircraft, len(models))
	for i, aircraftModel := range models {
		item, err := model.AircraftModelToDomain(aircraftModel)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		aircraft[i] = item
	}
	return aircraft, nil
}
