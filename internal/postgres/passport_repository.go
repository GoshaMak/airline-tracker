package postgres

import (
	"airline-tracker/internal/domain"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/do/v2"
)

// type PassportRepository interface {
// 	Save(ctx context.Context, p *domain.Passport) error
// 	GetById(ctx context.Context, id uint32)
// }

type PassportRepository struct {
	db *pgx.Conn
}

func NewPassportRepository(i do.Injector) (*PassportRepository, error) {
	return &PassportRepository{
		db: do.MustInvoke[*pgx.Conn](i),
	}, nil
}

var (
	ErrPassportRepositoryNil = errors.New("PassportRepository is nil")
	ErrPassportAlreadyExists = errors.New("Passport already exists")
	ErrPassportNotExists     = errors.New("Passport doesn't exist")
)

func (r *PassportRepository) Save(ctx context.Context, p *domain.Passport) error {
	var query string
	if p.SecondName == "" {
		query = "insert into passports(number, issue_date, name," +
			" surname, gender, birthday, birth_city, birth_country)" +
			" values ($1, $2, $3, $4, $5, $6, $7, $8);"
	} else {
		query = "insert into passports(number, issue_date, name," +
			" second_name, surname, gender, birthday, birth_city, birth_country)" +
			" values ($1, $2, $3, $4, $5, $6, $7, $8, $9);"
	}
	_, err := r.db.Exec(ctx, query,
		p.Number, p.IssueDate, p.Name, p.SecondName, p.Surname, p.Gender, p.Birthday, p.BirthCity, p.BirthCountry)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == string(RecordAlreadyExistsErrCode) {
			slog.Error("Passport already exists", "passport", *p)
			return ErrPassportAlreadyExists
		}
		slog.Error("Can't insert new user", "error", err, "user", *p)
		return ErrInsertFailure
	}
	return nil
}

func (r *PassportRepository) GetById(ctx context.Context, id uint32) (*domain.Passport, error) {
	op := "postgres.passport_repository.GetById"
	query := "select id, number, issue_date, name, second_name, surname," +
		" gender, birthday, birth_city, birth_country from passports" +
		" from passports" +
		" where id = $1;"
	p := &domain.Passport{}
	var secondName *string
	err := r.db.QueryRow(ctx, query, id).
		Scan(p.ID, p.Number, p.IssueDate, p.Name, &secondName,
			p.Surname, p.Gender, p.Birthday, p.BirthCity, p.BirthCountry)
	if err != nil {
		slog.Warn("Can't find passport", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if secondName != nil {
		p.SecondName = *secondName
	}
	return p, nil
}
