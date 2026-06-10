package postgres

import (
	"api/internal/fleet/domain"
	"api/internal/fleet/domain/repository"
	"api/internal/fleet/infra/postgres/model"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

type aircraftRepository struct {
	conn *pgxpool.Pool
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
	rows, _ := r.conn.Query(ctx, query)
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
