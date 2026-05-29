package mysql

import (
	"api/internal/airport/domain"
	"api/internal/airport/domain/repository"
	dbmysql "api/internal/infra/mysql"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type gateRepository struct {
	conn *sql.DB
}

func NewGateRepository(i do.Injector) (repository.GateRepository, error) {
	return &gateRepository{
		conn: do.MustInvoke[*sql.DB](i),
	}, nil
}

func (r *gateRepository) Save(
	ctx context.Context,
	g domain.Gate,
) error {
	query := `
	insert into gates(id, airport_id, number)
		values (?, ?, ?)
	`
	_, err := r.conn.ExecContext(ctx, query,
		g.Id.String(), g.AirportId.String(), g.Number.String(),
	)
	if err != nil {
		if dbmysql.IsDuplicate(err) {
			return repository.ErrGateAlreadyExists
		}
		return err
	}
	return nil
}

func (r *gateRepository) GetAirportByGateId(
	ctx context.Context,
	gid uuid.UUID,
) (domain.Airport, error) {
	const op = "GateRepository.GetAirportByGateId"
	query := `
	select a.id, a.iata_code, a.title, a.city, a.country
	from airports a
	where a.id = (select airport_id
					from gates g
					where g.id = ?)
	`
	a, err := scanAirport(r.conn.QueryRowContext(ctx, query, gid.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Airport{}, repository.ErrAirportNotFound
		}
		return domain.Airport{}, fmt.Errorf("%s: %w", op, err)
	}

	return a, nil
}

func (r *gateRepository) List(ctx context.Context) ([]domain.Gate, error) {
	const op = "GateRepository.List"
	query := `
	select id, airport_id, number
	from gates
	`
	rows, err := r.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	gates := make([]domain.Gate, 0)
	for rows.Next() {
		g, err := scanGate(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		gates = append(gates, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return gates, nil
}

func scanGate(scanner airportScanner) (domain.Gate, error) {
	var (
		id, airportId uuid.UUID
		number        string
	)
	if err := scanner.Scan(&id, &airportId, &number); err != nil {
		return domain.Gate{}, err
	}

	gateNumber, err := domain.NewGateNumber(number)
	if err != nil {
		return domain.Gate{}, err
	}

	return domain.Gate{
		Id:        id,
		AirportId: airportId,
		Number:    gateNumber,
	}, nil
}
