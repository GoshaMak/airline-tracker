create table if not exists users
(
    id            char(36) primary key default (uuid()),
    email         varchar(256) unique not null,
    password_hash varchar(128)        not null,
    role          varchar(20)         not null,

    constraint user_role_check check (role in ('user', 'admin'))
);

create table if not exists airports
(
    id        char(36) primary key default (uuid()),
    iata_code varchar(10) unique  not null,
    title     varchar(200) unique not null,
    city      varchar(200)        not null,
    country   varchar(10)         not null
);

create table if not exists gates
(
    id         char(36) primary key default (uuid()),
    airport_id char(36)                     not null,
    number     varchar(4)                   not null,

    constraint fk_gates_airport foreign key (airport_id) references airports (id),
    constraint unique_gate_in_airport unique (airport_id, number)
);

create table if not exists aircraft_models
(
    id           char(36) primary key default (uuid()),
    manufacturer varchar(50) not null,
    model        varchar(50) not null,
    mass         int         not null,
    max_altitude int         not null,
    max_speed    int         not null,

    constraint unique_model_per_manufacturer unique (manufacturer, model),
    constraint positive_mass check (mass > 0),
    constraint positive_max_altitude check (max_altitude > 0),
    constraint positive_max_speed check (max_speed > 0)
);

create table if not exists aircraft
(
    id                  char(36) primary key default (uuid()),
    aircraft_model_id   char(36)                            not null,
    registration_number varchar(10)                         not null unique,
    serial_number       varchar(10)                         not null unique,
    mileage             int                                 not null,

    constraint fk_aircraft_model foreign key (aircraft_model_id) references aircraft_models (id),
    constraint non_negative_mileage check (mileage >= 0)
);

create table if not exists flights
(
    id                  char(36) primary key default (uuid()),
    aircraft_id         char(36)                     not null,
    scheduled_departure datetime(6)                  not null,
    scheduled_arrival   datetime(6)                  not null,
    actual_departure    datetime(6)     default null,
    actual_arrival      datetime(6)     default null,
    status              varchar(20)                  not null,
    plan                varchar(200),

    constraint fk_flights_aircraft foreign key (aircraft_id) references aircraft (id),
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
    id                char(36) primary key default (uuid()),
    flight_id         char(36) not null,
    departure_gate_id char(36) not null,
    arrival_gate_id   char(36) not null,

    constraint fk_flight_routes_flight foreign key (flight_id) references flights (id),
    constraint fk_flight_routes_departure_gate foreign key (departure_gate_id) references gates (id),
    constraint fk_flight_routes_arrival_gate foreign key (arrival_gate_id) references gates (id)
);

create table if not exists subscriptions
(
    id        char(36) primary key default (uuid()),
    user_id   char(36) not null,
    flight_id char(36) not null,

    constraint fk_subscriptions_user foreign key (user_id) references users (id),
    constraint fk_subscriptions_flight foreign key (flight_id) references flights (id),
    constraint unique_flight_subscription_per_user unique (user_id, flight_id)
);

create table if not exists outbox
(
    id         char(36) primary key default (uuid()),
    topic      text        not null,
    payload    json        not null,
    created_at datetime(6) not null default (current_timestamp(6)),
    sent_at    datetime(6)
);

create table if not exists notifications
(
    id         char(36) primary key default (uuid()),
    payload    json        not null,
    created_at datetime(6) not null default (current_timestamp(6)),
    send_at    datetime(6) not null,
    status     varchar(20) not null,
    type       varchar(20) not null,

    constraint notification_status_check check (status in ('created', 'urgent', 'sent')),
    constraint notification_type_check check (type in ('subscribed', 'flight_updated'))
);
