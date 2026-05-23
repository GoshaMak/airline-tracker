create role admin; -- superuser
create role dispatcher; -- controls flights
create role builder; -- controls airports, ...
create role designer; -- aircraft, ...
create role manager; -- controls users and subs
create role spectator; -- only selects data
create role publisher; -- manages outbox (i.e. controls kafka)
create role notifier; -- manages notifications
create role app_user with login password 'app_pswd'; -- backend role

grant all privileges on schema public to admin;
grant all privileges on all tables in schema public to admin;

grant all privileges on table flights, flight_routes to dispatcher;

grant all privileges on table airports, gates to builder;

grant all privileges on table aircraft_models, aircraft to designer;

grant all privileges on table users, subscriptions to manager;

grant select on all tables in schema public to spectator;

grant all privileges on table outbox to publisher;

grant all privileges on table notifications to notifier;

grant manager to app_user;
grant spectator to app_user;
grant publisher to app_user;
grant notifier to app_user;