package postgres

import (
	"airline-tracker/internal/flight/domain"
	"airline-tracker/internal/flight/domain/repository"
	"context"
	"fmt"

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

func (r *flightRepository) SaveFlight(ctx context.Context, f domain.Flight) error {
	op := "FlightRepository.SaveFlight"
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
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *flightRepository) Exists(ctx context.Context, id uint32) (domain.Flight, error) {
	return domain.Flight{}, nil
}

func (r *flightRepository) UpdateByID(ctx context.Context, id uint32) error {
	return nil
}

func (r *flightRepository) ListAllFlights(ctx context.Context) ([]domain.Flight, error) {
	op := "FlightRepository.ListAllFlights"
	query := `
	select * from scan_flights_info()
	`
	rows, _ := r.conn.Query(ctx, query)
	flights, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Flight])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return flights, nil
}
