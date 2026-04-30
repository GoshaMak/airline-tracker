package postgres

import (
	"airline-tracker/internal/fleet/domain"
	"airline-tracker/internal/fleet/domain/repository"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

type aircraftRepository struct {
	pool *pgxpool.Pool
}

func NewAircraftRepository(i do.Injector) (repository.AircraftRepository, error) {
	return &aircraftRepository{
		pool: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

func (r *aircraftRepository) SaveAircraft(
	ctx context.Context,
	a domain.Aircraft,
) error {
	op := "AircraftRepository.SaveAircraft"
	query := `
	insert into
		aircraft(registration_number, aircraft_model_id, serial_number, mileage)
		values ($1, $2, $3, $4)
	`
	_, err := r.pool.Exec(ctx, query,
		a.RegistrationNumber.String(), a.AircraftModelID.String(),
		a.SerialNumber.String(), a.Mileage.String(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAircraftAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *aircraftRepository) Exists(
	ctx context.Context,
	a domain.Aircraft,
) (domain.Aircraft, error) {
	return domain.Aircraft{}, nil
}
