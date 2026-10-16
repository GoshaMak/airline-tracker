#!/bin/sh
set -eu

backend_dir=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
results_dir="$backend_dir/allure-results"
allure_bin=$(sh "$backend_dir/scripts/install-allure.sh")

mkdir -p "$results_dir"
find "$results_dir" -mindepth 1 -depth -delete

test_status=0
GOCACHE="${GOCACHE:-/tmp/airline-tracker-go-build}" \
ALLURE_RESULTS_DIR="$results_dir" \
ALLURE_LABEL_EPIC=api \
  make -C "$backend_dir" test-unit || test_status=$?

"$allure_bin" generate "$results_dir" \
  --clean --output "$backend_dir/allure-report"

exit "$test_status"
