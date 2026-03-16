create or replace function add_flight(
    p_scheduled_departure timestamp,
    p_scheduled_arrival timestamp,
    p_flight_status varchar,
    p_flight_plan varchar,
    p_registration_number varchar,
    p_departure_iata_code varchar,
    p_arrival_iata_code varchar,
    p_departure_gate_number varchar,
    p_arrival_gate_number varchar
)
    returns bigint
    language plpgsql
as
$$
declare
    v_aircraft_id          bigint;
    v_departure_airport_id bigint;
    v_arrival_airport_id   bigint;
    v_departure_gate_id    bigint;
    v_arrival_gate_id      bigint;
    v_flight_id            bigint;
begin
    if p_departure_iata_code = p_arrival_iata_code then
        raise exception 'airports must be different';
    end if;

    if p_scheduled_arrival <= p_scheduled_departure then
        raise exception 'can not land before departing';
    end if;

    if not exists(select 1 from aircraft where registration_number = p_registration_number) then
        raise exception 'airport does not exist';
    end if;

    v_aircraft_id := (select id from aircraft where registration_number = p_registration_number);

    if not exists(select 1 from airports where iata_code = p_departure_iata_code) then
        raise exception 'departure airport does not exist';
    end if;

    if not exists(select 1 from airports where iata_code = p_arrival_iata_code) then
        raise exception 'departure airport does not exist';
    end if;

    v_departure_airport_id := (select id from airports where iata_code = p_departure_iata_code);
    if not exists(select 1
                  from gates
                  where airport_id = v_departure_airport_id
                    and number = p_departure_gate_number) then
        raise exception 'departure gate does not exist';
    end if;
    v_departure_gate_id :=
            (select id from gates where airport_id = v_departure_airport_id and number = p_departure_gate_number);

    v_arrival_airport_id := (select id from airports where iata_code = p_arrival_iata_code);
    if not exists(select 1
                  from gates
                  where airport_id = v_arrival_airport_id
                    and number = p_arrival_gate_number) then
        raise exception 'arrival gate does not exist';
    end if;
    v_arrival_gate_id :=
            (select id from gates where airport_id = v_arrival_airport_id and number = p_arrival_gate_number);

    insert into flights (aircraft_id,
                         scheduled_departure,
                         scheduled_arrival,
                         status,
                         flight_plan)
    values (v_aircraft_id,
            p_scheduled_departure,
            p_scheduled_arrival,
            p_flight_status,
            p_flight_plan)
    returning id into v_flight_id;

    insert into visits (flight_id, departure_gate_id, arrival_gate_id)
    values (v_flight_id, v_departure_gate_id, v_arrival_gate_id);

    return v_flight_id;
end;
$$