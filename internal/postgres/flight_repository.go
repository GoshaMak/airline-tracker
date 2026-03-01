package postgres

import (
	"airline-tracker/internal/domain"
	"airline-tracker/internal/domain/repository"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/samber/do/v2"
)

type flightRepository struct {
	conn *pgx.Conn
}

func NewFlightRepository(i do.Injector) (repository.FlightRepository, error) {
	return &flightRepository{
		conn: do.MustInvoke[*pgx.Conn](i),
	}, nil
}

func (r *flightRepository) Save(ctx context.Context, u *domain.Flight) error

func (r *flightRepository) Exists(ctx context.Context, id uint32) (*domain.Flight, error)

func (r *flightRepository) UpdateById(ctx context.Context, id uint32) error
