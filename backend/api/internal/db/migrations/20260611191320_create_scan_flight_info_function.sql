-- +goose Up
-- +goose StatementBegin
create or replace function scan_flight_info(p_flight_id uuid)
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
               fr.departure_gate_id,
               fr.arrival_gate_id,
               gd.airport_id,
               ga.airport_id
        from flights f
                 join flight_routes fr on f.id = fr.flight_id
                 join gates gd on gd.id = fr.departure_gate_id
                 join gates ga on ga.id = fr.arrival_gate_id
        where f.id = p_flight_id;
end;
$$;
-- +goose StatementEnd

-- +goose Down
drop function if exists scan_flight_info(p_flight_id uuid);