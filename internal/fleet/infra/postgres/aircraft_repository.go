package postgres

import (
	"airline-tracker/internal/fleet/domain"
	"airline-tracker/internal/fleet/domain/repository"
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/do/v2"
)

type aircraftRepository struct {
	conn *pgx.Conn
}

func NewAircraftRepository(i do.Injector) (repository.AircraftRepository, error) {
	return &aircraftRepository{
		conn: do.MustInvoke[*pgx.Conn](i),
	}, nil
}

func (r *aircraftRepository) Save(
	ctx context.Context,
	a *domain.Aircraft,
) error {
	_, err := r.conn.Exec(ctx,
		"insert into"+
			" aircraft(registration_number, aircraft_model_id, serial_number, mileage)"+
			" values ($1, $2, $3, $4)",
		&a.RegistrationNumber, &a.AircraftModelID, &a.SerialNumber, &a.Mileage,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == string(RecordAlreadyExistsErrCode) {
			slog.Error("Aircraft already exists", "aircraft", *a)
			return ErrRecordAlreadyExists
		}
		slog.Error("Can't insert new aircraft", "error", err, "aircraft", *a)
		return ErrInsertFailure
	}
	return nil
}

func (r *aircraftRepository) Exists(
	ctx context.Context,
	a *domain.Aircraft,
) (*domain.Aircraft, error) {
	return nil, nil
}
