package repository

import (
	flightDomain "api/internal/flight/domain"
	"context"

	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, uid, fid uuid.UUID, notifyBeforeMinutes int64) (timerAdded bool, err error)

	AddTimers(ctx context.Context, uid, fid uuid.UUID, notifyBeforeMinutes []int64) ([]int64, error)

	Delete(ctx context.Context, uid, fid uuid.UUID) error

	ListFlights(ctx context.Context, uid uuid.UUID) ([]flightDomain.Flight, error)
}
