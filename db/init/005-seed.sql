-- AIRCRAFT
insert into aircraft_models(manufacturer, model, mass, max_altitude, max_speed)
values ('A', 'B', 999, 3, 1);

insert into aircraft(aircraft_model_id, registration_number, serial_number, mileage)
values ((select id from aircraft_models), '999', '123', 123);

-- AIRPORT
insert into airports(iata_code, title, city, country)
values ('123', '123', '123', '123'),
       ('456', '456', '456', '456');

insert into gates(airport_id, number)
values ((select id from airports where title = '123'), 'A1'),
       ((select id from airports where title = '456'), 'B1');

-- FLIGHTS
select add_flight(
               gen_random_uuid(),
               '2026-03-26 14:30:00'::timestamp,
               '2026-05-26 14:30:00'::timestamp,
               'waiting',
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
      limit 1) ag;