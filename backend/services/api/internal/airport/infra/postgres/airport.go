package postgres

import (
	"api/internal/airport/domain"
	"api/internal/airport/domain/repository"
	"api/internal/airport/infra/postgres/model"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
	const op = "AirportRepository.Save"
	cityQuery := `
	select city.id
	from cities city
		join countries cntr on cntr.id = city.country_id
	where city.name = $1
		and (cntr.code = $2 or cntr.name = $2)
	`
	query := `
	insert into airports(id, iata_code, title, city_id)
		values ($1, $2, $3, $4)
	`
	var cityId uuid.UUID
	row := r.conn.QueryRow(ctx, cityQuery, a.City.String(), a.Country.String())
	if err := row.Scan(&cityId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w: city or country not found", op, err)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err := r.conn.Exec(ctx, query,
		a.ID, a.IATACode.String(), a.Title.String(), cityId,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrAirportAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *airportRepository) ListAirports(ctx context.Context) ([]domain.Airport, error) {
	const op = "AirportRepository.ListAirports"
	const query = `
	select
		a.id as id,
		a.iata_code as iata_code,
		a.title as title,
		c.name as city,
		cntr.code as country
	from airports a
		join cities c on c.id = a.city_id
		join countries cntr on cntr.id = c.country_id
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
