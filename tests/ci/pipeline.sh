#!/bin/sh
set -eu
. tests/environment/stand.sh

pipeline_init() {
    environment_init "$CI_RUN_ID"
    export ALLURE_RESULTS_DIR="$CI_RESULTS/allure-results"
    mkdir -p "$ALLURE_RESULTS_DIR"
}

pipeline_prepare() {
    echo running > "$CI_RESULTS/prepare.status"
    { go version; git --version; java -version; allure --version; docker compose version; } > "$CI_RESULTS/tools.txt" 2>&1
    make test-env-up RUN_ID="$CI_RUN_ID" > "$CI_RESULTS/prepare.log" 2>&1
    echo passed > "$CI_RESULTS/prepare.status"
}

pipeline_unit() {
    echo running > "$CI_RESULTS/unit.status"
    make -C backend test-unit > "$CI_RESULTS/unit.log" 2>&1
    cp backend/api/coverage.out backend/api/coverage.html "$CI_RESULTS/"
    echo passed > "$CI_RESULTS/unit.status"
}

pipeline_integration() {
    echo running > "$CI_RESULTS/integration.status"
    make test-integration RUN_ID="$CI_RUN_ID" > "$CI_RESULTS/integration.log" 2>&1
    echo passed > "$CI_RESULTS/integration.status"
}

pipeline_e2e() {
    echo running > "$CI_RESULTS/e2e.status"
    make test-e2e RUN_ID="$CI_RUN_ID" > "$CI_RESULTS/e2e.log" 2>&1
    echo passed > "$CI_RESULTS/e2e.status"
}

pipeline_cleanup() {
    if [ -f "$results/baseline.fingerprint" ]; then
        environment_restore || return
    fi
    environment_down
}

pipeline_finish() {
    status=0
    dc logs --no-color > "$CI_RESULTS/services.log" 2>&1 || true
    if pipeline_cleanup > "$CI_RESULTS/final-cleanup.log" 2>&1; then
        echo passed > "$CI_RESULTS/cleanup.status"
    else
        echo failed > "$CI_RESULTS/cleanup.status"
        status=1
    fi

    python3 /opt/ci/report.py || status=1
    return "$status"
}
