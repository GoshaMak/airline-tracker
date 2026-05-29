package mysql

import (
	flightDomain "api/internal/flight/domain"
	dbmysql "api/internal/infra/mysql"
	"api/internal/user/domain"
	"api/internal/user/domain/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"shared/common"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type userRepository struct {
	conn *sql.DB
}

func NewUserRepository(i do.Injector) (repository.UserRepository, error) {
	return &userRepository{
		conn: do.MustInvoke[*sql.DB](i),
	}, nil
}

func (r *userRepository) SaveUser(ctx context.Context, user domain.User) error {
	const op = "UserRepository.Save"
	query := `
	insert into users(id, email, password_hash, role)
		values (?, ?, ?, ?)
	`
	_, err := r.conn.ExecContext(ctx, query,
		user.Id.String(),
		user.Email.String(),
		user.PasswordHash.String(),
		user.Role.String(),
	)
	if err != nil {
		if dbmysql.IsDuplicate(err) {
			return repository.ErrUserAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *userRepository) GetUser(ctx context.Context, email string) (domain.User, error) {
	const op = "UserRepository.GetUser"
	query := `
	select id, email, password_hash, role from users where email = ?
	`
	user, err := scanUser(r.conn.QueryRowContext(ctx, query, email))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (r *userRepository) Exist(ctx context.Context, uid uuid.UUID) (domain.User, error) {
	const op = "UserRepository.Exists"
	query := `
	select id, email, password_hash, role from users where id = ?
	`
	user, err := scanUser(r.conn.QueryRowContext(ctx, query, uid.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (r *userRepository) Subscribe(
	ctx context.Context,
	uid,
	fid uuid.UUID,
) error {
	const op = "UserRepository.Subscribe"
	query := `
	call subscribe(?, ?)
	`
	_, err := r.conn.ExecContext(ctx, query, uid.String(), fid.String())
	if err != nil {
		if dbmysql.IsDuplicate(err) {
			return repository.ErrUserAlreadySubscribed
		}
		if dbmysql.IsNotFoundSignal(err, "user not found") {
			return repository.ErrUserNotFound
		}
		if dbmysql.IsNotFoundSignal(err, "flight not found") {
			return repository.ErrFlightNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *userRepository) ListFlights(
	ctx context.Context,
	uid uuid.UUID,
) ([]flightDomain.Flight, error) {
	const op = "UserRepository.ListFlights"
	query := `
	call scan_user_flights_info(?)
	`
	rows, err := r.conn.QueryContext(ctx, query, uid.String())
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	flights := make([]flightDomain.Flight, 0)
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

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(scanner userScanner) (domain.User, error) {
	var (
		id               uuid.UUID
		email, pswd, rol string
	)
	if err := scanner.Scan(&id, &email, &pswd, &rol); err != nil {
		return domain.User{}, err
	}

	mail, err := common.NewEmail(email)
	if err != nil {
		return domain.User{}, err
	}
	role, err := domain.NewRole(rol)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:           id,
		Email:        mail,
		PasswordHash: domain.PasswordHashed(pswd),
		Role:         role,
	}, nil
}

func scanFlight(scanner userScanner) (flightDomain.Flight, error) {
	var (
		id, aircraftId, depAirportId, arrAirportId uuid.UUID
		depGateId, arrGateId                       uuid.UUID
		scheduledDeparture, scheduledArrival       time.Time
		actualDeparture, actualArrival             sql.NullTime
		status                                     string
		plan                                       sql.NullString
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
		return flightDomain.Flight{}, err
	}

	st, err := flightDomain.NewFlightStatus(status)
	if err != nil {
		return flightDomain.Flight{}, err
	}
	var p *flightDomain.FlightPlan
	if plan.Valid {
		pv, err := flightDomain.NewFlightPlan(plan.String)
		if err != nil {
			return flightDomain.Flight{}, err
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

	return flightDomain.Flight{
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
