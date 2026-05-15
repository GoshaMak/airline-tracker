package postgres

import (
	"api/internal/fleet/domain"
	"api/internal/fleet/domain/repository"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

type aircraftModelRepository struct {
	pool *pgxpool.Pool
}

func NewAircraftModelRepository(i do.Injector) (repository.AircraftModelRepository, error) {
	return &aircraftModelRepository{
		pool: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

func (r *aircraftModelRepository) SaveAircraftModel(
	ctx context.Context,
	am domain.AircraftModel,
) error {
	const op = "AircraftModelRepository.SaveAircraftModel"
	query := `
	insert into
		aircraft_models(manufacturer, model, mass, max_altitude, max_speed)
		values ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query,
		am.Manufacturer.String(), am.Model.String(),
		am.Mass.String(), am.MaxAltitude.String(), am.MaxSpeed.String(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAircraftModelAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *aircraftModelRepository) Exists(
	ctx context.Context,
	am domain.AircraftModel,
) (domain.AircraftModel, error) {
	panic("not implemented")
}
