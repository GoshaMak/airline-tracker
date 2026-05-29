package mysql

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/flight/domain"
	userDomain "api/internal/user/domain"
	"api/internal/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"shared/common"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type MySQLDB struct {
	conn *sql.DB
}

func NewMySQLDB(i do.Injector) (*MySQLDB, error) {
	return &MySQLDB{
		conn: do.MustInvoke[*sql.DB](i),
	}, nil
}

func (p *MySQLDB) Save(ctx context.Context, f domain.Flight) error {
	const op = "MySQLDB.SaveFlight"
	var plan *string
	if f.Plan != nil {
		plan = utils.Ptr(f.Plan.String())
	}
	_, err := p.conn.ExecContext(ctx,
		"call add_flight(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		f.Id.String(),
		f.ScheduledDeparture,
		f.ScheduledArrival,
		f.Status.String(),
		plan,
		f.AircraftId.String(),
		f.DepartureAirportId.String(),
		f.ArrivalAirportId.String(),
		f.DepartureGateId.String(),
		f.ArrivalGateId.String(),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (p *MySQLDB) Exist(ctx context.Context, fid uuid.UUID) (domain.Flight, error) {
	const op = "MySQLDB.Exist"
	query := `
	call scan_flight_info(?)
	`
	f, err := scanFlight(p.conn.QueryRowContext(ctx, query, fid.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Flight{}, ErrFlightNotFound
		}
		return domain.Flight{}, fmt.Errorf("%s: %w", op, err)
	}

	return f, nil
}

func (p *MySQLDB) Update(
	ctx context.Context,
	ufi domain.UpdateFlightInfo,
) error {
	const op = "MySQLDB.Update"
	query := `
	update flights
	set`
	args := []any{}
	if ufi.ScheduledDeparture != nil {
		query += " scheduled_departure = ?,\n"
		args = append(args, ufi.ScheduledDeparture)
	}
	if ufi.ActualDeparture != nil {
		query += " actual_departure = ?,\n"
		args = append(args, ufi.ActualDeparture)
	}
	if ufi.ScheduledArrival != nil {
		query += " scheduled_arrival = ?,\n"
		args = append(args, ufi.ScheduledArrival)
	}
	if ufi.ActualArrival != nil {
		query += " actual_arrival = ?,\n"
		args = append(args, ufi.ActualArrival)
	}
	if ufi.Status != nil {
		query += " status = ?,\n"
		args = append(args, ufi.Status)
	}
	if ufi.Plan != nil {
		query += " plan = ?,\n"
		args = append(args, ufi.Plan)
	}

	if len(args) == 0 {
		slog.Debug(op + ": nothing to update")
		return nil
	}

	query = query[:len(query)-2] + "\n"

	query += "where id = ?"
	args = append(args, ufi.FlightId.String())
	slog.Debug(op, "query", query, "args", args)

	_, err := p.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *MySQLDB) ListFlights(ctx context.Context) ([]domain.Flight, error) {
	const op = "MySQLDB.ListAllFlights"
	query := `
	call scan_flights_info()
	`
	rows, err := p.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	flights := make([]domain.Flight, 0)
	for rows.Next() {
		f, err := scanFlight(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		flights = append(flights, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return flights, nil
}

func (p *MySQLDB) GetFlightRoute(
	ctx context.Context,
	fid uuid.UUID,
) (domain.FlightRoute, error) {
	const op = "MySQLDB.GetFlightRouteId"
	query := `
	select id, flight_id, departure_gate_id, arrival_gate_id
	from flight_routes
	where flight_id = ?
	`
	route, err := scanFlightRoute(p.conn.QueryRowContext(ctx, query, fid.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.FlightRoute{}, ErrFlightRouteNotFound
		}
		return domain.FlightRoute{}, fmt.Errorf("%s: %w", op, err)
	}

	return route, nil
}

func (p *MySQLDB) ListSubscribers(
	ctx context.Context,
	fid uuid.UUID,
) ([]userDomain.User, error) {
	const op = "MySQLDB.ListSubscribers"
	query := `
	select u.id, u.email, u.password_hash, u.role
	from subscriptions s
		join users u on s.user_id = u.id
	where s.flight_id = ?
	`

	rows, err := p.conn.QueryContext(ctx, query, fid.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	users := make([]userDomain.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return users, nil
}

func (p *MySQLDB) GetFlightAirports(
	ctx context.Context,
	fid uuid.UUID,
) (dep airportDomain.Airport, arr airportDomain.Airport, err error) {
	const op = "MySQLDB.GetFlightAirports"
	query := `
	with arps as (select ag.airport_id as dep, dg.airport_id as arr
				  from flight_routes fr
						   join gates ag on fr.departure_gate_id = ag.id
						   join gates dg on fr.arrival_gate_id = dg.id
					  and fr.flight_id = ?)
	select
		da.id,
		da.iata_code,
		da.title,
		da.city,
		da.country,
		aa.id,
		aa.iata_code,
		aa.title,
		aa.city,
		aa.country
	from arps
			 join airports da on arps.dep = da.id
			 join airports aa on arps.arr = aa.id
	`
	dep, arr, err = scanFlightAirports(p.conn.QueryRowContext(ctx, query, fid.String()))
	if err != nil {
		return dep, arr, fmt.Errorf("%s: %w", op, err)
	}
	return dep, arr, nil
}

type flightScanner interface {
	Scan(dest ...any) error
}

func scanFlight(scanner flightScanner) (domain.Flight, error) {
	var (
		id, aircraftId, depGateId, arrGateId uuid.UUID
		depAirportId, arrAirportId           uuid.UUID
		scheduledDeparture, scheduledArrival time.Time
		actualDeparture, actualArrival       sql.NullTime
		status                               string
		plan                                 sql.NullString
	)
	if err := scanner.Scan(
		&id,
		&aircraftId,
		&scheduledDeparture,
		&scheduledArrival,
		&actualDeparture,
		&actualArrival,
		&status,
		&plan,
		&depGateId,
		&arrGateId,
		&depAirportId,
		&arrAirportId,
	); err != nil {
		return domain.Flight{}, err
	}

	st, err := domain.NewFlightStatus(status)
	if err != nil {
		return domain.Flight{}, err
	}
	var p *domain.FlightPlan
	if plan.Valid {
		pv, err := domain.NewFlightPlan(plan.String)
		if err != nil {
			return domain.Flight{}, err
		}
		p = &pv
	}
	var actDep *time.Time
	if actualDeparture.Valid {
		actDep = &actualDeparture.Time
	}
	var actArr *time.Time
	if actualArrival.Valid {
		actArr = &actualArrival.Time
	}

	return domain.Flight{
		Id:                 id,
		AircraftId:         aircraftId,
		ScheduledDeparture: scheduledDeparture,
		ScheduledArrival:   scheduledArrival,
		ActualDeparture:    actDep,
		ActualArrival:      actArr,
		Status:             st,
		Plan:               p,
		DepartureAirportId: depAirportId,
		ArrivalAirportId:   arrAirportId,
		DepartureGateId:    depGateId,
		ArrivalGateId:      arrGateId,
	}, nil
}

func scanFlightRoute(scanner flightScanner) (domain.FlightRoute, error) {
	var id, flightId, depGateId, arrGateId uuid.UUID
	if err := scanner.Scan(&id, &flightId, &depGateId, &arrGateId); err != nil {
		return domain.FlightRoute{}, err
	}
	return domain.FlightRoute{
		Id:              id,
		FlightId:        flightId,
		DepartureGateId: depGateId,
		ArrivalGateId:   arrGateId,
	}, nil
}

func scanUser(scanner flightScanner) (userDomain.User, error) {
	var (
		id               uuid.UUID
		email, pswd, rol string
	)
	if err := scanner.Scan(&id, &email, &pswd, &rol); err != nil {
		return userDomain.User{}, err
	}

	mail, err := common.NewEmail(email)
	if err != nil {
		return userDomain.User{}, err
	}
	role, err := userDomain.NewRole(rol)
	if err != nil {
		return userDomain.User{}, err
	}

	return userDomain.User{
		Id:           id,
		Email:        mail,
		PasswordHash: userDomain.PasswordHashed(pswd),
		Role:         role,
	}, nil
}

func scanFlightAirports(scanner flightScanner) (airportDomain.Airport, airportDomain.Airport, error) {
	var (
		depID, arrID                      uuid.UUID
		depIata, depTitle, depCity, depCn string
		arrIata, arrTitle, arrCity, arrCn string
	)
	if err := scanner.Scan(
		&depID, &depIata, &depTitle, &depCity, &depCn,
		&arrID, &arrIata, &arrTitle, &arrCity, &arrCn,
	); err != nil {
		return airportDomain.Airport{}, airportDomain.Airport{}, err
	}

	dep, err := newAirport(depID, depIata, depTitle, depCity, depCn)
	if err != nil {
		return airportDomain.Airport{}, airportDomain.Airport{}, err
	}
	arr, err := newAirport(arrID, arrIata, arrTitle, arrCity, arrCn)
	if err != nil {
		return airportDomain.Airport{}, airportDomain.Airport{}, err
	}
	return dep, arr, nil
}

func newAirport(id uuid.UUID, iata, title, city, country string) (airportDomain.Airport, error) {
	iataCode, err := airportDomain.NewIATACode(iata)
	if err != nil {
		return airportDomain.Airport{}, err
	}
	titleValue, err := airportDomain.NewTitle(title)
	if err != nil {
		return airportDomain.Airport{}, err
	}
	cityValue, err := common.NewCity(city)
	if err != nil {
		return airportDomain.Airport{}, err
	}
	countryValue, err := common.NewCountry(country)
	if err != nil {
		return airportDomain.Airport{}, err
	}
	return airportDomain.Airport{
		ID:       id,
		IATACode: iataCode,
		Title:    titleValue,
		City:     cityValue,
		Country:  countryValue,
	}, nil
}

func isNotFoundSignal(err error, message string) bool {
	return strings.Contains(err.Error(), message)
}
