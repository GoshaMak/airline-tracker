create table if not exists users
(
    id       serial primary key,
    email    varchar(256) unique, -- TODO: check format
    phone    varchar(32) unique,-- TODO: check format
    password varchar(128) not null,
    role     varchar(20)
);

create table if not exists airports
(
    id        serial primary key,
    iata_code varchar(10) unique not null,
    title     varchar(128),
    city      varchar(128),
    country   varchar(128)
);

create table if not exists gates
(
    id         serial primary key,
    airport_id bigint references airports (id),
    number     varchar(128)
);

create table if not exists aircraft_models
(
    id           serial primary key,
    manufacturer varchar(50),
    model        varchar(50),
    mass         int, -- kg
    max_altitude int, -- meters
    max_speed    int  -- km/h
);

create table if not exists aircraft
(
    id                  serial primary key,
    aircraft_model_id   bigint references aircraft_models (id),
    registration_number varchar(10) unique,
    serial_number       varchar(10),
    mileage             int
);

create table if not exists flights
(
    id                  serial primary key,
    aircraft_id         bigint references aircraft (id),
    scheduled_departure timestamp,
    scheduled_arrival   timestamp,
    actual_departure    timestamp,
    actual_arrival      timestamp,
    status              varchar(20),
    flight_plan         varchar(128)
);

create table if not exists visits
(
    id                serial primary key,
    flight_id         bigint references flights (id),
    departure_gate_id bigint references gates (id),
    arrival_gate_id   bigint references gates (id)
);