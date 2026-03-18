package postgres

import (
	"airline-tracker/internal/airport/domain"
	"airline-tracker/internal/airport/domain/repository"
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/do/v2"
)

type gateRepository struct {
	conn *pgx.Conn
}

func NewGateRepository(i do.Injector) (repository.GateRepository, error) {
	return &gateRepository{
		conn: do.MustInvoke[*pgx.Conn](i),
	}, nil
}

func (r *gateRepository) Save(
	ctx context.Context,
	g *domain.Gate,
) error {
	_, err := r.conn.Exec(ctx,
		"insert into"+
			" gates(airport_id, number)"+
			" values ($1, $2)",
		&g.AirportID, &g.Number,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == string(RecordAlreadyExistsErrCode) {
			slog.Error("Gate already exists", "gate", *g)
			return ErrRecordAlreadyExists
		}
		slog.Error("Can't insert new gate", "error", err, "gate", *g)
		return ErrInsertFailure
	}
	return nil
}

func (r *gateRepository) Exists(
	ctx context.Context,
	a *domain.Gate,
) (*domain.Gate, error) {
	return nil, nil
}
