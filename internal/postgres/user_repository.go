package postgres

import (
	"airline-ticketing-svc/internal/domain"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/do/v2"
)

type UserRepository struct {
	db *pgx.Conn
}

func NewUserRepository(i do.Injector) (*UserRepository, error) {
	return &UserRepository{
		db: do.MustInvoke[*pgx.Conn](i),
	}, nil
}

var (
	ErrUserRepositoryNil = errors.New("UserRepository is nil")
	ErrArgNil            = errors.New("Argument is nil")
	ErrInsertFailure     = errors.New("Can't insert")
	ErrUserAlreadyExists = errors.New("User already exists")
	ErrUserNotExists     = errors.New("User doesn't exist")
)

// User creates with null passport and card ids
func (r *UserRepository) Save(ctx context.Context, u *domain.User) error {
	_, err := r.db.Exec(
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

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	op := "postgres.user_repository.GetByEmail"
	u := &domain.User{}
	var pid *uint32
	var cid *uint32
	err := r.db.QueryRow(
		ctx,
		"select id, passport_id, card_id, email, phone, password"+
			" from users"+
			" where email = $1;", email).
		Scan(&u.ID, &pid, &cid, &u.Email, &u.Phone, &u.Password)
	if err != nil {
		slog.Warn("Can't find user", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if pid != nil {
		u.PassportID = *pid
	}
	if cid != nil {
		u.CardID = *cid
	}
	return u, nil
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	op := "postgres.user_repository.GetByPhone"
	u := &domain.User{}
	var pid *uint32
	var cid *uint32
	err := r.db.QueryRow(
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
	if cid != nil {
		u.CardID = *cid
	}
	return u, nil
}
