#!/bin/sh
set -eu

: "${REPLICATION_PASSWORD:?REPLICATION_PASSWORD is required}"
case "${REPLICATION_SETUP_TIMEOUT:-300}" in
    ''|*[!0-9]*) printf '%s\n' 'Invalid REPLICATION_SETUP_TIMEOUT' >&2; exit 1 ;;
esac
replicationDeadline=$(( $(date +%s) + ${REPLICATION_SETUP_TIMEOUT:-300} ))
schemaFile=$(mktemp /tmp/airline-schema.XXXXXX.sql)
upgradeFile=$(mktemp /tmp/airline-upgrade.XXXXXX.sql)
trap 'rm -f "$schemaFile" "$upgradeFile"' EXIT HUP INT TERM

query() {
    psql -X --set=ON_ERROR_STOP=1 --host="$1" --tuples-only --no-align --command="$2"
}

require_time() {
    if [ "$(date +%s)" -ge "$replicationDeadline" ]; then
        printf '%s\n' 'Replication initialization timed out; inspect subscriber logs and pg_subscription_rel.' >&2
        exit 1
    fi
}

if [ "$(query postgres "SELECT current_setting('wal_level') = 'logical'")" != t ]; then
    printf '%s\n' 'Primary wal_level must be logical.' >&2
    exit 1
fi
if [ "$(PGOPTIONS='' query postgres "SELECT current_setting('synchronous_commit') = 'remote_apply'")" != t ]; then
    printf '%s\n' 'Primary synchronous_commit must be remote_apply.' >&2
    exit 1
fi

psql -X --set=ON_ERROR_STOP=1 --host=postgres \
    --set=replication_password="$REPLICATION_PASSWORD" --file=/replication/publisher.sql

# DDL, functions, grants and defaults are not transferred by logical replication
pg_dump --host=postgres --schema-only --no-owner --no-publications --no-subscriptions \
    --schema=api_schema --schema=notifier_schema --schema=replication_meta \
    --exclude-table='*.goose_db_version' --file="$schemaFile"

for replicaIndex in 1 2; do
    replicaHost="postgres-replica-$replicaIndex"
    subscriptionName="airline_replica_$replicaIndex"
    if [ "$(query "$replicaHost" "SELECT to_regclass('api_schema.users') IS NULL")" = t ]; then
        psql -X --set=ON_ERROR_STOP=1 --host="$replicaHost" --single-transaction --file="$schemaFile"
    else
        for migration in /replica-migrations/outbox.sql /replica-migrations/notifications.sql /replica-migrations/inbox-source.sql; do
            sed '/^-- +goose Down/,$d' "$migration" > "$upgradeFile"
            psql -X --set=ON_ERROR_STOP=1 --host="$replicaHost" --single-transaction --file="$upgradeFile"
        done
    fi

    if [ "$(query "$replicaHost" "SELECT bool_and(has_table_privilege('app_reader', c.oid, 'SELECT') AND NOT has_table_privilege('app_reader', c.oid, 'INSERT,UPDATE,DELETE,TRUNCATE')) FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname IN ('api_schema', 'replication_meta') AND c.relkind IN ('r', 'p')")" != t ]; then
        printf '%s\n' "$replicaHost: app_reader must have SELECT and no data modification privileges." >&2
        exit 1
    fi

    if [ "$(query "$replicaHost" "SELECT EXISTS (SELECT FROM pg_subscription WHERE subname = '$subscriptionName')")" = f ]; then
        if [ "$(query postgres "SELECT EXISTS (SELECT FROM pg_replication_slots WHERE slot_name = '$subscriptionName')")" = t ]; then
            printf '%s\n' "Orphan slot $subscriptionName exists: restore its replica volume or resolve it manually; no slot is dropped automatically." >&2
            exit 1
        fi
        # Never overwrite existing subscriber data or risk duplicate-key conflicts.
        psql -X --set=ON_ERROR_STOP=1 --host="$replicaHost" --file=/replication/empty-subscriber.sql
    fi

    # Quote the password as a libpq value, then quote the whole connection in SQL.
    escapedPassword=$(printf '%s' "$REPLICATION_PASSWORD" | sed "s/\\\\/\\\\\\\\/g; s/'/\\\\'/g")
    connection="host=postgres port=5432 dbname=airline_tracker user=replication_user password='$escapedPassword' application_name=$subscriptionName connect_timeout=5"
    psql -X --set=ON_ERROR_STOP=1 --host="$replicaHost" \
        --set=subscription_name="$subscriptionName" --set=connection="$connection" \
        --file=/replication/subscription.sql
done

publicationTables=$(query postgres "SELECT count(*) FROM pg_publication_tables WHERE pubname = 'airline_tracker_publication'")
for replicaIndex in 1 2; do
    replicaHost="postgres-replica-$replicaIndex"
    subscriptionName="airline_replica_$replicaIndex"
    until [ "$(query "$replicaHost" "SELECT count(*) = $publicationTables AND bool_and(srsubstate = 'r') FROM pg_subscription_rel WHERE srsubid = (SELECT oid FROM pg_subscription WHERE subname = '$subscriptionName')")" = t ]; do
        require_time
        sleep 1
    done
done

until [ "$(query postgres "SELECT count(*) = 2 AND count(DISTINCT application_name) = 2 FROM pg_stat_replication WHERE application_name IN ('airline_replica_1', 'airline_replica_2') AND state = 'streaming' AND sync_state = 'sync'")" = t ]; do
    require_time
    sleep 1
done

printf '%s\n' 'Initial copy completed; both synchronous subscribers are connected.'
