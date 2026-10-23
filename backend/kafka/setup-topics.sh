#!/bin/sh
set -eu

topics=/opt/kafka/bin/kafka-topics.sh
for topic in "$SUBSCRIPTION_CREATED_TOPIC" "$FLIGHT_UPDATED_TOPIC"; do
    "$topics" --bootstrap-server "$KAFKA_PEERS" --create --if-not-exists \
        --topic "$topic" --partitions 3 --replication-factor 1
    partitions=$("$topics" --bootstrap-server "$KAFKA_PEERS" --describe --topic "$topic" \
        | sed -n 's/.*PartitionCount: \([0-9]*\).*/\1/p')
    if [ "$partitions" -lt 3 ]; then
        "$topics" --bootstrap-server "$KAFKA_PEERS" --alter --topic "$topic" --partitions 3
    fi
done
