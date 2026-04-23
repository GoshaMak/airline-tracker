-- AIRCRAFT
insert into aircraft_models(manufacturer, model, mass, max_altitude, max_speed)
values ('A', 'B', 999, 3, 1);

insert into aircraft(aircraft_model_id, registration_number, serial_number, mileage)
values ((select id from aircraft_models), '999', '123', 123);

-- AIRPORT
insert into airports(iata_code, title, city, country)
values ('SVO', 'Sheremetyevo International Airport', 'Moscow', 'RU'),
       ('LED', 'Pulkovo Airport', 'St. Petersburg', 'RU');

insert into gates(airport_id, number)
values ((select id from airports where title = 'Sheremetyevo International Airport'), 'A1'),
       ((select id from airports where title = 'Pulkovo Airport'), 'B1');

-- USERS
insert into users(id, email, password, role)
values ('d60a8f31-f314-4e98-81e9-bc67085f6d20', 'ab@cd.ef',
        '$2a$10$XdFK03wD8YLIQMKc/lklJOfNPos2tsUeUj2nilqGzln7M3cqnz9uS', 'user');

insert into users(id, email, password, role)
values ('81008edc-1361-4500-a131-5959d7dad3e1', 'a@b.c',
        '$2a$10$muwnpf/NFUp.NIKa2IAdQO6PjZcoL5nueNIMko82Z.fzSafubyMby', 'admin');

-- FLIGHTS
select add_flight(
               gen_random_uuid(),
               '2026-03-26 14:30:00'::timestamp,
               '2026-05-26 14:30:00'::timestamp,
               'scheduled',
               'ABCDE A99 ABCDE ABC A66',
               a.id,
               da.id,
               aa.id,
               dg.id,
               ag.id
       )
from (select id from aircraft limit 1) a,
     (select id from airports where title = '123') da,
     (select id from airports where title = '456') aa,
     (select g.id
      from gates g
               join airports a on g.airport_id = a.id
      where a.title = '123'
      limit 1) dg,
     (select g.id
      from gates g
               join airports a on g.airport_id = a.id
      where a.title = '456'
      limit 1) ag
;