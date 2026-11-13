#!/bin/sh
set -eu

environment_init() {
    export TEST_RUN_ID="${1:-local}"
    export E2E_ADMIN_PASSWORD="${E2E_ADMIN_PASSWORD:-Aa1!local-e2e-admin-password}"
    environment_dir=tests/environment
    results="$environment_dir/results/$TEST_RUN_ID"
    mkdir -p "$results"
}

dc() {
    if [ -n "${TEST_COMPOSE_EXTRA:-}" ]; then
        set -- -f "$TEST_COMPOSE_EXTRA" "$@"
    fi
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

environment_restore() {
    for service in integration-tests e2e-tests traffic; do
        for container in $(docker ps -aq \
            --filter "label=com.docker.compose.project=airline-test-$TEST_RUN_ID" \
            --filter "label=com.docker.compose.service=$service"); do
            docker rm -f "$container" || return
        done
    done
    dc stop api publisher notifier || return
    dc run --rm --no-deps drain >> "$results/kafka-drain.jsonl" || return
    db pg_restore -U postgres -d airline_tracker --clean --if-exists --exit-on-error \
        /var/lib/postgresql/data/test-baseline.dump || return
    verify_db || return
    dc exec -T redis redis-cli FLUSHALL || return
    test "$(dc exec -T redis redis-cli DBSIZE)" = 0 || return
    dc run --rm --no-deps tools python /test-support/check.py clear-mail || return
}

environment_reset() {
    environment_restore || return
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
        configure_admin_password || return
        snapshot
    else
        configure_admin_password || return
    fi
    topics
    dc up -d --no-deps --wait --wait-timeout 120 api notifier publisher
    environment_check
}

configure_admin_password() {
    dc exec -T -e E2E_ADMIN_PASSWORD="$E2E_ADMIN_PASSWORD" postgres \
        psql -X -q -v ON_ERROR_STOP=1 -U postgres -d airline_tracker <<'SQL'
\getenv admin_password E2E_ADMIN_PASSWORD
UPDATE api_schema.users
SET password_hash = crypt(:'admin_password', gen_salt('bf'))
WHERE email = 'admin@airline.test';
SQL
}

environment_check() {
    dc ps
    dc run --rm --no-deps tools
}

environment_down() {
    dc down --volumes --remove-orphans || return
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
    environment_up || return
    dc stop api publisher notifier || return
    environment_run dc run --build --rm --no-deps integration-tests
}

environment_e2e() {
    environment_up
    environment_run dc run --build --rm --no-deps e2e-tests
}

environment_traffic() {
    environment_up
    environment_run dc run --build --rm --no-deps traffic
}
