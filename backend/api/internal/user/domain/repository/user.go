package repository

import (
	flightDomain "api/internal/flight/domain"
	"api/internal/user/domain"
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	SaveUser(ctx context.Context, user domain.User) error

	GetUser(ctx context.Context, email string) (domain.User, error)

	UpdatePassword(ctx context.Context, uid uuid.UUID, passwordHash domain.PasswordHashed) error

	Exist(ctx context.Context, uid uuid.UUID) (domain.User, error)

	Subscribe(ctx context.Context, uid, fid uuid.UUID) error

	ListFlights(ctx context.Context, uid uuid.UUID) ([]flightDomain.Flight, error)
}
