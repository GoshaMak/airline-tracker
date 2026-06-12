-- +goose Up
grant all privileges on table notifications to app_user;

-- +goose Down
revoke all privileges on table notifications from app_user;
