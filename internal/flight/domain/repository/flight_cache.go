package repository

import (
	"airline-tracker/internal/flight/domain"
	"context"

	"github.com/google/uuid"
)

type FlightCache interface {
	Set(ctx context.Context, f *domain.Flight) error

	GetByID(ctx context.Context, id *uuid.UUID) (*domain.Flight, error)

	Delete(ctx context.Context, id *uuid.UUID) error
}
