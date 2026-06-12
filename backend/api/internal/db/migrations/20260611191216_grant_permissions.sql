-- +goose Up
create role admin; -- superuser
create role dispatcher; -- controls flights
create role builder; -- controls airports, ...
create role designer; -- aircraft, ...
create role manager; -- controls users and subs
create role spectator; -- only selects data
create role app_user with login password 'app_pswd'; -- backend role

grant all privileges on schema public to admin;
grant all privileges on all tables in schema public to admin;

grant all privileges on table flights, flight_routes to dispatcher;

grant all privileges on table airports, gates to builder;

grant all privileges on table aircraft_models, aircraft to designer;

grant all privileges on table users, subscriptions, outbox to manager;

grant select on all tables in schema public to spectator;

grant dispatcher to app_user;
grant builder to app_user;
grant designer to app_user;
grant manager to app_user;
grant spectator to app_user;

-- +goose Down
revoke spectator from app_user;
revoke manager from app_user;
revoke designer from app_user;
revoke builder from app_user;
revoke dispatcher from app_user;

revoke select on all tables in schema public from spectator;

revoke all privileges on table users, subscriptions, outbox from manager;

revoke all privileges on table aircraft_models, aircraft from designer;

revoke all privileges on table airports, gates from builder;

revoke all privileges on table flights, flight_routes from dispatcher;

revoke all privileges on all tables in schema public from admin;

drop role if exists app_user;
drop role if exists spectator;
drop role if exists manager;
drop role if exists designer;
drop role if exists builder;
drop role if exists dispatcher;
drop role if exists admin;