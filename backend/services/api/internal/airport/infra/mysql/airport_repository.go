package mysql

import (
	"api/internal/airport/domain"
	"api/internal/airport/domain/repository"
	dbmysql "api/internal/infra/mysql"
	"context"
	"database/sql"
	"fmt"
	"shared/common"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type airportRepository struct {
	conn *sql.DB
}

func NewAirportRepository(i do.Injector) (repository.AirportRepository, error) {
	return &airportRepository{
		conn: do.MustInvoke[*sql.DB](i),
	}, nil
}

func (r *airportRepository) Save(
	ctx context.Context,
	a domain.Airport,
) error {
	const op = "AirportRepository.Save"
	query := `
	insert into airports(id, iata_code, title, city, country)
		values (?, ?, ?, ?, ?)
	`

	_, err := r.conn.ExecContext(ctx, query,
		a.ID.String(), a.IATACode.String(), a.Title.String(), a.City.String(), a.Country.String(),
	)
	if err != nil {
		if dbmysql.IsDuplicate(err) {
			return repository.ErrAirportAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *airportRepository) ListAirports(ctx context.Context) ([]domain.Airport, error) {
	const op = "AirportRepository.ListAirports"
	query := `
	select id, iata_code, title, city, country
	from airports
	`
	rows, err := r.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	airports := make([]domain.Airport, 0)
	for rows.Next() {
		a, err := scanAirport(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		airports = append(airports, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return airports, nil
}

type airportScanner interface {
	Scan(dest ...any) error
}

func scanAirport(scanner airportScanner) (domain.Airport, error) {
	var (
		id                    uuid.UUID
		iata, title, city, cn string
	)
	if err := scanner.Scan(&id, &iata, &title, &city, &cn); err != nil {
		return domain.Airport{}, err
	}

	iataCode, err := domain.NewIATACode(iata)
	if err != nil {
		return domain.Airport{}, err
	}
	titleValue, err := domain.NewTitle(title)
	if err != nil {
		return domain.Airport{}, err
	}
	cityValue, err := common.NewCity(city)
	if err != nil {
		return domain.Airport{}, err
	}
	countryValue, err := common.NewCountry(cn)
	if err != nil {
		return domain.Airport{}, err
	}

	return domain.Airport{
		ID:       id,
		IATACode: iataCode,
		Title:    titleValue,
		City:     cityValue,
		Country:  countryValue,
	}, nil
}
