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

func (r *flightRepository) Save(
	ctx context.Context,
	flight *domain.Flight,
	aircraft *domain.Aircraft,
	departureAirport, arrivalAirport *domain.Airport,
	departureGate, arrivalGate *domain.Gate,
) error {
	var id uint
	err := r.conn.QueryRow(ctx,
		"select add_flight($1, $2, $3, $4, $5, $6, $7, $8)",
		flight.ScheduledDeparture,
		flight.ScheduledArrival,
		flight.Status,
		flight.FlightPlan,
		aircraft.RegistrationNumber,
		departureAirport.IATACode,
		arrivalAirport.IATACode,
		departureGate.Number,
		arrivalGate.Number,
	).Scan(&id)
	if err != nil {
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
	rows, err := r.conn.Query(ctx,
		"select id, aircraft_id,"+
			" scheduled_departure, scheduled_arrival,"+
			" actual_departure, actual_arrival, status, flight_plan"+
			" from flights")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flights []domain.Flight
	// flights, err := pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Flight])
	for rows.Next() {
		var fl domain.Flight
		if err := rows.Scan(&fl.ID, &fl.AircraftID,
			&fl.ScheduledDeparture, &fl.ScheduledArrival,
			&fl.ActualDeparture, &fl.ActualArrival,
			&fl.Status, &fl.FlightPlan); err != nil {
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
