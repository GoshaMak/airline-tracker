create role app_reader login password 'reader_pswd' nosuperuser nocreatedb nocreaterole
    noreplication nobypassrls;
grant spectator to app_reader;
alter role app_reader set default_transaction_read_only = on;
alter role app_user nologin;
-- Schema dumps contain ACL entries for the publisher's replication account.
-- It does not need to log in or stream data from a subscriber.
create role replication_user nologin noreplication;
