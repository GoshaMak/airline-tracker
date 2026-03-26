create table if not exists users
(
    id       uuid primary key default gen_random_uuid(),
    email    varchar(256) unique, -- TODO: check format
    phone    varchar(32) unique,-- TODO: check format
    password varchar(128) not null,
    role     varchar(20)
);

create table if not exists airports
(
    id        uuid primary key default gen_random_uuid(),
    iata_code varchar(10) unique not null,
    title     varchar(128),
    city      varchar(128),
    country   varchar(128)
);

create table if not exists gates
(
    id         uuid primary key default gen_random_uuid(),
    airport_id uuid references airports (id),
    number     varchar(128)
);

create table if not exists aircraft_models
(
    id           uuid primary key default gen_random_uuid(),
    manufacturer varchar(50),
    model        varchar(50),
    mass         int, -- kg
    max_altitude int, -- meters
    max_speed    int  -- km/h
);

create table if not exists aircraft
(
    id                  uuid primary key default gen_random_uuid(),
    aircraft_model_id   uuid references aircraft_models (id),
    registration_number varchar(10) unique,
    serial_number       varchar(10),
    mileage             int
);

create table if not exists flights
(
    id                  uuid primary key default gen_random_uuid(),
    aircraft_id         uuid references aircraft (id),
    scheduled_departure timestamp,
    scheduled_arrival   timestamp,
    actual_departure    timestamp,
    actual_arrival      timestamp,
    status              varchar(20),
    flight_plan         varchar(128)
);

create table if not exists visits
(
    id                uuid primary key default gen_random_uuid(),
    flight_id         uuid references flights (id),
    departure_gate_id uuid references gates (id),
    arrival_gate_id   uuid references gates (id)
);