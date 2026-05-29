create role if not exists 'admin';
create role if not exists 'dispatcher';
create role if not exists 'builder';
create role if not exists 'designer';
create role if not exists 'manager';
create role if not exists 'spectator';
create role if not exists 'publisher';
create role if not exists 'notifier';

create user if not exists 'app_user'@'%' identified by 'app_pswd';

grant all privileges on airline_tracker.* to 'admin';

grant all privileges on airline_tracker.flights to 'dispatcher';
grant all privileges on airline_tracker.flight_routes to 'dispatcher';

grant all privileges on airline_tracker.airports to 'builder';
grant all privileges on airline_tracker.gates to 'builder';

grant all privileges on airline_tracker.aircraft_models to 'designer';
grant all privileges on airline_tracker.aircraft to 'designer';

grant all privileges on airline_tracker.users to 'manager';
grant all privileges on airline_tracker.subscriptions to 'manager';

grant select on airline_tracker.* to 'spectator';

grant all privileges on airline_tracker.outbox to 'publisher';

grant all privileges on airline_tracker.notifications to 'notifier';

grant 'dispatcher' to 'app_user'@'%';
grant 'manager' to 'app_user'@'%';
grant 'spectator' to 'app_user'@'%';
grant 'publisher' to 'app_user'@'%';
grant 'notifier' to 'app_user'@'%';

set default role all to 'app_user'@'%';
