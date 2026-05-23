package repository

import (
	"context"
	"notifier/internal/receiver/domain"
)

type NotificationRepository interface {
	Save(ctx context.Context, n domain.Notification) error

	ListByStatus(ctx context.Context, status domain.NotificationStatus) ([]domain.Notification, error)

	Mark(ctx context.Context, notification domain.Notification, newStatus domain.NotificationStatus) error
}
