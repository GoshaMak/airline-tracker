-- +goose Up
create table if not exists subscriptions
(
    id        uuid primary key default gen_random_uuid(),
    user_id   uuid references users (id)   not null,
    flight_id uuid references flights (id) not null,

    constraint unique_flight_subscription_per_user unique (user_id, flight_id)
);

-- +goose Down
drop table if exists subscriptions;