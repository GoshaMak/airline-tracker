package repository

import (
	"context"
	"errors"
	"notifier/internal/receiver/domain"
	"time"

	"github.com/google/uuid"
)

var ErrLeaseLost = errors.New("notification lease lost")

type NotificationRepository interface {
	Save(ctx context.Context, n domain.Notification) error
	ApplyEvent(
		ctx context.Context,
		eventID string,
		version int,
		carrier map[string]string,
		notifications []domain.Notification,
	) (bool, error)
	RecordEventFailure(ctx context.Context, event domain.EventFailure, cause string, maxAttempts int) (bool, error)
	ClaimDue(ctx context.Context, workerID string, limit int, lease time.Duration) ([]domain.Notification, error)
	Complete(ctx context.Context, id, leaseToken uuid.UUID) error
	Retry(ctx context.Context, id, leaseToken uuid.UUID, nextAttempt time.Time, cause string, maxAttempts int) error
}
