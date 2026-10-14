-- +goose Up
grant all privileges on schema api_schema to admin;
grant all privileges on all tables in schema api_schema to admin;
alter default privileges in schema api_schema grant all privileges on tables to admin;

grant all privileges on table api_schema.flights, api_schema.flight_routes to dispatcher;

grant all privileges on table api_schema.airports, api_schema.gates, api_schema.countries, api_schema.cities to builder;

grant all privileges on table api_schema.aircraft_models, api_schema.aircraft to designer;

grant all privileges on table api_schema.users, api_schema.subscriptions, api_schema.outbox to manager;

grant usage on schema api_schema to spectator;
grant select on all tables in schema api_schema to spectator;
alter default privileges in schema api_schema grant select on tables to spectator;

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

alter default privileges in schema api_schema revoke select on tables from spectator;
revoke select on all tables in schema api_schema from spectator;
revoke usage on schema api_schema from spectator;

revoke all privileges on table api_schema.users, api_schema.subscriptions, api_schema.outbox from manager;

revoke all privileges on table api_schema.aircraft_models, api_schema.aircraft from designer;

revoke all privileges on table api_schema.airports, api_schema.gates, api_schema.countries, api_schema.cities from builder;

revoke all privileges on table api_schema.flights, api_schema.flight_routes from dispatcher;

alter default privileges in schema api_schema revoke all privileges on tables from admin;
revoke all privileges on all tables in schema api_schema from admin;
revoke all privileges on schema api_schema from admin;
