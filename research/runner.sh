#!/bin/bash

set -e

# for step in $(seq 1000 1000 5000); do
#     echo "=== step: $step ==="
#
#     sed -i "s/^FLIGHTS_AMT=.*/FLIGHTS_AMT=$step/" ../utils/.env
#
#     make up
#
#     sleep 1
# done
# exit 0

for step in $(seq 10000 5000 50000); do
    echo "=== step: $step ==="

    sed -i "s/^FLIGHTS_AMT=.*/FLIGHTS_AMT=$step/" ../utils/.env

    make up

    sleep 1
done
