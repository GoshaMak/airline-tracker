-- +goose Up
create table if not exists subscription_notification_timers
(
    id                    uuid primary key default gen_random_uuid(),
    subscription_id       uuid references subscriptions (id) on delete cascade not null,
    notify_before_minutes bigint not null,

    constraint notify_before_minutes_non_negative check (notify_before_minutes >= 0),
    constraint unique_notification_timer_per_subscription
        unique (subscription_id, notify_before_minutes)
);

grant all privileges on table subscription_notification_timers to manager;

-- +goose Down
revoke all privileges on table subscription_notification_timers from manager;
drop table if exists subscription_notification_timers;
