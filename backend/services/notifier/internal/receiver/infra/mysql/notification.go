package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"notifier/internal/receiver/domain"
	"notifier/internal/receiver/domain/repository"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type notificationRepository struct {
	conn *sql.DB
}

func NewNotificationRepository(i do.Injector) (repository.NotificationRepository, error) {
	return &notificationRepository{
		conn: do.MustInvoke[*sql.DB](i),
	}, nil
}

func (r *notificationRepository) Save(
	ctx context.Context,
	n domain.Notification,
) error {
	const op = "NotificationRepository.Save"
	query := `
	insert into notifications(id, payload, created_at, send_at, status, type)
	values (?, ?, ?, ?, ?, ?)
	`

	_, err := r.conn.ExecContext(ctx, query,
		n.Id.String(), n.Payload, n.CreatedAt, n.SendAt, n.Status.String(), n.Type.String(),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *notificationRepository) ListNotSent(
	ctx context.Context,
) ([]domain.Notification, error) {
	const op = "NotificationRepository.ListByStatus"
	query := `
	select id, payload, created_at, send_at, status, type
	from notifications
	where status <> ?
	`
	rows, err := r.conn.QueryContext(ctx, query, domain.NotificationSent.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	notifications := make([]domain.Notification, 0)
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return notifications, nil
}

func (r *notificationRepository) Mark(
	ctx context.Context,
	n domain.Notification,
	newStatus domain.NotificationStatus,
) error {
	const op = "NotificationRepository.Mark"
	query := `
	update notifications
	set status = ?
	where id = ?
	`

	_, err := r.conn.ExecContext(ctx, query, newStatus.String(), n.Id.String())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

type notificationScanner interface {
	Scan(dest ...any) error
}

func scanNotification(scanner notificationScanner) (domain.Notification, error) {
	var (
		id                uuid.UUID
		payload           []byte
		createdAt, sendAt time.Time
		status, typ       string
	)
	if err := scanner.Scan(&id, &payload, &createdAt, &sendAt, &status, &typ); err != nil {
		return domain.Notification{}, err
	}

	notificationStatus, err := domain.NewNotificationStatus(status)
	if err != nil {
		return domain.Notification{}, err
	}
	notificationType, err := domain.NewNotificationType(typ)
	if err != nil {
		return domain.Notification{}, err
	}

	return domain.Notification{
		Id:        id,
		Payload:   payload,
		CreatedAt: createdAt,
		SendAt:    sendAt,
		Status:    notificationStatus,
		Type:      notificationType,
	}, nil
}
