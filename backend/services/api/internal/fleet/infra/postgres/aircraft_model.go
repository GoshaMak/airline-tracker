package postgres

import (
	"api/internal/fleet/domain"
	"api/internal/fleet/domain/repository"
	"api/internal/fleet/infra/postgres/model"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

type aircraftModelRepository struct {
	conn *pgxpool.Pool
}

func NewAircraftModelRepository(i do.Injector) (repository.AircraftModelRepository, error) {
	return &aircraftModelRepository{
		conn: do.MustInvoke[*pgxpool.Pool](i),
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
	_, err := r.conn.Exec(ctx, query,
		am.Manufacturer.String(), am.Model.String(),
		am.Mass.String(), am.MaxAltitude.String(), am.MaxSpeed.String(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrAircraftModelAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *aircraftModelRepository) GetAircraftModelById(
	ctx context.Context,
	id uuid.UUID,
) (domain.AircraftModel, error) {
	const op = "AircraftModelRepository.GetAircraftModelById"
	query := `
	select  *
	from aircraft_models
	where id = $1
	`
	rows, _ := r.conn.Query(ctx, query, id)
	amm, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.AircraftModelModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AircraftModel{}, repository.ErrAircraftModelNotFound
		}
		return domain.AircraftModel{}, fmt.Errorf("%s: %w", op, err)
	}

	amd, err := model.AircraftModelModelToDomain(amm)
	if err != nil {
		return domain.AircraftModel{}, fmt.Errorf("%s: %w", op, err)
	}

	return amd, nil
}
