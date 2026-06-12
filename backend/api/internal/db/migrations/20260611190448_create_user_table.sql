-- +goose Up
create table if not exists users
(
    id            uuid primary key default gen_random_uuid(),
    email         varchar(256) unique not null,
    password_hash varchar(128)        not null,
    role          varchar(20)         not null,

    constraint user_role_check check (role in ('user', 'admin'))
);

-- +goose Down
drop table if exists users;