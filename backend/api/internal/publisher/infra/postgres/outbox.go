package postgres

import (
	"api/internal/infra/postgres"
	"api/internal/publisher/domain"
	"api/internal/publisher/infra/postgres/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"shared/observability"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

var (
	ErrOperationNotPermitted = errors.New("repository operation not permitted")
	ErrLeaseLost             = errors.New("outbox lease lost")
)

type OutboxRepository interface {
	Save(ctx context.Context, ob domain.Outbox) error
	Claim(ctx context.Context, limit int, lease time.Duration) ([]domain.Outbox, error)
	Complete(ctx context.Context, id, leaseToken uuid.UUID) error
	Retry(ctx context.Context, id, leaseToken uuid.UUID, nextAttempt time.Time, cause string, maxAttempts int) error
}

type outboxDB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	Begin(context.Context) (pgx.Tx, error)
}

type outboxRepository struct {
	conn outboxDB
}

func NewOutboxRepository(i do.Injector) (OutboxRepository, error) {
	return &outboxRepository{
		conn: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

func (r *outboxRepository) Save(
	ctx context.Context,
	ob domain.Outbox,
) error {
	const op = "OutboxRepository.Save"
	const query = `
	insert into api_schema.outbox(id, topic, payload, created_at, trace_carrier, correlation_id)
	values($1, $2, $3::jsonb, $4, $5::jsonb, $6)
	`

	obm, err := model.OutboxDomainToModel(ob)
	if err != nil {
		return fmt.Errorf("%s: %w", op, translateError(err))
	}

	carrier := observability.Inject(ctx)
	carrierJSON, err := json.Marshal(carrier)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = postgres.Executor(ctx, r.conn).Exec(ctx, query,
		obm.Id, obm.Topic, obm.Payload, obm.CreatedAt, carrierJSON, observability.CorrelationID(ctx),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, translateError(err))
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
	const query = `
	select id, topic, payload, created_at, sent_at, attempts, lease_token,
		next_attempt_at, trace_carrier, correlation_id
	from api_schema.outbox
	where sent_at is null and failed_at is null
	`

	rows, err := r.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, translateError(err))
	}
	obms, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.OutboxModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, translateError(err))
	}
	if len(obms) > 0 {
		slog.DebugContext(ctx, op, "obms", obms)
	}
	obs := make([]domain.Outbox, len(obms))
	for i, obm := range obms {
		payload := newPayload()
		ob, err := model.OutboxModelToDomain(obm, payload)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, translateError(err))
		}
		obs[i] = ob
	}

	return obs, nil
}

func (r *outboxRepository) MarkAsSent(ctx context.Context, ob domain.Outbox) error {
	const op = "OutboxRepository.MarkNotSent"
	const query = `
	update outbox
	set sent_at = $1
	where id = $2
	`

	_, err := r.conn.Exec(ctx, query, ob.SentAt, ob.Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, translateError(err))
	}
	return nil
}

func (r *outboxRepository) Claim(
	ctx context.Context,
	limit int,
	lease time.Duration,
) ([]domain.Outbox, error) {
	const op = "OutboxRepository.Claim"
	const query = `
	with exhausted as (
		update api_schema.outbox
		set failed_at=now(),
			last_error=coalesce(last_error,'lease expired after maximum attempts'),
			lease_token=null,
			lease_until=null
		where sent_at is null and
			failed_at is null and
			lease_until < now() and
			attempts >= 8
	), eligible as (
		select id from api_schema.outbox
		where sent_at is null and
			failed_at is null and
			attempts < 8 and
			next_attempt_at <= now() and
			(lease_until is null or lease_until < now())
		order by created_at, id
		for update skip locked limit $1
	), claimed as (
		update api_schema.outbox o
		set lease_token = gen_random_uuid(),
			lease_until = now() + ($2 * interval '1 millisecond'),
			attempts = attempts + 1
	 	from eligible e where o.id = e.id
	 	returning o.id, o.topic, o.payload, o.created_at, o.sent_at, o.attempts,
			o.lease_token, o.next_attempt_at, o.trace_carrier, o.correlation_id
	)
	select id, topic, payload, created_at, sent_at, attempts,
		lease_token, next_attempt_at, trace_carrier, correlation_id
	from claimed
	`
	if limit < 1 || limit > 100 || lease <= 0 {
		return nil, fmt.Errorf("%s: invalid claim limits", op)
	}
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, translateError(err))
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, query, limit, lease.Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, translateError(err))
	}
	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.OutboxModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	items := make([]domain.Outbox, 0, len(models))
	for _, m := range models {
		item, err := model.OutboxModelToDomain(m, &rawPayload{})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		items = append(items, item)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, translateError(err))
	}
	stats, statsErr := r.conn.Query(ctx, `
		select count(*), coalesce(extract(epoch from (now()-min(created_at))),0)
		from api_schema.outbox
		where sent_at is null and
			failed_at is null
		`)
	if statsErr == nil {
		type queueStats struct {
			pending int64
			oldest  float64
		}
		if result, collectErr := pgx.CollectExactlyOneRow(stats, func(row pgx.CollectableRow) (queueStats, error) {
			var value queueStats
			err := row.Scan(&value.pending, &value.oldest)
			return value, err
		}); collectErr == nil {
			observability.SetQueue("publisher", "outbox", float64(result.pending), result.oldest)
		}
	}
	return items, nil
}

// rawPayload preserves the JSON bytes until the event is sent to Kafka
type rawPayload []byte

func (p *rawPayload) MarshalJSON() ([]byte, error) {
	return append([]byte(nil), (*p)...), nil
}

func (p *rawPayload) UnmarshalJSON(b []byte) error {
	*p = append((*p)[:0], b...)
	return nil
}

func (r *outboxRepository) Complete(ctx context.Context, id, leaseToken uuid.UUID) error {
	const op = "OutboxRepository.Complete"
	const query = `
	update api_schema.outbox
	set sent_at=now(),
		lease_token=null,
		lease_until=null,
		last_error=null
	where id=$1 and
		lease_token=$2 and
		sent_at is null and
		failed_at is null
	`
	tag, err := r.conn.Exec(ctx, query, id, leaseToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, translateError(err))
	}
	if tag.RowsAffected() != 1 {
		return ErrLeaseLost
	}
	return nil
}

func (r *outboxRepository) Retry(
	ctx context.Context,
	id,
	leaseToken uuid.UUID,
	nextAttempt time.Time,
	cause string,
	maxAttempts int,
) error {
	const op = "OutboxRepository.Retry"
	const query = `
	update api_schema.outbox
	set failed_at=case when attempts >= $4 then now() else null end,
		next_attempt_at=case when attempts >= $4 then next_attempt_at else $3 end,
		last_error=left($5, 1000),
		lease_token=null,
		lease_until=null
	where id=$1 and
		lease_token=$2 and
		sent_at is null and
		failed_at is null
	`
	tag, err := r.conn.Exec(ctx, query, id, leaseToken, nextAttempt, maxAttempts, cause)
	if err != nil {
		return fmt.Errorf("%s: %w", op, translateError(err))
	}
	if tag.RowsAffected() != 1 {
		return ErrLeaseLost
	}
	return nil
}

func translateError(err error) error {
	if postgres.IsAccessError(err) {
		slog.Error("PostgreSQL operation rejected", "err", err)
		return ErrOperationNotPermitted
	}
	return err
}
