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

type gateRepository struct {
	conn *pgxpool.Pool
}

func NewGateRepository(i do.Injector) (repository.GateRepository, error) {
	return &gateRepository{
		conn: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

func (r *gateRepository) Save(
	ctx context.Context,
	g domain.Gate,
) error {
	query := `
	insert into gates(id, airport_id, number)
		values ($1, $2, $3)
	`
	_, err := r.conn.Exec(ctx, query,
		g.Id, g.AirportId, g.Number.String(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
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
	select *
	from airports a
	where a.id = (select airport_id
					from gates g
					where g.id = $1)
	`
	rows, _ := r.conn.Query(ctx, query, gid)
	am, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.AirportModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Airport{}, repository.ErrAirportNotFound
		}
		return domain.Airport{}, fmt.Errorf("%s: %w", op, err)
	}

	ad, err := model.AirportModelToDomain(am)
	if err != nil {
		return domain.Airport{}, fmt.Errorf("%s: %w", op, err)
	}

	return ad, nil
}

func (r *gateRepository) List(ctx context.Context) ([]domain.Gate, error) {
	const op = "GateRepository.List"
	query := `
	select *
	from gates
	`
	rows, _ := r.conn.Query(ctx, query)
	gms, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.GateModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	gsd := make([]domain.Gate, len(gms))
	for i, gm := range gms {
		gd, err := model.GateModelToDomain(gm)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		gsd[i] = gd
	}
	return gsd, nil
}
