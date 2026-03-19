package repository

import (
	"airline-tracker/internal/airport/domain"
	"context"
)

type GateRepository interface {
	Save(ctx context.Context, g *domain.Gate) error

	Exists(ctx context.Context, g *domain.Gate) (*domain.Gate, error)
}
