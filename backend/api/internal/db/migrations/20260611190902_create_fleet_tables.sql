-- +goose Up
create table if not exists aircraft_models
(
    id           uuid primary key default gen_random_uuid(),
    manufacturer varchar(50) not null,
    model        varchar(50) not null,
    mass         int         not null, -- kg
    max_altitude int         not null, -- meters
    max_speed    int         not null, -- km/h

    constraint unique_model_per_manufacturer unique (manufacturer, model),
    constraint positive_mass check (mass > 0),
    constraint positive_max_altitude check (max_altitude > 0),
    constraint positive_max_speed check (max_speed > 0)
);

create table if not exists aircraft
(
    id                  uuid primary key default gen_random_uuid(),
    aircraft_model_id   uuid references aircraft_models (id) not null,
    registration_number varchar(10)                          not null unique,
    serial_number       varchar(10)                          not null unique,
    mileage             int                                  not null,

    constraint non_negative_mileage check (mileage >= 0)
);

-- +goose Down
drop table if exists aircraft, aircraft_models;