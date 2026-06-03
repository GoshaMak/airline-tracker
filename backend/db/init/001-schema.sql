create table if not exists users
(
    id            uuid primary key default gen_random_uuid(),
    email         varchar(256) unique not null,
    password_hash varchar(128)        not null,
    role          varchar(20)         not null,

    constraint user_role_check check (role in ('user', 'admin'))
);

create table if not exists airports
(
    id        uuid primary key default gen_random_uuid(),
    iata_code varchar(10) unique  not null,
    title     varchar(200) unique not null,
    city      varchar(200)        not null,
    country   varchar(10)         not null
);

create table if not exists gates
(
    id         uuid primary key default gen_random_uuid(),
    airport_id uuid references airports (id) not null,
    number     varchar(4)                    not null,

    constraint unique_gate_in_airport unique (airport_id, number)
);

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

create table if not exists flights
(
    id                  uuid primary key default gen_random_uuid(),
    aircraft_id         uuid references aircraft (id) not null,
    scheduled_departure timestamp                     not null,
    scheduled_arrival   timestamp                     not null,
    actual_departure    timestamp        default null,
    actual_arrival      timestamp        default null,
    status              varchar(20)                   not null,
    plan                varchar(200),

    constraint sch_arrival_after_sch_departure check (scheduled_arrival > scheduled_departure),
    constraint act_arrival_after_act_departure check (actual_departure is null or actual_arrival is null or
                                                      actual_arrival > actual_departure),
    constraint status_check check (status in ('scheduled', 'boarding',
                                              'departed', 'landed',
                                              'arrived', 'delayed',
                                              'cancelled', 'rescheduled'))
);

create table if not exists flight_routes
(
    id                uuid primary key default gen_random_uuid(),
    flight_id         uuid references flights (id) not null, -- TODO: перенести в flights
    departure_gate_id uuid references gates (id)   not null,
    arrival_gate_id   uuid references gates (id)   not null
);

create table if not exists subscriptions
(
    id        uuid primary key default gen_random_uuid(),
    user_id   uuid references users (id)   not null,
    flight_id uuid references flights (id) not null,
    -- TODO: delay timestamp,

    constraint unique_flight_subscription_per_user unique (user_id, flight_id)
);

create table if not exists outbox
(
    id         uuid primary key   default gen_random_uuid(),
    topic      text      not null, -- TODO: поменять на type
    payload    jsonb     not null,
    created_at timestamp not null default now(),
    sent_at    timestamp
);

create table if not exists notifications
(
    id         uuid primary key     default gen_random_uuid(),
    payload    jsonb       not null,
    created_at timestamp   not null default now(),
    send_at    timestamp   not null,
    status     varchar(20) not null,
    type       varchar(20) not null,

    constraint notification_status_check check (status in ('created', 'urgent', 'sent')),
    constraint notification_type_check check (type in ('subscribed', 'flight_updated'))
);