package postgres

import (
	"airline-tracker/internal/user/domain"
	"airline-tracker/internal/user/domain/repository"
	"context"
	"errors"
	"fmt"
	"log/slog"

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

func (r *userRepository) Save(ctx context.Context, u *domain.User) error {
	query := `
	insert into users(id, email, password, role)
		values ($1, $2, $3, $4)
	`
	_, err := r.conn.Exec(ctx, query,
		u.ID, u.Email, u.Password, u.Role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			slog.Error("User already exists", "user", *u)
			return ErrRecordAlreadyExists
		}
		slog.Error("Can't insert new user", "error", err, "user", *u)
		return ErrInsertFailure
	}
	return nil
}

func (r *userRepository) GetUser(ctx context.Context, email string) (domain.User, error) {
	op := "postgres.user_repository.GetUser"
	query := `
	select * from users where email = $1
	`

	row, err := r.conn.Query(ctx, query)
	if err != nil {
		return domain.User{}, err
	}
	defer row.Close()

	u, err := pgx.CollectOneRow(row, pgx.RowToStructByName[domain.User])
	if err != nil {
		slog.Warn("Can't find user", "err", err)
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return u, nil
}

func (r *userRepository) Exists(ctx context.Context, id uint32) (domain.User, error) {
	return domain.User{}, nil
}

func (r *userRepository) UpdateById(ctx context.Context, id uint32) error {
	return nil
}
