#!/bin/sh
set -eu

backend_dir=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
tools_dir=${1:-"$backend_dir/.tools"}
version=2.46.1
checksum=37428164f310c8d137cfdd2c3a25d049b233261ddab13e1bd4dfd7627b5cbf8d
allure_bin="$tools_dir/allure-$version/bin/allure"

if [ ! -x "$allure_bin" ]; then
    archive="$tools_dir/allure-$version.tgz"
    mkdir -p "$tools_dir"
    curl -fL --retry 3 \
        "https://github.com/allure-framework/allure2/releases/download/$version/allure-$version.tgz" \
        -o "$archive"
    printf '%s  %s\n' "$checksum" "$archive" | sha256sum -c - >&2
    tar -xzf "$archive" -C "$tools_dir"
    rm "$archive"
fi

printf '%s\n' "$allure_bin"
