package postgres

import (
	"airline-tracker/internal/airport/domain"
	"airline-tracker/internal/airport/domain/repository"
	"airline-tracker/internal/airport/infra"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

type airportRepository struct {
	conn *pgxpool.Pool
}

func NewAirportRepository(i do.Injector) (repository.AirportRepository, error) {
	return &airportRepository{
		conn: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

func (r *airportRepository) Save(
	ctx context.Context,
	a *domain.Airport,
) error {
	query := `
	insert into airports(id, iata_code, title, city, country)
		values ($1, $2, $3, $4, $5)
	`

	_, err := r.conn.Exec(ctx, query,
		a.ID, a.IATACode, a.Title, a.City, a.Country,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return infra.ErrAirportAlreadyExists
		}
		return err
	}
	return nil
}

func (r *airportRepository) Exists(
	ctx context.Context,
	a *domain.Airport,
) (*domain.Airport, error) {
	return nil, nil
}
