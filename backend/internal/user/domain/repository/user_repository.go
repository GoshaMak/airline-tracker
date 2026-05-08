package repository

import (
	flightDomain "airline-tracker/internal/flight/domain"
	"airline-tracker/internal/user/domain"
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	SaveUser(ctx context.Context, user domain.User) error

	GetUser(ctx context.Context, email string) (domain.User, error)

	Exists(ctx context.Context, uid uuid.UUID) (domain.User, error)

	UpdateById(ctx context.Context, uid uuid.UUID) error

	Subscribe(ctx context.Context, uid, fid uuid.UUID) error

	ListFlights(ctx context.Context, uid uuid.UUID) ([]flightDomain.Flight, error)
}
