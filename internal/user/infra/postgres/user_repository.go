package postgres

import (
	flightDomain "airline-tracker/internal/flight/domain"
	"airline-tracker/internal/user/domain"
	"airline-tracker/internal/user/domain/repository"
	"context"
	"errors"
	"fmt"
	"log/slog"
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
		user.ID, user.Email, user.Password, user.Role)
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
	u, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[domain.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return u, nil
}

func (r *userRepository) Exists(ctx context.Context, uid uuid.UUID) (domain.User, error) {
	return domain.User{}, nil
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
		if errors.As(err, &pgErr) && pgErr.Code == "P0002" {
			slog.Debug(op, "err", err)
			if strings.HasPrefix(pgErr.Message, "user") {
				return repository.ErrUserNotFound
			} else if strings.HasPrefix(pgErr.Message, "flight") {
				return repository.ErrFlightNotFound
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
	flights, err := pgx.CollectRows(rows, pgx.RowToStructByName[flightDomain.Flight])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return flights, nil
}
