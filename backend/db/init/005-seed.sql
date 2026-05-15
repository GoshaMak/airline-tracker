-- -- AIRCRAFT
-- insert into aircraft_models(manufacturer, model, mass, max_altitude, max_speed)
-- values ('Boeing', 'Boeing 737/737 MAX', 500_000, 10_000, 5_000);
--
-- insert into aircraft(aircraft_model_id, registration_number, serial_number, mileage)
-- values ((select id from aircraft_models), 'N12345', 'C-FBBB', 15_000);
--
-- -- AIRPORT
-- insert into airports(iata_code, title, city, country)
-- values ('SVO', 'Sheremetyevo International Airport', 'Moscow', 'RU'),
--        ('LED', 'Pulkovo Airport', 'St. Petersburg', 'RU');
--
-- insert into gates(airport_id, number)
-- values ((select id from airports where title = 'Sheremetyevo International Airport'), 'A1'),
--        ((select id from airports where title = 'Pulkovo Airport'), 'B1');
--
-- USERS
-- pswd: myStrong123
insert into users(id, email, password_hash, role)
values ('39f2f5f8-5ec3-4436-be16-341f5ef4771f', 'ab@cd.ef',
        '$2a$10$KMSc6YuqxohcaO1Zo7Cs7eSssudoNr6jzvrTCdVCWbR4ht8.9RyyK', 'user');

insert into users(id, email, password_hash, role)
values ('1caa1bd8-63f1-4cd4-adda-fb9a054407cb', 'a@b.c',
        '$2a$10$Gxo54TNRn/9qg0yoO03Csuw9asc1Htm.NuvBr3g/oWsFfL09ViQr.', 'admin');
--
-- -- FLIGHTS
-- select add_flight(
--                gen_random_uuid(),
--                '2026-03-26 14:30:00'::timestamp,
--                '2026-05-26 14:30:00'::timestamp,
--                'scheduled',
--                'ABCDE A99 ABCDE ABC A66',
--                a.id,
--                da.id,
--                aa.id,
--                dg.id,
--                ag.id
--        )
-- from (select id from aircraft limit 1) a,
--      (select id from airports where title = 'Sheremetyevo International Airport') da,
--      (select id from airports where title = 'Pulkovo Airport') aa,
--      (select g.id
--       from gates g
--                join airports a on g.airport_id = a.id
--       where a.title = 'Sheremetyevo International Airport'
--       limit 1) dg,
--      (select g.id
--       from gates g
--                join airports a on g.airport_id = a.id
--       where a.title = 'Pulkovo Airport'
--       limit 1) ag
-- ;