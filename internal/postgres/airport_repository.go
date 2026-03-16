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

type airportRepository struct {
	conn *pgx.Conn
}

func NewAirportRepository(i do.Injector) (repository.AirportRepository, error) {
	return &airportRepository{
		conn: do.MustInvoke[*pgx.Conn](i),
	}, nil
}

func (r *airportRepository) Save(
	ctx context.Context,
	a *domain.Airport,
) error {
	_, err := r.conn.Exec(ctx,
		"insert into"+
			" airports(iata_code, title, city, country)"+
			" values ($1, $2, $3, $4)",
		&a.IATACode, &a.Title, &a.City, &a.Country,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == string(RecordAlreadyExistsErrCode) {
			slog.Error("Airport already exists", "airport", *a)
			return ErrRecordAlreadyExists
		}
		slog.Error("Can't insert new airport", "error", err, "airport", *a)
		return ErrInsertFailure
	}
	return nil
}

func (r *airportRepository) Exists(
	ctx context.Context,
	a *domain.Airport,
) (*domain.Airport, error) {
	return nil, nil
}
