package postgres

import (
	flightDomain "airline-tracker/internal/flight/domain"
	"airline-tracker/internal/flight/domain/repository"
	"context"
	"log/slog"

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

func (r *flightRepository) Save(ctx context.Context, f *flightDomain.Flight) error {
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

func (r *flightRepository) Exists(ctx context.Context, id uint32) (*flightDomain.Flight, error) {
	return nil, nil
}

func (r *flightRepository) UpdateByID(ctx context.Context, id uint32) error {
	return nil
}

func (r *flightRepository) ListAllFlights(ctx context.Context) ([]flightDomain.Flight, error) {
	rows, err := r.conn.Query(ctx,
		"select id, aircraft_id,"+
			" scheduled_departure, scheduled_arrival,"+
			" actual_departure, actual_arrival, status, flight_plan"+
			" from flights")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flights []flightDomain.Flight
	// flights, err := pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Flight])
	for rows.Next() {
		var fl flightDomain.Flight
		if err := rows.Scan(&fl.ID, &fl.AircraftID,
			&fl.ScheduledDeparture, &fl.ScheduledArrival,
			&fl.ActualDeparture, &fl.ActualArrival,
			&fl.Status, &fl.Plan); err != nil {
			return nil, err
		}
		flights = append(flights, fl)
	}

	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return flights, nil
}
