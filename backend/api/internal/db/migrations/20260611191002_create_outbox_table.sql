-- +goose Up
create table if not exists outbox
(
    id         uuid primary key   default gen_random_uuid(),
    topic      text      not null, -- TODO: mb change to type
    payload    jsonb     not null,
    created_at timestamp not null default now(),
    sent_at    timestamp
);

-- +goose Down
drop table if exists outbox;