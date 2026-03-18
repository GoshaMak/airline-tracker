package repository

import (
	"airline-tracker/internal/airport/domain"
	"context"
)

type GateRepository interface {
	Save(ctx context.Context, a *domain.Gate) error

	Exists(ctx context.Context, a *domain.Gate) (*domain.Gate, error)
}
