create table if not exists users
(
    id       uuid primary key default gen_random_uuid(),
    email    varchar(256) unique, -- TODO: check format
    password varchar(128) not null,
    role     varchar(20)  not null
);

create table if not exists airports
(
    id        uuid primary key default gen_random_uuid(),
    iata_code varchar(10) unique not null,
    title     varchar(128)       not null,
    city      varchar(128)       not null,
    country   varchar(128)       not null
);

create table if not exists gates
(
    id         uuid primary key default gen_random_uuid(),
    airport_id uuid references airports (id) not null,
    number     varchar(128)                  not null
);

create table if not exists aircraft_models
(
    id           uuid primary key default gen_random_uuid(),
    manufacturer varchar(50) not null,
    model        varchar(50) not null,
    mass         int         not null, -- kg
    max_altitude int         not null, -- meters
    max_speed    int         not null  -- km/h
);

create table if not exists aircraft
(
    id                  uuid primary key default gen_random_uuid(),
    aircraft_model_id   uuid references aircraft_models (id) not null,
    registration_number varchar(10)                          not null unique,
    serial_number       varchar(10)                          not null,
    mileage             int                                  not null
);

create table if not exists flights
(
    id                  uuid primary key default gen_random_uuid(),
    aircraft_id         uuid references aircraft (id) not null,
    scheduled_departure timestamp                     not null,
    scheduled_arrival   timestamp                     not null,
    actual_departure    timestamp        default null,
    actual_arrival      timestamp        default null,
    status              varchar(20)                   not null,
    plan                varchar(128)
);

create table if not exists visits
(
    id                uuid primary key default gen_random_uuid(),
    flight_id         uuid references flights (id) not null,
    departure_gate_id uuid references gates (id)   not null,
    arrival_gate_id   uuid references gates (id)   not null
);

create table if not exists subscriptions
(
    id        uuid primary key default gen_random_uuid(),
    user_id   uuid references users (id)   not null,
    flight_id uuid references flights (id) not null
);