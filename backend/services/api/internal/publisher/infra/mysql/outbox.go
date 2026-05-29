package mysql

import (
	"api/internal/publisher/domain"
	"api/internal/publisher/domain/repository"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type outboxRepository struct {
	conn *sql.DB
}

func NewOutboxRepository(i do.Injector) (repository.OutboxRepository, error) {
	return &outboxRepository{
		conn: do.MustInvoke[*sql.DB](i),
	}, nil
}

func (r *outboxRepository) Save(
	ctx context.Context,
	ob domain.Outbox,
) error {
	const op = "OutboxRepository.Save"
	query := `
	insert into outbox(id, topic, payload, created_at)
		values(?, ?, ?, ?)
	`

	payload, err := ob.Payload.MarshalJSON()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = r.conn.ExecContext(ctx, query,
		ob.Id.String(), ob.Topic, payload, ob.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *outboxRepository) ListNotSent(
	ctx context.Context,
	newPayload func() domain.Payload,
) ([]domain.Outbox, error) {
	const op = "OutboxRepository.ListNotSent"
	query := `
	select id, topic, payload, created_at, sent_at
	from outbox
	where sent_at is null
	`

	rows, err := r.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	obs := make([]domain.Outbox, 0)
	for rows.Next() {
		ob, err := scanOutbox(rows, newPayload())
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		obs = append(obs, ob)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if len(obs) > 0 {
		slog.Debug(op, "obs", obs)
	}

	return obs, nil
}

func (r *outboxRepository) MarkAsSent(ctx context.Context, ob domain.Outbox) error {
	const op = "OutboxRepository.MarkNotSent"
	query := `
	update outbox
	set sent_at = ?
	where id = ?
	`

	_, err := r.conn.ExecContext(ctx, query, ob.SentAt, ob.Id.String())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

type outboxScanner interface {
	Scan(dest ...any) error
}

func scanOutbox(scanner outboxScanner, payload domain.Payload) (domain.Outbox, error) {
	var (
		id        uuid.UUID
		topic     string
		body      []byte
		createdAt time.Time
		sentAt    sql.NullTime
	)
	if err := scanner.Scan(&id, &topic, &body, &createdAt, &sentAt); err != nil {
		return domain.Outbox{}, err
	}
	if err := payload.UnmarshalJSON(body); err != nil {
		return domain.Outbox{}, err
	}

	var sentAtPtr *time.Time
	if sentAt.Valid {
		sentAtPtr = &sentAt.Time
	}

	return domain.Outbox{
		Id:        id,
		Topic:     topic,
		Payload:   payload,
		CreatedAt: createdAt,
		SentAt:    sentAtPtr,
	}, nil
}
