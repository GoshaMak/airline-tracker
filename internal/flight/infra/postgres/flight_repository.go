package postgres

import (
	"airline-tracker/internal/flight/domain"
	"airline-tracker/internal/flight/domain/repository"
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

type flightRepository struct {
	conn *pgxpool.Pool
}

func NewFlightRepository(i do.Injector) (repository.FlightRepository, error) {
	return &flightRepository{
		conn: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

func (r *flightRepository) Save(ctx context.Context, f *domain.Flight) error {
	op := "FlightRepository.Save"
	_, err := r.conn.Exec(ctx,
		"call add_flight($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		f.ID,
		f.ScheduledDeparture,
		f.ScheduledArrival,
		f.Status,
		f.Plan,
		f.AircraftID,
		f.DepartureAirportID,
		f.ArrivalAirportID,
		f.DepartureGateID,
		f.ArrivalGateID,
	)
	if err != nil {
		slog.Debug(op, "error in add_flight", err)
		return err
	}
	return nil
}

func (r *flightRepository) Exists(ctx context.Context, id uint32) (*domain.Flight, error) {
	return nil, nil
}

func (r *flightRepository) UpdateByID(ctx context.Context, id uint32) error {
	return nil
}

func (r *flightRepository) ListAllFlights(ctx context.Context) ([]domain.Flight, error) {
	query := `
	select * from scan_flights_info()
	`

	rows, err := r.conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flights, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Flight])

	if err != nil {
		return nil, err
	}
	return flights, nil
}
