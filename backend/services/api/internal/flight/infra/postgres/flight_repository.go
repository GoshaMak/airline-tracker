package postgres

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/flight/domain"
	"api/internal/flight/infra/postgres/model"
	userDomain "api/internal/user/domain"
	userModel "api/internal/user/infra/postgres/model"

	"api/internal/utils"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

type PostgresDB struct {
	conn *pgxpool.Pool
}

func NewPostgresDB(i do.Injector) (*PostgresDB, error) {
	return &PostgresDB{
		conn: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

func (p *PostgresDB) Save(ctx context.Context, f domain.Flight) error {
	const op = "PostgresDB.SaveFlight"
	var plan *string
	if f.Plan != nil {
		plan = utils.Ptr(f.Plan.String())
	}
	_, err := p.conn.Exec(ctx,
		"select add_flight($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
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

func (p *PostgresDB) Exist(ctx context.Context, fid uuid.UUID) (domain.Flight, error) {
	const op = "PostgresDB.Exist"
	query := `
	select *
	from scan_flight_info($1)
	`
	row, _ := p.conn.Query(ctx, query, fid)
	fm, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[model.FlightModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Flight{}, ErrFlightNotFound
		}
		return domain.Flight{}, fmt.Errorf("%s: %w", op, err)
	}

	fd, err := model.FlightModelToDomain(fm)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("%s: %w", op, err)
	}
	return fd, nil
}

func (p *PostgresDB) Update(
	ctx context.Context,
	ufi domain.UpdateFlightInfo,
) error {
	const op = "PostgresDB.Update"
	query := `
	update flights
	set`
	args := []any{}
	if ufi.ScheduledDeparture != nil {
		query += " scheduled_departure = $" + strconv.FormatInt(int64(len(args)+1), 10) + ",\n"
		args = append(args, ufi.ScheduledDeparture)
	}
	if ufi.ActualDeparture != nil {
		query += " actual_departure = $" + strconv.FormatInt(int64(len(args)+1), 10) + ",\n"
		args = append(args, ufi.ActualDeparture)
	}
	if ufi.ScheduledArrival != nil {
		query += " scheduled_arrival = $" + strconv.FormatInt(int64(len(args)+1), 10) + ",\n"
		args = append(args, ufi.ScheduledArrival)
	}
	if ufi.ActualArrival != nil {
		query += " actual_arrival = $" + strconv.FormatInt(int64(len(args)+1), 10) + ",\n"
		args = append(args, ufi.ActualArrival)
	}
	if ufi.Status != nil {
		query += " status = $" + strconv.FormatInt(int64(len(args)+1), 10) + ",\n"
		args = append(args, ufi.Status)
	}
	if ufi.Plan != nil {
		query += " plan = $" + strconv.FormatInt(int64(len(args)+1), 10) + ",\n"
		args = append(args, ufi.Plan)
	}

	if len(args) == 0 {
		slog.Debug(op + ": nothing to update")
		return nil
	}

	query = query[:len(query)-2] + "\n"

	query += "where id = $" + strconv.FormatInt(int64(len(args)+1), 10)
	args = append(args, ufi.FlightId)
	slog.Debug(op, "query", query, "args", args)

	_, err := p.conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *PostgresDB) ListFlights(ctx context.Context) ([]domain.Flight, error) {
	const op = "PostgresDB.ListAllFlights"
	query := `
	select *
	from scan_flights_info()
	`
	rows, _ := p.conn.Query(ctx, query)
	flightsModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.FlightModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	flights := make([]domain.Flight, len(flightsModels))
	for i, m := range flightsModels {
		f, err := model.FlightModelToDomain(m)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		flights[i] = f
	}

	return flights, nil
}

func (p *PostgresDB) GetFlightRoute(
	ctx context.Context,
	fid uuid.UUID,
) (domain.FlightRoute, error) {
	const op = "PostgresDB.GetFlightRouteId"
	query := `
	select *
	from flight_routes
	where flight_id = $1
	`
	rows, _ := p.conn.Query(ctx, query, fid)
	rm, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.FlightRouteModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FlightRoute{}, ErrFlightRouteNotFound
		}
		return domain.FlightRoute{}, fmt.Errorf("%s: %w", op, err)
	}

	rd, err := model.FlightRouteModelToDomain(rm)
	if err != nil {
		return domain.FlightRoute{}, fmt.Errorf("%s: %w", op, err)
	}

	return rd, nil
}

func (p *PostgresDB) ListSubscribers(
	ctx context.Context,
	fid uuid.UUID,
) ([]userDomain.User, error) {
	const op = "PostgresDB.ListSubscribers"
	query := `
	select u.id as id, u.email as email, u.password_hash as password_hash, u.role as role
	from subscriptions s
		join users u on s.user_id = u.id
	where s.flight_id = $1::uuid;
	`

	rows, _ := p.conn.Query(ctx, query, fid)
	userMs, err := pgx.CollectRows(rows, pgx.RowToStructByName[userModel.UserModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	users := make([]userDomain.User, len(userMs))
	for i, um := range userMs {
		u, err := userModel.UserModelToDomain(um)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		users[i] = u
	}
	return users, nil
}

func (p *PostgresDB) GetFlightAirports(
	ctx context.Context,
	fid uuid.UUID,
) (dep airportDomain.Airport, arr airportDomain.Airport, err error) {
	const op = "PostgresDB.GetFlightAirports"
	query := `
	with arps as (
		select
			ag.airport_id as dep,
			dg.airport_id as arr
		from flight_routes fr
			join gates ag on ag.id = fr.departure_gate_id
			join gates dg on dg.id = fr.arrival_gate_id
		where fr.flight_id = $1::uuid
	)
	select
		da.id        as departure_airport_id,
		da.iata_code as departure_airport_iata_code,
		da.title     as departure_airport_title,
		dc.name      as departure_airport_city,
		dcntr.code   as departure_airport_country,

		aa.id        as arrival_airport_id,
		aa.iata_code as arrival_airport_iata_code,
		aa.title     as arrival_airport_title,
		ac.name      as arrival_airport_city,
		acntr.code   as arrival_airport_country
	from arps
		join airports da on da.id = arps.dep
		join cities dc on dc.id = da.city_id
		join countries dcntr on dcntr.id = dc.country_id

		join airports aa on aa.id = arps.arr
		join cities ac on ac.id = aa.city_id
		join countries acntr on acntr.id = ac.country_id
	`
	rows, _ := p.conn.Query(ctx, query, fid)
	fam, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.FlightAirportsModel])
	slog.Debug(op, "fam", fam)
	if err != nil {
		return dep, arr, fmt.Errorf("%s: %w", op, err)
	}
	dep, arr, err = model.FlightAirportsModelToDomain(fam)
	if err != nil {
		return dep, arr, fmt.Errorf("%s: %w", op, err)
	}
	return dep, arr, nil
}
