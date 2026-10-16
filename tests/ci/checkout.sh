#!/bin/sh
set -eu
git config --global --add safe.directory /source

test ! -e "$CI_WORK/repo"
mkdir -p "$CI_RESULTS"
echo running > "$CI_RESULTS/checkout.status"
git init -q "$CI_WORK/repo"
git -C "$CI_WORK/repo" fetch --depth=1 /source "$CI_REVISION"
git -C "$CI_WORK/repo" checkout --detach FETCH_HEAD
git -C "$CI_WORK/repo" rev-parse HEAD > "$CI_RESULTS/commit.txt"
echo passed > "$CI_RESULTS/checkout.status"
