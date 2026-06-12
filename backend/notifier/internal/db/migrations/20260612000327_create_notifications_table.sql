-- +goose Up
create table if not exists notifications
(
    id         uuid primary key     default gen_random_uuid(),
    payload    jsonb       not null,
    created_at timestamp   not null default now(),
    send_at    timestamp   not null,
    status     varchar(20) not null,
    type       varchar(20) not null,

    constraint notification_status_check check (status in ('created', 'urgent', 'sent')),
    constraint notification_type_check check (type in ('subscribed', 'flight_updated'))
);

-- +goose Down
drop table if exists notifications;
