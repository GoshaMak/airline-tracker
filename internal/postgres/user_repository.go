package postgres

import (
	"airline-tracker/internal/domain"
	"airline-tracker/internal/domain/repository"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/do/v2"
)

type userRepository struct {
	conn *pgx.Conn
}

func NewUserRepository(i do.Injector) (repository.UserRepository, error) {
	return &userRepository{
		conn: do.MustInvoke[*pgx.Conn](i),
	}, nil
}

var (
	ErrUserRepositoryNil = errors.New("UserRepository is nil")
	ErrArgNil            = errors.New("Argument is nil")
	ErrInsertFailure     = errors.New("Can't insert")
	ErrUserAlreadyExists = errors.New("User already exists")
	ErrUserNotExists     = errors.New("User doesn't exist")
)

// FIX:
func (r *userRepository) Save(ctx context.Context, u *domain.User) error {
	_, err := r.conn.Exec(
		ctx,
		"insert into users(passport_id, card_id, email, phone, password)"+
			"values (null, null, $1, $2, $3);",
		u.Email, u.Phone, u.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == string(RecordAlreadyExistsErrCode) {
			slog.Error("User already exists", "user", *u)
			return ErrUserAlreadyExists
		}
		slog.Error("Can't insert new user", "error", err, "user", *u)
		return ErrInsertFailure
	}
	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	op := "postgres.user_repository.GetByEmail"
	u := &domain.User{}
	var pid *uint32
	var cid *uint32
	err := r.conn.QueryRow(
		ctx,
		"select id, passport_id, card_id, email, phone, password"+
			" from users"+
			" where email = $1;", email).
		Scan(&u.ID, &pid, &cid, &u.Email, &u.Phone, &u.Password)
	if err != nil {
		slog.Warn("Can't find user", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return u, nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	op := "postgres.user_repository.GetByPhone"
	u := &domain.User{}
	var pid *uint32
	var cid *uint32
	err := r.conn.QueryRow(
		ctx,
		"select id, passport_id, card_id, email, phone, password"+
			"from users"+
			"where phone = $1;", phone).
		Scan(&u.ID, &pid, &cid, &u.Email, &u.Phone, &u.Password)
	if err != nil {
		slog.Warn("Can't find user", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if pid != nil {
		u.PassportID = *pid
	}
	return u, nil
}

func (r *userRepository) Exists(ctx context.Context, id uint32) (*domain.User, error)

func (r *userRepository) UpdateById(ctx context.Context, id uint32) error
