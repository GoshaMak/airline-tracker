-- +goose Up
alter table notifier_schema.notifications
    add column if not exists event_id text,
    add column if not exists recipient text,
    add column if not exists attempts integer not null default 0,
    add column if not exists next_attempt_at timestamptz not null default now(),
    add column if not exists lease_token uuid,
    add column if not exists lease_until timestamptz,
    add column if not exists last_error text,
    add column if not exists failed_at timestamptz,
    add column if not exists trace_carrier jsonb not null default '{}'::jsonb;

alter table notifier_schema.notifications drop constraint if exists notification_status_check;
alter table notifier_schema.notifications add constraint notification_status_check
    check (status in ('created', 'urgent', 'sent', 'failed'));

create unique index if not exists notifications_event_recipient_uidx
    on notifier_schema.notifications(event_id, recipient) where event_id is not null and recipient is not null;
create index if not exists notifications_due_idx
    on notifier_schema.notifications(send_at, next_attempt_at, created_at)
    where status not in ('sent', 'failed') and failed_at is null;

create table if not exists notifier_schema.inbox_events (
    event_id text primary key,
    schema_version integer not null default 1,
    trace_carrier jsonb not null default '{}'::jsonb,
    status varchar(20) not null default 'processing' check (status in ('processing', 'processed', 'failed')),
    attempts integer not null default 0,
    last_error text,
    created_at timestamptz not null default now(),
    completed_at timestamptz,
    failed_at timestamptz
);

grant select, insert, update, delete
on notifier_schema.inbox_events
to app_user
;

-- +goose Down
drop table if exists notifier_schema.inbox_events;
drop index if exists notifier_schema.notifications_due_idx;
drop index if exists notifier_schema.notifications_event_recipient_uidx;
alter table notifier_schema.notifications drop constraint if exists notification_status_check;
alter table notifier_schema.notifications add constraint notification_status_check
    check (status in ('created', 'urgent', 'sent'));
alter table notifier_schema.notifications
    drop column if exists trace_carrier,
    drop column if exists failed_at,
    drop column if exists last_error,
    drop column if exists lease_until,
    drop column if exists lease_token,
    drop column if exists next_attempt_at,
    drop column if exists attempts,
    drop column if exists recipient,
    drop column if exists event_id;
