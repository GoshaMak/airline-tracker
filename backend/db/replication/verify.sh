#!/bin/sh
set -eu

query() {
    psql -X --set=ON_ERROR_STOP=1 --host="$1" --tuples-only --no-align --command="$2"
}

if [ "$(PGOPTIONS='' query postgres "SELECT current_setting('synchronous_commit') = 'remote_apply'")" != t ]; then
    printf '%s\n' 'Primary synchronous_commit must be remote_apply.' >&2
    exit 1
fi
if [ "$(query postgres "SELECT count(*) = 2 AND count(DISTINCT application_name) = 2 FROM pg_stat_replication WHERE application_name IN ('airline_replica_1', 'airline_replica_2') AND state = 'streaming' AND sync_state = 'sync'")" != t ]; then
    printf '%s\n' 'Both synchronous subscribers must be connected before the check.' >&2
    exit 1
fi

# Only a dedicated diagnostic row is changed. No business rows are modified.
generation=$(PGOPTIONS='-c synchronous_commit=remote_apply -c statement_timeout=10000' \
    psql -X --set=ON_ERROR_STOP=1 --host=postgres --tuples-only --no-align --quiet \
    --command='INSERT INTO replication_meta.probe VALUES (1, gen_random_uuid()) ON CONFLICT (id) DO UPDATE SET generation = EXCLUDED.generation RETURNING generation')

for replicaIndex in 1 2; do
    replicaHost="postgres-replica-$replicaIndex"
    if [ "$(query "$replicaHost" 'SELECT generation FROM replication_meta.probe WHERE id = 1')" != "$generation" ]; then
        printf '%s\n' "Replica $replicaIndex does not contain the acknowledged transaction." >&2
        exit 1
    fi
    printf '%s\n' "$replicaHost: acknowledged transaction is visible"

    if [ "$(PGUSER=app_reader PGPASSWORD="${READER_PASSWORD:-reader_pswd}" query "$replicaHost" "SELECT current_setting('transaction_read_only')")" != on ]; then
        printf '%s\n' "$replicaHost: app_reader is not read-only by default" >&2
        exit 1
    fi
    for transactionMode in 'BEGIN' 'BEGIN READ WRITE'; do
        # ROLLBACK protects data even if the expected permission check fails.
        if errorOutput=$(PGUSER=app_reader PGPASSWORD="${READER_PASSWORD:-reader_pswd}" \
            psql -X --set=ON_ERROR_STOP=1 --set=VERBOSITY=sqlstate --host="$replicaHost" \
            --command="$transactionMode; UPDATE replication_meta.probe SET generation = gen_random_uuid() WHERE id = 1; ROLLBACK;" 2>&1); then
            printf '%s\n' "$replicaHost: app_reader unexpectedly accepted a write" >&2
            exit 1
        else
            case "$errorOutput" in
                *25006*|*42501*) printf '%s\n' "$replicaHost: write rejected ($transactionMode)" ;;
                *) printf '%s\n' "$replicaHost: unexpected failure during permission check" >&2; exit 1 ;;
            esac
        fi
    done
    psql -X --set=ON_ERROR_STOP=1 --host="$replicaHost" --command='
        SELECT subname, pid, received_lsn, latest_end_lsn,
            now() - last_msg_receipt_time AS feedback_age
        FROM pg_stat_subscription WHERE relid IS NULL;'
done

psql -X --set=ON_ERROR_STOP=1 --host=postgres --command="
    SELECT application_name, state, sync_state, write_lag, flush_lag, replay_lag,
        pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) AS replay_distance_bytes
    FROM pg_stat_replication WHERE application_name IN ('airline_replica_1', 'airline_replica_2') ORDER BY application_name;"
printf '%s\n' 'PASS: both replicas applied the synchronous commit; application writes are denied on both.'
