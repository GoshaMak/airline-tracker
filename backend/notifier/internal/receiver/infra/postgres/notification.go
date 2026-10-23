package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"notifier/internal/receiver/domain"
	"notifier/internal/receiver/domain/repository"
	"notifier/internal/receiver/infra/postgres/model"
	"shared/observability"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

type notificationRepository struct{ conn notificationDB }

type notificationDB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	Begin(context.Context) (pgx.Tx, error)
}

func NewNotificationRepository(i do.Injector) (repository.NotificationRepository, error) {
	return &notificationRepository{conn: do.MustInvoke[*pgxpool.Pool](i)}, nil
}

func (r *notificationRepository) Save(ctx context.Context, n domain.Notification) error {
	const op = "NotificationRepository.Save"
	const query = `
	insert into notifier_schema.notifications
		(id, payload, created_at, send_at, status, type, event_id, recipient, attempts, next_attempt_at, trace_carrier)
		values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb)`
	if n.NextAttemptAt.IsZero() {
		n.NextAttemptAt = time.Now().UTC()
	}
	nm, err := model.NotificationDomainToModel(n)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = r.conn.Exec(ctx, query,
		nm.Id, nm.Payload, nm.CreatedAt, nm.SendAt, nm.Status, nm.Type, nm.EventID,
		nm.Recipient, nm.Attempts, nm.NextAttemptAt, nm.TraceCarrier)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *notificationRepository) ListNotSent(ctx context.Context) ([]domain.Notification, error) {
	const op = "NotificationRepository.ListNotSent"
	const query = `
	select id, payload, created_at, send_at, status, type, event_id, recipient,
		attempts, lease_token, next_attempt_at, trace_carrier
	from notifier_schema.notifications
	where status not in ('sent','failed') and
		failed_at is null
	`
	rows, err := r.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.NotificationModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	items := make([]domain.Notification, 0, len(models))
	for _, item := range models {
		n, err := model.NotificationModelToDomain(item)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		items = append(items, n)
	}
	return items, nil
}

func (r *notificationRepository) Mark(
	ctx context.Context,
	n domain.Notification,
	status domain.NotificationStatus,
) error {
	_, err := r.conn.Exec(ctx, `update notifier_schema.notifications set status=$1 where id=$2`, status.String(), n.Id)
	return err
}

func (r *notificationRepository) ApplyEvent(
	ctx context.Context,
	eventID string,
	version int,
	carrier map[string]string,
	notifications []domain.Notification,
) (bool, error) {
	const op = "NotificationRepository.ApplyEvent"
	const query = `
	insert into notifier_schema.inbox_events(event_id,schema_version,trace_carrier,status)
		values($1,$2,$3::jsonb,'processing')
	on conflict(event_id) do nothing
	returning event_id
	`
	if eventID == "" || version < 1 {
		return false, fmt.Errorf("%s: invalid event identity/version", op)
	}
	carrierJSON, err := json.Marshal(carrier)
	if err != nil {
		return false, fmt.Errorf("%s: marshal trace carrier: %w", op, err)
	}
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, query, eventID, version, carrierJSON)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	inserted, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (string, error) { var id string; err := row.Scan(&id); return id, err })
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if len(inserted) == 0 {
		statusRows, err := tx.Query(ctx, `
			select status
			from notifier_schema.inbox_events
			where event_id=$1
			for update`, eventID)
		if err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
		statuses, err := pgx.CollectRows(statusRows, func(row pgx.CollectableRow) (string, error) {
			var status string
			err := row.Scan(&status)
			return status, err
		})
		if err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
		if len(statuses) != 1 {
			return false, fmt.Errorf("%s: inbox event disappeared", op)
		}
		if statuses[0] == "processed" || statuses[0] == "failed" {
			if err := tx.Commit(ctx); err != nil {
				return false, fmt.Errorf("%s: %w", op, err)
			}
			return true, nil
		}
	}
	for _, notification := range notifications {
		notification.EventID = eventID
		if notification.Recipient == "" {
			notification.Recipient = recipientFromPayload(notification)
		}
		if notification.Recipient == "" {
			return false, fmt.Errorf("%s: notification has no recipient", op)
		}
		if notification.TraceCarrier == nil {
			notification.TraceCarrier = carrier
		}
		nm, err := model.NotificationDomainToModel(notification)
		if err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
		_, err = tx.Exec(ctx, `
			insert into notifier_schema.notifications
				(id, payload, created_at, send_at, status, type, event_id, recipient, attempts, next_attempt_at, trace_carrier)
				values($1, $2, $3, $4, $5, $6, $7, $8, 0, now(), $9::jsonb)
			on conflict(event_id,recipient)
			where event_id is not null and
			recipient is not null
			do nothing`,
			nm.Id, nm.Payload, nm.CreatedAt, nm.SendAt, nm.Status, nm.Type, eventID, notification.Recipient, carrierJSON)
		if err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}
	_, err = tx.Exec(ctx, `
		update notifier_schema.inbox_events
		set status='processed',
			schema_version=$2,
			trace_carrier=$3::jsonb,
			completed_at=now(),
			failed_at=null,
			last_error=null
		where event_id=$1`, eventID, version, carrierJSON)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return false, nil
}

func recipientFromPayload(n domain.Notification) string {
	var payload struct {
		Email string   `json:"email"`
		Users []string `json:"users"`
	}
	if json.Unmarshal(n.Payload, &payload) != nil {
		return ""
	}
	if payload.Email != "" {
		return payload.Email
	}
	if len(payload.Users) == 1 {
		return payload.Users[0]
	}
	return ""
}

func (r *notificationRepository) RecordEventFailure(
	ctx context.Context,
	event domain.EventFailure,
	cause string,
	maxAttempts int,
) (bool, error) {
	const op = "NotificationRepository.RecordEventFailure"
	const query = `
	insert into notifier_schema.inbox_events
		(event_id, schema_version, trace_carrier, status, attempts, last_error, failed_at, source_topic, source_partition, source_offset, source_payload)
		values($1, $2,$3::jsonb, case when $5<=1 then 'failed' else 'processing' end, 1, left($4, 1000), case when $5<=1 then now() else null end, $6, $7, $8, $9)
	on conflict(event_id) do update set attempts=notifier_schema.inbox_events.attempts+1,
		last_error=left($4,1000),
		schema_version=$2,
		trace_carrier=$3::jsonb,
		source_topic=$6,
		source_partition=$7,
		source_offset=$8,
		source_payload=$9,
		status=case when notifier_schema.inbox_events.status='processed' then 'processed'
			when notifier_schema.inbox_events.status='failed' or notifier_schema.inbox_events.attempts+1 >= $5 then 'failed' else 'processing' end,
		failed_at=case when notifier_schema.inbox_events.status='processed' then null
			when notifier_schema.inbox_events.status='failed' or notifier_schema.inbox_events.attempts+1 >= $5 then now() else null end
	returning status
	`
	if maxAttempts < 1 {
		return false, fmt.Errorf("%s: max attempts must be positive", op)
	}
	if event.EventID == "" || event.Topic == "" {
		return false, fmt.Errorf("%s: event id and source topic are required", op)
	}
	carrierJSON, err := json.Marshal(event.Carrier)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	rows, err := r.conn.Query(ctx, query, event.EventID, event.SchemaVersion, carrierJSON, cause, maxAttempts, event.Topic, event.Partition, event.Offset, event.Payload)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	statuses, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (string, error) {
		var status string
		err := row.Scan(&status)
		return status, err
	})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if len(statuses) != 1 {
		return false, fmt.Errorf("%s: expected one inbox row", op)
	}
	return statuses[0] == "failed", nil
}

func (r *notificationRepository) ClaimDue(ctx context.Context, workerID string, limit int, lease time.Duration) ([]domain.Notification, error) {
	const op = "NotificationRepository.ClaimDue"
	if workerID == "" || limit < 1 || limit > 100 || lease <= 0 {
		return nil, fmt.Errorf("%s: invalid claim limits", op)
	}
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		update notifier_schema.notifications
		set status='failed',
			failed_at=now(),
			last_error=coalesce(last_error, 'lease expired after maximum attempts'),
			lease_token=null,
			lease_until=null
		where status not in ('sent','failed') and
			failed_at is null and
			lease_until < now() and
			attempts >= 8
		`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	rows, err := tx.Query(ctx, `
		with eligible as (
			select id from notifier_schema.notifications
			where status not in ('sent','failed') and
				failed_at is null and
				(status='urgent' or send_at <= now()) and
				next_attempt_at <= now() and
				(lease_until is null or lease_until < now()) and
				attempts < 8
			order by case when status='urgent' then 0 else 1 end, send_at, created_at, id
			for update skip locked limit $1
		), claimed as (
			update notifier_schema.notifications n
			set attempts=attempts+1,
				lease_token=gen_random_uuid(),
				lease_until=now()+($2*interval '1 millisecond')
			from eligible e
			where n.id=e.id
			returning n.id, n.payload, n.created_at, n.send_at, n.status, n.type, n.event_id, n.recipient, n.attempts, n.lease_token, n.next_attempt_at, n.trace_carrier
		)
		select id, payload, created_at, send_at, status, type, event_id, recipient, attempts, lease_token, next_attempt_at, trace_carrier
		from claimed
		`, limit, lease.Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.NotificationModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	items := make([]domain.Notification, 0, len(models))
	for _, item := range models {
		n, err := model.NotificationModelToDomain(item)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		items = append(items, n)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	stats, statsErr := r.conn.Query(ctx, `
		select count(*), greatest(coalesce(extract(epoch from now()-min(send_at)),0),0)
		from notifier_schema.notifications
		where status not in ('sent','failed') and
			failed_at is null
		`)
	if statsErr == nil {
		type queueStats struct {
			pending int64
			oldest  float64
		}
		if result, collectErr := pgx.CollectExactlyOneRow(stats, func(row pgx.CollectableRow) (queueStats, error) {
			var v queueStats
			e := row.Scan(&v.pending, &v.oldest)
			return v, e
		}); collectErr == nil {
			observability.SetQueue("notifier", "email", float64(result.pending), result.oldest)
		}
	}
	return items, nil
}

func (r *notificationRepository) Complete(ctx context.Context, id, leaseToken uuid.UUID) error {
	tag, err := r.conn.Exec(ctx, `
		update notifier_schema.notifications
		set status='sent',
			lease_token=null,
			lease_until=null,
			last_error=null
		where id=$1 and
			lease_token=$2 and
			status not in ('sent','failed')
		`, id, leaseToken)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return repository.ErrLeaseLost
	}
	return nil
}

func (r *notificationRepository) Retry(ctx context.Context, id, leaseToken uuid.UUID, nextAttempt time.Time, cause string, maxAttempts int) error {
	tag, err := r.conn.Exec(ctx, `
		update notifier_schema.notifications
		set status=case when attempts >= $5 then 'failed' else status end,
		failed_at=case when attempts >= $5 then now() else null end,
		next_attempt_at=case when attempts >= $5 then next_attempt_at else $3 end,
		last_error=left($4,1000),
		lease_token=null,
		lease_until=null
		where id=$1 and
			lease_token=$2 and
			status not in ('sent','failed')
		`, id, leaseToken, nextAttempt, cause, maxAttempts)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return repository.ErrLeaseLost
	}
	return nil
}

func translateError(err error) error {
	if err == nil {
		return nil
	}
	slog.Error("notification persistence error", "err", err)
	return err
}
