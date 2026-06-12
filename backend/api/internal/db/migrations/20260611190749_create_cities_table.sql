-- +goose Up
create table if not exists cities
(
    id         uuid primary key default gen_random_uuid(),
    country_id uuid         not null references countries (id),
    name       varchar(200) not null,

    unique (name, country_id)
);

-- +goose Down
drop table if exists cities;