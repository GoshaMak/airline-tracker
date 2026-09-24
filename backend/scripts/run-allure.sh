#!/bin/sh
set -eu

backend_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
results_dir="$backend_dir/allure-results"
mode=${1:-results}

mkdir -p "$results_dir"
find "$results_dir" -mindepth 1 -depth -delete

test_status=0
for module in api notifier; do
  if (
    cd "$backend_dir/$module"
    GOCACHE=/tmp/airline-tracker-go-build \
    GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local \
    ALLURE_RESULTS_DIR="$results_dir" \
    ALLURE_LABEL_EPIC="$module" \
      go test -count=1 ./...
  ); then
    :
  else
    test_status=1
  fi
done

if [ "$mode" = report ]; then
  "${ALLURE_BIN:?Allure CLI path is required}" generate "$results_dir" \
    --clean --output "$backend_dir/allure-report"
fi

exit "$test_status"
