package postgres

import (
	flightDomain "api/internal/flight/domain"
	flightModel "api/internal/flight/infra/postgres/model"
	"api/internal/user/domain"
	"api/internal/user/domain/repository"
	"api/internal/user/infra/postgres/model"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

type userRepository struct {
	conn *pgxpool.Pool
}

func NewUserRepository(i do.Injector) (repository.UserRepository, error) {
	return &userRepository{
		conn: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

func (r *userRepository) SaveUser(ctx context.Context, user domain.User) error {
	op := "UserRepository.Save"
	query := `
	insert into users(id, email, password, role)
		values ($1, $2, $3, $4)
	`
	_, err := r.conn.Exec(ctx, query,
		user.Id, user.Email.String(),
		user.PasswordHash.String(), user.Role.String())
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrUserAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *userRepository) GetUser(ctx context.Context, email string) (domain.User, error) {
	op := "UserRepository.GetUser"
	query := `
	select * from users where email = $1
	`
	row, _ := r.conn.Query(ctx, query, email)
	um, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[model.UserModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	ud, err := model.UserModelToDomain(um)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return ud, nil
}

func (r *userRepository) Exist(ctx context.Context, uid uuid.UUID) (domain.User, error) {
	op := "UserRepository.Exists"
	query := `
	select * from users where id = $1
	`
	row, _ := r.conn.Query(ctx, query, uid)
	um, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[model.UserModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	ud, err := model.UserModelToDomain(um)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return ud, nil
}

func (r *userRepository) UpdateById(ctx context.Context, uid uuid.UUID) error {
	return nil
}

func (r *userRepository) Subscribe(
	ctx context.Context,
	uid, fid uuid.UUID,
) error {
	op := "UserRepository.Subscribe"
	query := `
	select subscribe($1, $2)
	`
	_, err := r.conn.Exec(ctx, query, uid, fid)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "P0002" {
				if strings.HasPrefix(pgErr.Message, "user") {
					return repository.ErrUserNotFound
				} else if strings.HasPrefix(pgErr.Message, "flight") {
					return repository.ErrFlightNotFound
				}
			} else if pgErr.Code == "23505" &&
				pgErr.ConstraintName == "unique_flight_subscription_per_user" {
				return repository.ErrUserAlreadySubscribed
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *userRepository) ListFlights(
	ctx context.Context,
	uid uuid.UUID,
) ([]flightDomain.Flight, error) {
	op := "UserRepository.ListFlights"
	query := `
	select * from scan_user_flights_info($1)
	`
	rows, _ := r.conn.Query(ctx, query, uid)
	fsMs, err := pgx.CollectRows(rows, pgx.RowToStructByName[flightModel.FlightModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	fsDs := make([]flightDomain.Flight, len(fsMs))
	for i, fm := range fsMs {
		fd, err := flightModel.FlightModelToDomain(fm)
		if err != nil {
			return nil, err
		}
		fsDs[i] = fd
	}
	return fsDs, nil
}
