package repository

import (
	"airline-ticketing-svc/internal/domain"
	"context"
)

type UserRepository interface {
	Save(ctx context.Context, u *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
}
