package postgres

import (
	"airline-tracker/internal/flight/domain"
	"airline-tracker/internal/flight/domain/repository"
	"context"
	"log/slog"
	"time"

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
	rows, err := r.conn.Query(ctx,
		"select id,"+
			" aircraft_id,"+
			" scheduled_departure,"+
			" scheduled_arrival,"+
			" actual_departure,"+
			" actual_arrival,"+
			" status,"+
			" plan,"+
			" departure_gate_id,"+
			" arrival_gate_id,"+
			" departure_airport_id,"+
			" arrival_airport_id"+
			" from scan_flights_info()")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flights []domain.Flight
	for rows.Next() {
		var fl domain.Flight
		actualDeparture := &time.Time{}
		actualArrival := &time.Time{}
		slog.Debug("pre scan")
		if err := rows.Scan(
			&fl.ID, &fl.AircraftID,
			&fl.ScheduledDeparture, &fl.ScheduledArrival,
			&actualDeparture, &actualArrival,
			&fl.Status, &fl.Plan,
			&fl.DepartureGateID, &fl.ArrivalGateID,
			&fl.DepartureAirportID, &fl.ArrivalAirportID); err != nil {
			return nil, err
		}
		if actualDeparture != nil {
			fl.ActualDeparture = *actualDeparture
		}
		if actualArrival != nil {
			fl.ActualArrival = *actualArrival
		}
		slog.Debug("after scan")
		flights = append(flights, fl)
	}

	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return flights, nil
}
