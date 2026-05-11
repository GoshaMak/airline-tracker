package postgres

import (
	"airline-tracker/internal/flight/domain"
	"airline-tracker/internal/flight/domain/repository"
	"airline-tracker/internal/flight/infra/postgres/model"
	"airline-tracker/internal/utils"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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

func (r *flightRepository) Save(ctx context.Context, f domain.Flight) error {
	op := "FlightRepository.SaveFlight"
	var plan *string
	if f.Plan != nil {
		plan = utils.Ptr(f.Plan.String())
	}
	_, err := r.conn.Exec(ctx,
		"call add_flight($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		f.Id,
		f.ScheduledDeparture,
		f.ScheduledArrival,
		f.Status.String(),
		plan,
		f.AircraftId,
		f.DepartureAirportId,
		f.ArrivalAirportId,
		f.DepartureGateId,
		f.ArrivalGateId,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *flightRepository) Exist(ctx context.Context, fid uuid.UUID) (domain.Flight, error) {
	// op := "FlightRepository.Exist"
	// query := `
	// select * from
	// `
	return domain.Flight{}, nil
}

func (r *flightRepository) UpdateById(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (r *flightRepository) ListFlights(ctx context.Context) ([]domain.Flight, error) {
	op := "FlightRepository.ListAllFlights"
	query := `
	select *
	from scan_flights_info()
	`
	rows, _ := r.conn.Query(ctx, query)
	flightsModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.FlightModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	flights := make([]domain.Flight, len(flightsModels))
	for i, m := range flightsModels {
		f, err := model.FlightModelToDomain(m)
		if err != nil {
			return nil, err
		}
		flights[i] = f
	}

	return flights, nil
}

func (r *flightRepository) GetFlightRoute(
	ctx context.Context,
	fid uuid.UUID,
) (domain.FlightRoute, error) {
	op := "FlightRepository.GetFlightRouteId"
	query := `
	select *
	from flight_routes
	where flight_id = $1
	`
	rows, _ := r.conn.Query(ctx, query, fid)
	rm, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.FlightRouteModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FlightRoute{}, repository.ErrFlightRouteNotFound
		}
		return domain.FlightRoute{}, fmt.Errorf("%s: %w", op, err)
	}

	rd, err := model.FlightRouteModelToDomain(rm)
	if err != nil {
		return domain.FlightRoute{}, fmt.Errorf("%s: %w", op, err)
	}

	return rd, nil
}
