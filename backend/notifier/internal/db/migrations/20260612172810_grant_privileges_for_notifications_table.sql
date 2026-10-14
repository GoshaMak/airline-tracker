-- +goose Up
grant all privileges on all tables in schema notifier_schema to app_user;
alter default privileges in schema notifier_schema grant all privileges on tables to app_user;

-- +goose Down
alter default privileges in schema notifier_schema revoke all privileges on tables from app_user;
revoke all privileges on all tables in schema notifier_schema from app_user;
