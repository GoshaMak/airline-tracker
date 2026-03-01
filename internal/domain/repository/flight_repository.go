package repository

import (
	"airline-tracker/internal/domain"
	"context"
)

type FlightRepository interface {
	Save(ctx context.Context, u *domain.Flight) error

	Exists(ctx context.Context, id uint32) (*domain.Flight, error)

	UpdateById(ctx context.Context, id uint32) error
}
