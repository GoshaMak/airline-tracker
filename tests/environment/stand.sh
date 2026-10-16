#!/bin/sh
set -eu

environment_init() {
    export TEST_RUN_ID="${1:-local}"
    environment_dir=tests/environment
    results="$environment_dir/results/$TEST_RUN_ID"
    mkdir -p "$results"
}

dc() {
    docker compose --env-file /dev/null -p "airline-test-$TEST_RUN_ID" -f "$environment_dir/compose.yml" "$@";
}

db() {
    dc exec -T postgres "$@"
}

topics() {
    for topic in subscription-created flight-updated; do
        dc exec -T kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 \
            --create --if-not-exists --topic "$topic" --partitions 1 --replication-factor 1 || return
    done
}

verify_db() {
    db psql -X -qAt -U postgres -d airline_tracker -f /test-support/fingerprint.sql > "$results/current.fingerprint" || return
    diff -u "$results/baseline.fingerprint" "$results/current.fingerprint"
}

snapshot() {
    db pg_dump -U postgres -d airline_tracker -Fc -f /var/lib/postgresql/data/test-baseline.dump
    db psql -X -qAt -U postgres -d airline_tracker -f /test-support/fingerprint.sql > "$results/baseline.fingerprint"
}

environment_reset() {
    dc stop api publisher notifier || return
    db pg_restore -U postgres -d airline_tracker --clean --if-exists --exit-on-error \
        /var/lib/postgresql/data/test-baseline.dump || return
    verify_db || return
    dc exec -T redis redis-cli FLUSHALL || return
    test "$(dc exec -T redis redis-cli DBSIZE)" = 0 || return
    dc run --rm --no-deps tools python /test-support/check.py clear-mail || return

    dc rm -sfv kafka || return
    dc up -d --wait --wait-timeout 180 kafka || return
    topics || return
    echo "Storage restored and verified: $TEST_RUN_ID"
}

environment_up() {
    dc build api
    dc up -d --wait --wait-timeout 180 postgres redis kafka mailpit
    if [ ! -f "$results/baseline.fingerprint" ]; then
        dc run --rm --no-deps api-migrate
        dc run --rm --no-deps notifier-migrate
        db psql -X -U postgres -d airline_tracker -f /test-support/seed.sql
        snapshot
    fi
    topics
    dc up -d --no-deps --wait --wait-timeout 120 api notifier publisher
    environment_check
}

environment_check() {
    dc ps
    dc run --rm --no-deps tools
}

environment_down() {
    dc down --volumes --remove-orphans
    if [ -f "$results/baseline.fingerprint" ]; then
        mv "$results/baseline.fingerprint" "$results/last-baseline.fingerprint"
    fi
}

environment_cleanup() {
    status=$?
    trap - EXIT INT TERM
    if [ "$status" -ne 0 ]; then
        dc logs --no-color > "$results/failure.log" 2>&1 || true
    fi
    if ! environment_reset > "$results/cleanup.log" 2>&1; then
        cat "$results/cleanup.log" >&2
        exit 1
    fi
    exit "$status"
}

environment_run() (
    trap environment_cleanup EXIT
    trap 'exit 130' INT
    trap 'exit 143' TERM
    "$@"
)

environment_integration() {
    dc stop api publisher notifier
    environment_run dc run --build --rm --no-deps integration-tests
}
