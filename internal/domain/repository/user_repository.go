package repository

import (
	"airline-tracker/internal/domain"
	"context"
)

type UserRepository interface {
	Save(ctx context.Context, u *domain.User) error

	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	GetByPhone(ctx context.Context, phone string) (*domain.User, error)

	Exists(ctx context.Context, id uint32) (*domain.User, error)

	UpdateById(ctx context.Context, id uint32) error
}
