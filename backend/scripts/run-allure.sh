#!/bin/sh
set -eu

backend_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
results_dir="$backend_dir/allure-results"
allure_version=2.46.1
allure_sha256=37428164f310c8d137cfdd2c3a25d049b233261ddab13e1bd4dfd7627b5cbf8d
allure_bin="$backend_dir/.tools/allure-$allure_version/bin/allure"

if [ ! -x "$allure_bin" ]; then
  archive="$backend_dir/.tools/allure-$allure_version.tgz"
  mkdir -p "$backend_dir/.tools"
  curl -fL --retry 2 \
    "https://github.com/allure-framework/allure2/releases/download/$allure_version/allure-$allure_version.tgz" \
    -o "$archive"
  printf '%s  %s\n' "$allure_sha256" "$archive" | sha256sum -c -
  tar -xzf "$archive" -C "$backend_dir/.tools"
fi

mkdir -p "$results_dir"
find "$results_dir" -mindepth 1 -depth -delete

test_status=0
if (
  cd "$backend_dir/api"
  GOCACHE="${GOCACHE:-/tmp/airline-tracker-go-build}" \
  ALLURE_RESULTS_DIR="$results_dir" \
  ALLURE_LABEL_EPIC=api \
    go test -cover -count=1 -coverprofile=coverage.out ./... 2>&1 | grep -v '\[no test files\]' | grep -v 'coverage: 0.0% of statements'
); then
  cd "$backend_dir/api" && python3 ../scripts/render-coverage.py
else
  test_status=1
fi

"$allure_bin" generate "$results_dir" \
  --clean --output "$backend_dir/allure-report"

exit "$test_status"
