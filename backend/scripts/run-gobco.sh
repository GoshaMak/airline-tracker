#!/bin/sh
set -eu

backend_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
gobco_package=${1:-./internal/flight/usecase}
gobco_tmp_dir=$(mktemp -d)
trap 'rm -rf "$gobco_tmp_dir"' EXIT HUP INT TERM

cd "$backend_dir/api"
cp go.mod go.sum "$gobco_tmp_dir/"
GOWORK=off go mod edit \
  -modfile="$gobco_tmp_dir/go.mod" \
  -require=shared@v0.0.0 \
  -replace="shared=$backend_dir/shared"

GOWORK=off \
GOFLAGS="${GOFLAGS:-} -modfile=$gobco_tmp_dir/go.mod" \
GOCACHE="${GOCACHE:-/tmp/airline-tracker-go-build}" \
ALLURE_RESULTS_DIR="$gobco_tmp_dir/allure-results" \
  gobco "$gobco_package"
