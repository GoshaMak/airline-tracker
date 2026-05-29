drop procedure if exists add_flight;
drop procedure if exists scan_flight_info;
drop procedure if exists scan_flights_info;
drop procedure if exists subscribe;
drop procedure if exists scan_user_flights_info;

delimiter //

create procedure add_flight(
    in p_flight_id char(36),
    in p_scheduled_departure datetime(6),
    in p_scheduled_arrival datetime(6),
    in p_flight_status varchar(20),
    in p_flight_plan varchar(200),
    in p_aircraft_id char(36),
    in p_departure_airport_id char(36),
    in p_arrival_airport_id char(36),
    in p_departure_gate_id char(36),
    in p_arrival_gate_id char(36)
)
begin
    declare exit handler for sqlexception
    begin
        rollback;
        resignal;
    end;

    if p_scheduled_arrival <= p_scheduled_departure then
        signal sqlstate '45000' set message_text = 'can not land before departing';
    end if;

    if not exists(select 1 from aircraft where id = p_aircraft_id) then
        signal sqlstate '45000' set message_text = 'aircraft does not exist';
    end if;

    if p_departure_airport_id = p_arrival_airport_id then
        signal sqlstate '45000' set message_text = 'airports must be different';
    end if;

    if not exists(select 1 from airports where id = p_departure_airport_id) then
        signal sqlstate '45000' set message_text = 'departure airport does not exist';
    end if;

    if not exists(select 1 from airports where id = p_arrival_airport_id) then
        signal sqlstate '45000' set message_text = 'arrival airport does not exist';
    end if;

    if not exists(select 1 from gates where id = p_departure_gate_id and airport_id = p_departure_airport_id) then
        signal sqlstate '45000' set message_text = 'departure gate does not exist';
    end if;

    if not exists(select 1 from gates where id = p_arrival_gate_id and airport_id = p_arrival_airport_id) then
        signal sqlstate '45000' set message_text = 'arrival gate does not exist';
    end if;

    start transaction;

    insert into flights (id,
                         aircraft_id,
                         scheduled_departure,
                         scheduled_arrival,
                         status,
                         plan)
    values (p_flight_id,
            p_aircraft_id,
            p_scheduled_departure,
            p_scheduled_arrival,
            p_flight_status,
            p_flight_plan);

    insert into flight_routes (flight_id, departure_gate_id, arrival_gate_id)
    values (p_flight_id, p_departure_gate_id, p_arrival_gate_id);

    commit;
end//

create procedure scan_flight_info(
    in p_flight_id char(36)
)
begin
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
           gd.airport_id as departure_airport_id,
           ga.airport_id as arrival_airport_id
    from flights f
             join flight_routes fr on f.id = fr.flight_id
             join gates gd on gd.id = fr.departure_gate_id
             join gates ga on ga.id = fr.arrival_gate_id
    where f.id = p_flight_id;
end//

create procedure scan_flights_info()
begin
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
           gd.airport_id as departure_airport_id,
           ga.airport_id as arrival_airport_id
    from flights f
             join flight_routes v on f.id = v.flight_id
             join gates gd on gd.id = v.departure_gate_id
             join gates ga on ga.id = v.arrival_gate_id;
end//

create procedure subscribe(
    in p_uid char(36),
    in p_fid char(36)
)
begin
    if not exists(select 1 from users where id = p_uid) then
        signal sqlstate '45000' set message_text = 'user not found';
    end if;

    if not exists(select 1 from flights where id = p_fid) then
        signal sqlstate '45000' set message_text = 'flight not found';
    end if;

    insert into subscriptions(user_id, flight_id)
    values (p_uid, p_fid);
end//

create procedure scan_user_flights_info(
    in p_uid char(36)
)
begin
    if not exists(select 1 from users u where u.id = p_uid) then
        signal sqlstate '45000' set message_text = 'user not found';
    end if;

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
           gd.airport_id as departure_airport_id,
           ga.airport_id as arrival_airport_id
    from flights f
             join flight_routes v on f.id = v.flight_id
             join gates gd on gd.id = v.departure_gate_id
             join gates ga on ga.id = v.arrival_gate_id
             join subscriptions s on s.flight_id = f.id and s.user_id = p_uid;
end//

delimiter ;
