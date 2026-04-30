package postgres

import (
	"airline-tracker/internal/airport/domain"
	"airline-tracker/internal/airport/domain/repository"
	"context"
	"errors"

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
	g *domain.Gate,
) error {
	query := `
	insert into gates(id, airport_id, number)
		values ($1, $2, $3)
	`
	_, err := r.conn.Exec(ctx, query,
		g.ID, g.AirportID, g.Number.String(),
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

func (r *gateRepository) Exists(
	ctx context.Context,
	a *domain.Gate,
) (*domain.Gate, error) {
	return nil, nil
}
