-- +goose Up
create role app_reader login password 'reader_pswd' nosuperuser nocreatedb nocreaterole
    noreplication nobypassrls;

grant connect on database airline_tracker to app_reader;
grant usage on schema api_schema to app_reader;
grant spectator to app_reader;

alter role app_reader set default_transaction_read_only = on;

-- +goose Down
revoke spectator from app_reader;
revoke usage on schema api_schema from app_reader;
revoke connect on database airline_tracker from app_reader;
drop role app_reader;
