package postgres

import (
	infraPostgres "api/internal/infra/postgres"
	"errors"
	"log/slog"
)

var (
	ErrFlightNotFound        = errors.New("flight not found")
	ErrFlightRouteNotFound   = errors.New("flight route not found")
	ErrOperationNotPermitted = errors.New("postgres operation not permitted")
)

func translateError(err error) error {
	if infraPostgres.IsAccessError(err) {
		slog.Error("PostgreSQL operation rejected", "err", err)
		return ErrOperationNotPermitted
	}
	return err
}
