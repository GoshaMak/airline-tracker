SELECT 'CREATE ROLE replication_user LOGIN REPLICATION'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'replication_user')
\gexec

ALTER ROLE replication_user LOGIN REPLICATION PASSWORD :'replication_password';
GRANT CONNECT ON DATABASE airline_tracker TO replication_user;

CREATE SCHEMA IF NOT EXISTS replication_meta;
CREATE TABLE IF NOT EXISTS replication_meta.probe (
    id smallint PRIMARY KEY CHECK (id = 1),
    generation uuid NOT NULL
);

GRANT USAGE ON SCHEMA api_schema, notifier_schema, replication_meta TO replication_user;
GRANT SELECT ON ALL TABLES IN SCHEMA api_schema, notifier_schema, replication_meta TO replication_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA api_schema, notifier_schema, replication_meta
    GRANT SELECT ON TABLES TO replication_user;

-- Backfill privileges if the old migrations had already run on this volume.
GRANT USAGE ON SCHEMA api_schema TO spectator;
GRANT SELECT ON ALL TABLES IN SCHEMA api_schema TO spectator;
ALTER DEFAULT PRIVILEGES IN SCHEMA api_schema GRANT SELECT ON TABLES TO spectator;
GRANT USAGE ON SCHEMA replication_meta TO app_reader;
GRANT SELECT ON TABLE replication_meta.probe TO app_reader;

-- Publish business tables, but not Goose's per-node migration bookkeeping.
SELECT 'CREATE PUBLICATION airline_tracker_publication'
WHERE NOT EXISTS (SELECT FROM pg_publication WHERE pubname = 'airline_tracker_publication')
\gexec

SELECT format('ALTER PUBLICATION airline_tracker_publication SET TABLE %s',
    string_agg(format('%I.%I', schemaname, tablename), ', ' ORDER BY schemaname, tablename))
FROM pg_tables
WHERE schemaname IN ('api_schema', 'notifier_schema', 'replication_meta')
    AND tablename <> 'goose_db_version'
HAVING count(*) > 0
\gexec
