-- +goose Up
-- +goose StatementBegin
create or replace function scan_flights_info()
    returns
        table
        (
            id                   uuid,
            aircraft_id          uuid,
            scheduled_departure  timestamp,
            scheduled_arrival    timestamp,
            actual_departure     timestamp,
            actual_arrival       timestamp,
            status               varchar(20),
            plan                 varchar(128),
            departure_gate_id    uuid,
            arrival_gate_id      uuid,
            departure_airport_id uuid,
            arrival_airport_id   uuid
        )
    language plpgsql
as
$$
begin
    return query
        select f.id,
               f.aircraft_id,
               f.scheduled_departure,
               f.scheduled_arrival,
               f.actual_departure,
               f.actual_arrival,
               f.status,
               f.plan,
               v.departure_gate_id,
               v.arrival_gate_id,
               gd.airport_id,
               ga.airport_id
        from flights f
                 join flight_routes v on f.id = v.flight_id
                 join gates gd on gd.id = v.departure_gate_id
                 join gates ga on ga.id = v.arrival_gate_id;
end;
$$;
-- +goose StatementEnd

-- +goose Down
drop function if exists scan_flights_info();