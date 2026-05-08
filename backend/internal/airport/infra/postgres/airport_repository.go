package postgres

import (
	"airline-tracker/internal/airport/domain"
	"airline-tracker/internal/airport/domain/repository"
	"airline-tracker/internal/airport/infra/postgres/model"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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
	a domain.Airport,
) error {
	query := `
	insert into airports(id, iata_code, title, city, country)
		values ($1, $2, $3, $4, $5)
	`

	_, err := r.conn.Exec(ctx, query,
		a.ID, a.IATACode.String(), a.Title.String(), a.City.String(), a.Country.String(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrAirportAlreadyExists
		}
		return err
	}
	return nil
}

func (r *airportRepository) ListAirports(ctx context.Context) ([]domain.Airport, error) {
	op := "AirportRepository.ListAirports"
	query := `	
	select *
	from airports
	`
	rows, _ := r.conn.Query(ctx, query)
	airportsModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.AirportModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	airports := make([]domain.Airport, len(airportsModels))
	for i, am := range airportsModels {
		a, err := model.AirportModelToDomain(am)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		airports[i] = a
	}
	return airports, nil
}
