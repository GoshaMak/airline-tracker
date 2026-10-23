#!/bin/sh
set -u

output=$(mktemp)
if "$@" >"$output" 2>&1; then
    status=0
else
    status=$?
fi

grep -v '\[no test files\]' "$output" | grep -v 'coverage: 0.0% of statements' || true
rm -f "$output"
exit "$status"
