-- +goose Up
create table if not exists airports
(
    id        uuid primary key default gen_random_uuid(),
    iata_code varchar(10) unique not null,
    title     varchar(200)       not null,
    city_id   uuid               not null references cities (id)
);

create table if not exists gates
(
    id         uuid primary key default gen_random_uuid(),
    airport_id uuid references airports (id) not null,
    number     varchar(4)                    not null,

    constraint unique_gate_in_airport unique (airport_id, number)
);

-- +goose Down
drop table if exists gates, airports;