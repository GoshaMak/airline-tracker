package postgres

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/flight/domain"
	"api/internal/flight/infra/postgres/model"
	infraPostgres "api/internal/infra/postgres"
	"api/internal/pagination"
	userDomain "api/internal/user/domain"
	userModel "api/internal/user/infra/postgres/model"

	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/do/v2"
)

type PostgresClient struct {
	conn flightDB
}

type flightDB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	Begin(context.Context) (pgx.Tx, error)
}

func NewPostgresDB(i do.Injector) (*PostgresClient, error) {
	return &PostgresClient{
		conn: do.MustInvokeAs[flightDB](i),
	}, nil
}

func (p *PostgresClient) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return infraPostgres.WithTransaction(ctx, p.conn, fn)
}

func (p *PostgresClient) Save(ctx context.Context, f domain.Flight) error {
	const op = "PostgresClient.Save"
	var plan *string
	if f.Plan != nil {
		plan = new(f.Plan.String())
	}
	_, err := infraPostgres.Executor(ctx, p.conn).Exec(ctx,
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
		return fmt.Errorf("%s: %w", op, translateError(err))
	}
	return nil
}

func (p *PostgresClient) Exist(ctx context.Context, fid uuid.UUID) (domain.Flight, error) {
	const op = "PostgresClient.Exist"
	query := `
	select *
	from scan_flight_info($1)
	`
	row, err := infraPostgres.Executor(ctx, p.conn).Query(ctx, query, fid)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("%s: %w", op, translateError(err))
	}
	fm, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[model.FlightModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Flight{}, ErrFlightNotFound
		}
		return domain.Flight{}, fmt.Errorf("%s: %w", op, translateError(err))
	}

	fd, err := model.FlightModelToDomain(fm)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("%s: %w", op, translateError(err))
	}
	return fd, nil
}

func (p *PostgresClient) Update(
	ctx context.Context,
	ufi domain.UpdateFlightInfo,
) error {
	const op = "PostgresClient.Update"
	query := `
	update flights
	set`

	if _, err := p.Exist(ctx, ufi.FlightId); err != nil {
		return fmt.Errorf("%s: %w", op, translateError(err))
	}

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

	_, err := infraPostgres.Executor(ctx, p.conn).Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, translateError(err))
	}

	return nil
}

type currentFlight struct {
	Id                 uuid.UUID  `db:"id"`
	ScheduledDeparture time.Time  `db:"scheduled_departure"`
	ScheduledArrival   time.Time  `db:"scheduled_arrival"`
	ActualDeparture    *time.Time `db:"actual_departure"`
	ActualArrival      *time.Time `db:"actual_arrival"`
	Status             string     `db:"status"`
	Plan               *string    `db:"plan"`
}

func (p *PostgresClient) UpdateIfChanged(ctx context.Context, ufi domain.UpdateFlightInfo) (bool, error) {
	op := "PostgresClient.UpdateIfChanged"
	query := `
	select id, scheduled_departure, scheduled_arrival, actual_departure, actual_arrival, status, plan
	from flights
	where id=$1
	for update
	`
	rows, err := infraPostgres.Executor(ctx, p.conn).Query(ctx, query, ufi.FlightId)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, translateError(err))
	}
	current, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[currentFlight])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrFlightNotFound
		}
		return false, fmt.Errorf("%s: %w", op, translateError(err))
	}
	if sameUpdate(current, ufi) {
		return false, nil
	}
	if err := p.Update(ctx, ufi); err != nil {
		return false, err
	}
	return true, nil
}

func sameUpdate(current currentFlight, update domain.UpdateFlightInfo) bool {
	if update.ScheduledDeparture != nil && !current.ScheduledDeparture.Equal(*update.ScheduledDeparture) {
		return false
	}
	if update.ScheduledArrival != nil && !current.ScheduledArrival.Equal(*update.ScheduledArrival) {
		return false
	}
	if update.ActualDeparture != nil && !sameTime(current.ActualDeparture, update.ActualDeparture) {
		return false
	}
	if update.ActualArrival != nil && !sameTime(current.ActualArrival, update.ActualArrival) {
		return false
	}
	if update.Status != nil && current.Status != *update.Status {
		return false
	}
	if update.Plan != nil && (current.Plan == nil || *current.Plan != *update.Plan) {
		return false
	}
	return true
}

func sameTime(current, update *time.Time) bool {
	if current == nil || update == nil {
		return current == nil && update == nil
	}
	return current.Equal(*update)
}

func (p *PostgresClient) ListFlights(ctx context.Context) ([]domain.Flight, error) {
	const op = "PostgresClient.ListFlights"
	query := `
	select *
	from scan_flights_info()
	`
	rows, err := infraPostgres.Executor(ctx, p.conn).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, translateError(err))
	}
	flightsModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.FlightModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, translateError(err))
	}

	flights := make([]domain.Flight, len(flightsModels))
	for i, m := range flightsModels {
		f, err := model.FlightModelToDomain(m)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, translateError(err))
		}
		flights[i] = f
	}

	return flights, nil
}

func (p *PostgresClient) ListFlightsPage(
	ctx context.Context,
	params pagination.Params,
) ([]domain.Flight, error) {
	const op = "PostgresDB.ListFlightsPage"
	query := `
	select *
	from scan_flights_info()
	where ($1::uuid is null or id > $1)
	order by id
	limit $2
	`
	rows, err := infraPostgres.Executor(ctx, p.conn).Query(ctx, query, params.CursorValue(), params.FetchLimit())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, translateError(err))
	}
	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.FlightModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, translateError(err))
	}
	flights := make([]domain.Flight, len(models))
	for i, flightModel := range models {
		flight, err := model.FlightModelToDomain(flightModel)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, translateError(err))
		}
		flights[i] = flight
	}
	return flights, nil
}

func (p *PostgresClient) GetFlightRoute(
	ctx context.Context,
	fid uuid.UUID,
) (domain.FlightRoute, error) {
	const op = "PostgresDB.GetFlightRouteId"
	query := `
	select *
	from flight_routes
	where flight_id = $1
	`
	rows, err := infraPostgres.Executor(ctx, p.conn).Query(ctx, query, fid)
	if err != nil {
		return domain.FlightRoute{}, fmt.Errorf("%s: %w", op, translateError(err))
	}
	rm, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.FlightRouteModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FlightRoute{}, ErrFlightRouteNotFound
		}
		return domain.FlightRoute{}, fmt.Errorf("%s: %w", op, translateError(err))
	}

	rd, err := model.FlightRouteModelToDomain(rm)
	if err != nil {
		return domain.FlightRoute{}, fmt.Errorf("%s: %w", op, translateError(err))
	}

	return rd, nil
}

func (p *PostgresClient) ListSubscribers(
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

	rows, err := infraPostgres.Executor(ctx, p.conn).Query(ctx, query, fid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, translateError(err))
	}
	userMs, err := pgx.CollectRows(rows, pgx.RowToStructByName[userModel.UserModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, translateError(err))
	}
	users := make([]userDomain.User, len(userMs))
	for i, um := range userMs {
		u, err := userModel.UserModelToDomain(um)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, translateError(err))
		}
		users[i] = u
	}
	return users, nil
}

func (p *PostgresClient) GetFlightAirports(
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
	rows, err := infraPostgres.Executor(ctx, p.conn).Query(ctx, query, fid)
	if err != nil {
		return dep, arr, fmt.Errorf("%s: %w", op, translateError(err))
	}
	fam, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.FlightAirportsModel])
	slog.Debug(op, "fam", fam)
	if err != nil {
		return dep, arr, fmt.Errorf("%s: %w", op, translateError(err))
	}
	dep, arr, err = model.FlightAirportsModelToDomain(fam)
	if err != nil {
		return dep, arr, fmt.Errorf("%s: %w", op, translateError(err))
	}
	return dep, arr, nil
}
