package postgres

import (
	"airline-tracker/internal/domain"
	"airline-tracker/internal/domain/repository"
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/do/v2"
)

type aircraftModelRepository struct {
	conn *pgx.Conn
}

func NewAircraftModelRepository(i do.Injector) (repository.AircraftModelRepository, error) {
	return &aircraftModelRepository{
		conn: do.MustInvoke[*pgx.Conn](i),
	}, nil
}

func (r *aircraftModelRepository) Save(
	ctx context.Context,
	am *domain.AircraftModel,
) error {
	_, err := r.conn.Exec(ctx,
		"insert into"+
			" aircraft_models(manufacturer, model, mass, max_altitude, max_speed)"+
			" values ($1, $2, $3, $4, $5)",
		&am.Manufacturer, &am.Model, &am.Mass, &am.MaxAltitude, &am.MaxSpeed,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == string(RecordAlreadyExistsErrCode) {
			slog.Error("AircraftModel already exists", "aircraft model", *am)
			return ErrRecordAlreadyExists
		}
		slog.Error("Can't insert new aircraftModel", "error", err, "aircraft model", *am)
		return ErrInsertFailure
	}
	return nil
}

func (r *aircraftModelRepository) Exists(
	ctx context.Context,
	am *domain.AircraftModel,
) (*domain.AircraftModel, error) {
	return nil, nil
}
