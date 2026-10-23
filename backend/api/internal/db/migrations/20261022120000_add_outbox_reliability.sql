-- +goose Up
alter table api_schema.outbox
    add column if not exists attempts integer not null default 0,
    add column if not exists next_attempt_at timestamptz not null default now(),
    add column if not exists lease_token uuid,
    add column if not exists lease_until timestamptz,
    add column if not exists last_error text,
    add column if not exists failed_at timestamptz,
    add column if not exists trace_carrier jsonb not null default '{}'::jsonb,
    add column if not exists correlation_id text;

create index if not exists outbox_ready_idx
    on api_schema.outbox (next_attempt_at, created_at)
    where sent_at is null and failed_at is null;

-- +goose Down
drop index if exists api_schema.outbox_ready_idx;
alter table api_schema.outbox
    drop column if exists correlation_id,
    drop column if exists trace_carrier,
    drop column if exists failed_at,
    drop column if exists last_error,
    drop column if exists lease_until,
    drop column if exists lease_token,
    drop column if exists next_attempt_at,
    drop column if exists attempts;
