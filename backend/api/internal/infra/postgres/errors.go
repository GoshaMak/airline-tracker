package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func IsAccessError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr != nil {
		switch pgErr.Code {
		case "25006", "42501":
			return true
		}
	}
	return false
}
