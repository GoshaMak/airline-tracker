package repository

import (
	"context"
	"notifier/internal/receiver/domain"
)

type NotificationRepository interface {
	Save(ctx context.Context, n domain.Notification) error

	ListNotSent(ctx context.Context) ([]domain.Notification, error)

	Mark(ctx context.Context, notification domain.Notification, newStatus domain.NotificationStatus) error
}
