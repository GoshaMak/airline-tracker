package postgres

import (
	"context"
	"fmt"
	"notifier/internal/receiver/domain"
	"notifier/internal/receiver/domain/repository"
	"notifier/internal/receiver/infra/postgres/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

type notificationRepository struct {
	conn *pgxpool.Pool
}

func NewNotificationRepository(i do.Injector) (repository.NotificationRepository, error) {
	return &notificationRepository{
		conn: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

func (r *notificationRepository) Save(
	ctx context.Context,
	n domain.Notification,
) error {
	const op = "NotificationRepository.Save"
	query := `
	insert into notifications(id, payload, created_at, send_at, status)
	values ($1, $2, $3, $4, $5)
	`

	_, err := r.conn.Exec(ctx, query,
		n.Id, n.Payload, n.CreatedAt, n.SendAt, n.Status)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *notificationRepository) ListByStatus(
	ctx context.Context,
	status domain.NotificationStatus,
) ([]domain.Notification, error) {
	const op = "NotificationRepository.ListByStatus"
	query := `
	select *
	from notifications
	where status = $1
	`
	rows, _ := r.conn.Query(ctx, query, status)
	nms, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.NotificationModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	nds := make([]domain.Notification, len(nms))
	for i, nm := range nms {
		nd, err := model.NotificationModelToDomain(nm)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		nds[i] = nd
	}

	return nds, nil
}

func (r *notificationRepository) Mark(
	ctx context.Context,
	n domain.Notification,
	newStatus domain.NotificationStatus,
) error {
	const op = "NotificationRepository.Mark"
	query := `
	update notifications
	set status = $1
	where id = $2
	`

	_, err := r.conn.Exec(ctx, query, newStatus, n.Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
