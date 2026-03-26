create or replace procedure add_flight(
    p_flight_id uuid,
    p_scheduled_departure timestamp,
    p_scheduled_arrival timestamp,
    p_flight_status varchar,
    p_flight_plan varchar,
    p_aircraft_id uuid,
    p_departure_airport_id uuid,
    p_arrival_airport_id uuid,
    p_departure_gate_id uuid,
    p_arrival_gate_id uuid
)
    language plpgsql
as
$$
begin
    if p_scheduled_arrival <= p_scheduled_departure then
        raise exception 'can not land before departing';
    end if;

    if not exists(select 1 from aircraft where id = p_aircraft_id) then
        raise exception 'aircraft does not exist';
    end if;

    if p_departure_airport_id = p_arrival_airport_id then
        raise exception 'airports must be different';
    end if;

    if not exists(select 1 from airports where id = p_departure_airport_id) then
        raise exception 'departure airport does not exist';
    end if;

    if not exists(select 1 from airports where id = p_arrival_airport_id) then
        raise exception 'arrival airport does not exist';
    end if;

    if not exists(select 1 from gates where id = p_departure_gate_id and airport_id = p_departure_airport_id) then
        raise exception 'departure gate does not exist';
    end if;

    if not exists(select 1 from gates where id = p_arrival_gate_id and airport_id = p_arrival_airport_id) then
        raise exception 'arrival gate does not exist';
    end if;

    insert into flights (id,
                         aircraft_id,
                         scheduled_departure,
                         scheduled_arrival,
                         status,
                         flight_plan)
    values (p_flight_id,
            p_aircraft_id,
            p_scheduled_departure,
            p_scheduled_arrival,
            p_flight_status,
            p_flight_plan);

    insert into visits (flight_id, departure_gate_id, arrival_gate_id)
    values (p_flight_id, p_departure_gate_id, p_arrival_gate_id);
end;
$$