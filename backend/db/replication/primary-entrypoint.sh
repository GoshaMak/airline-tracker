#!/bin/sh
set -eu

# The initial CREATE DATABASE / role setup precedes both subscribers.
# This affects only initdb's psql clients, not external application connections.
if [ ! -s "${PGDATA:-/var/lib/postgresql/data}/PG_VERSION" ]; then
    export PGOPTIONS="${PGOPTIONS:-} -c synchronous_commit=local"
fi
exec /usr/local/bin/docker-entrypoint.sh "$@"
