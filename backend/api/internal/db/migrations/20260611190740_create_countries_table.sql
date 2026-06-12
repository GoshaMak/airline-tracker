-- +goose Up
create table if not exists countries
(
    id   uuid primary key default gen_random_uuid(),
    code varchar(5) unique not null,
    name varchar(100)      not null
);

-- +goose Down
drop table if exists countries;