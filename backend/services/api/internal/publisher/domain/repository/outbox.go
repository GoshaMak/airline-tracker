package repository

import (
	"api/internal/publisher/domain"
	"context"
)

type OutboxRepository interface {
	Save(ctx context.Context, ob domain.Outbox) error

	ListNotSent(ctx context.Context) ([]domain.Outbox, error)

	MarkAsSent(ctx context.Context, ob domain.Outbox) error
}
