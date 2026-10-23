-- +goose Up
alter table notifier_schema.inbox_events
    add column if not exists source_topic text,
    add column if not exists source_partition integer,
    add column if not exists source_offset bigint,
    add column if not exists source_payload bytea;

-- +goose Down
alter table notifier_schema.inbox_events
    drop column if exists source_payload,
    drop column if exists source_offset,
    drop column if exists source_partition,
    drop column if exists source_topic;
