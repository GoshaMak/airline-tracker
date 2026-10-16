#!/bin/sh
set -eu

ci_init() {
    export CI_RUN_ID="${1:-local}"
    CI_ROOT=$(pwd)
    export CI_ROOT
    export CI_WORK="$CI_ROOT/tests/ci/work/$CI_RUN_ID"
    export CI_RESULTS="$CI_WORK/repo/tests/environment/results/$CI_RUN_ID"
    export CI_HISTORY="$CI_ROOT/tests/ci/history"
    export CI_REVISION="${CI_REVISION:-$(git rev-parse HEAD)}"
    export CI_IMAGE="airline-tracker-ci:$CI_RUN_ID"
    docker_endpoint=$(docker context inspect --format '{{.Endpoints.docker.Host}}')
    ci_socket=${DOCKER_SOCKET:-${docker_endpoint#unix://}}
    mkdir -p "$CI_WORK" "$CI_HISTORY" "$CI_ROOT/tests/ci/cache/go-mod" "$CI_ROOT/tests/ci/cache/go-build"
}

ci_build() {
    docker build -t "$CI_IMAGE" -f tests/ci/Dockerfile .
}

ci_container() {
    docker run --rm --init --name "airline-ci-$CI_RUN_ID" \
        -v "$CI_ROOT:$CI_ROOT" -v "$CI_ROOT:/source:ro" \
        -v "$ci_socket:/var/run/docker.sock" \
        -v "$CI_ROOT/tests/ci/cache/go-mod:/go/pkg/mod" \
        -v "$CI_ROOT/tests/ci/cache/go-build:/root/.cache/go-build" \
        -w "$CI_WORK" \
        -e CI_RUN_ID -e CI_WORK -e CI_RESULTS -e CI_HISTORY \
        -e CI_REVISION -e GITHUB_RUN_ID -e GITHUB_REPOSITORY -e GITHUB_SERVER_URL \
        "$CI_IMAGE" "$@"
}

ci_checkout() {
    ci_container sh /opt/ci/checkout.sh
}

ci_step() {
    if ci_container sh -c 'cd "$CI_WORK/repo"; . tests/ci/pipeline.sh; pipeline_init; "pipeline_$1"' sh "$1"; then
        ci_step_status=0
    else
        ci_step_status=$?
    fi
    if [ -f "$CI_RESULTS/$1.log" ]; then
        tail -n 40 "$CI_RESULTS/$1.log"
    fi
    return "$ci_step_status"
}

ci_finish() {
    docker rm -f "airline-ci-$CI_RUN_ID" >/dev/null 2>&1 || true
    if [ -f "$CI_WORK/repo/tests/ci/pipeline.sh" ]; then
        ci_step finish
    elif docker image inspect "$CI_IMAGE" >/dev/null 2>&1; then
        ci_container python3 /opt/ci/report.py
    fi
}

ci_complete() {
    ci_exit_status=$?
    trap - EXIT INT TERM
    ci_finish || ci_exit_status=1
    exit "$ci_exit_status"
}

ci_all() (
    trap ci_complete EXIT
    trap 'exit 130' INT
    trap 'exit 143' TERM
    ci_build
    ci_checkout
    ci_step prepare
    ci_step unit
    ci_step integration
    ci_step e2e
)
