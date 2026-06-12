create role admin; -- superuser
create role dispatcher; -- controls flights
create role builder; -- controls airports, ...
create role designer; -- aircraft, ...
create role manager; -- controls users and subs
create role spectator; -- only selects data
create role app_user with login password 'app_pswd'; -- backend role