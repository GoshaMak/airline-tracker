package postgres

import (
	"api/internal/publisher/domain"
	"api/internal/publisher/domain/repository"
	"api/internal/publisher/infra/postgres/model"
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

type outboxRepository struct {
	conn *pgxpool.Pool
}

func NewOutboxRepository(i do.Injector) (repository.OutboxRepository, error) {
	return &outboxRepository{
		conn: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

func (r *outboxRepository) Save(
	ctx context.Context,
	ob domain.Outbox,
) error {
	const op = "OutboxRepository.Save"
	query := `
	insert into outbox(id, topic, payload, created_at)
		values($1, $2, $3::jsonb, $4)
	`

	obm, err := model.OutboxDomainToModel(ob)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = r.conn.Exec(ctx, query,
		obm.Id, obm.Topic, obm.Payload, obm.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func clone(p domain.Payload) domain.Payload {
	if p == nil {
		return nil
	}

	rv := reflect.ValueOf(p)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return p
	}

	elem := rv.Elem()
	cp := reflect.New(elem.Type())
	cp.Elem().Set(elem)

	return cp.Interface().(domain.Payload)
}

func (r *outboxRepository) ListNotSent(
	ctx context.Context,
	newPayload func() domain.Payload,
) ([]domain.Outbox, error) {
	const op = "OutboxRepository.ListNotSent"
	query := `
	select *
	from outbox
	where sent_at is null
	`

	rows, _ := r.conn.Query(ctx, query)
	obms, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.OutboxModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if len(obms) > 0 {
		slog.Debug(op, "obms", obms)
	}
	obs := make([]domain.Outbox, len(obms))
	for i, obm := range obms {
		payload := newPayload()
		ob, err := model.OutboxModelToDomain(obm, payload)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		obs[i] = ob
	}

	return obs, nil
}

func (r *outboxRepository) MarkAsSent(ctx context.Context, ob domain.Outbox) error {
	const op = "OutboxRepository.MarkNotSent"
	query := `
	update outbox
	set sent_at = $1
	where id = $2
	`

	_, err := r.conn.Exec(ctx, query, ob.SentAt, ob.Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
