-- +goose Up
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

-- +goose Down
drop table if exists flight_routes, flights;