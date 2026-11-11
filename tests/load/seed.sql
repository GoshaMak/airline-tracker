-- The isolated stand supplies the reference airports, gates, aircraft and administrator.
-- One passenger and flight per VU: no growing subscriber fan-out between iterations.
begin;
set search_path = api_schema, public;

insert into users (id, email, password_hash, role)
select md5('load-user-' || n)::uuid,
        'load-' || n || '@airline.test',
        password.hash,
        'user'
from generate_series(1, 10000) as n
cross join (select crypt('Passenger123!', '$2a$10$abcdefghijklmnopqrstuu') as hash offset 0) as password
on conflict do nothing;

insert into flights (id, aircraft_id, scheduled_departure, scheduled_arrival, status, plan)
select md5('load-flight-' || n)::uuid,
        '00000000-0000-4000-8000-000000000041',
        '2099-01-15 10:00:00',
        '2099-01-15 11:30:00',
        'scheduled',
        'LOAD ' || n
from generate_series(1, 10000) as n
on conflict do nothing;

insert into flight_routes (id, flight_id, departure_gate_id, arrival_gate_id)
select md5('load-route-' || n)::uuid,
        md5('load-flight-' || n)::uuid,
        '00000000-0000-4000-8000-000000000031',
        '00000000-0000-4000-8000-000000000032'
from generate_series(1, 10000) as n
on conflict do nothing;
commit;
