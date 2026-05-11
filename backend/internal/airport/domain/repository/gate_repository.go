package repository

import (
	"airline-tracker/internal/airport/domain"
	"context"

	"github.com/google/uuid"
)

type GateRepository interface {
	Save(ctx context.Context, g domain.Gate) error

	Exists(ctx context.Context, g domain.Gate) (domain.Gate, error)

	GetAirportByGateId(ctx context.Context, gid uuid.UUID) (domain.Airport, error)
}
