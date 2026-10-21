#!/bin/sh
set -eu

api=http://api:8080/api/v1
mailpit=http://mailpit:8025
flight=00000000-0000-4000-8000-000000000050
marker="http-$(cat /proc/sys/kernel/random/uuid)"
email="passenger-$marker@airline.test"
response=/results/response.json

# shellcheck disable=SC2329
finish() {
    status=$?
    trap - EXIT INT TERM
    kill -INT "$capture" 2>/dev/null || true
    wait "$capture" || status=1
    tcpdump -nn -A -r /results/http.pcap > /results/http.txt 2>&1 || status=1
    chmod -R a+rX /results
    exit "$status"
}

: > /results/http.pcap
tcpdump --immediate-mode -i any -p -nn -s 0 -U -w /results/http.pcap \
    'tcp port 8080 or tcp port 8025' 2>/results/tcpdump.log &
capture=$!
trap finish EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
attempt=0
until [ -s /results/http.pcap ]; do
    kill -0 "$capture"
    attempt=$((attempt + 1))
    [ "$attempt" -lt 50 ]
    sleep 0.1
done

request() {
    method=$1 url=$2 expected=$3
    shift 3
    echo "$method $url"
    body='(empty)'
    previous=
    for argument do
        if [ "$previous" = --json ]; then
            body=$argument
            break
        fi
        previous=$argument
    done
    printf 'Request body: %s\n' "$body"
    code=$(curl --silent --show-error --connect-timeout 5 --max-time 10 \
        -X "$method" -o "$response" -w '%{http_code}' "$url" "$@")
    echo "HTTP $code"
    cat "$response"
    echo
    [ "$code" = "$expected" ]
}

login() {
    request POST "$api/auth/tokens" 200 --json "$1"
    token=$(jq -er '.token | select(length > 0)' "$response")
}

credentials=$(jq -nc --arg email "$email" '{email:$email,password:"Passenger123!"}')
request POST "$api/users" 201 --json "$credentials"
login "$credentials"

request GET "$api/flights?limit=10" 200
jq -e --arg id "$flight" '.items[] | select(.id == $id and .status == "scheduled")' "$response"

subscription=$(jq -nc --arg id "$flight" '{flight_id:$id,notify_before_minutes:60}')
request POST "$api/subscriptions" 200 -H "Authorization: Bearer $token" --json "$subscription"
request GET "$api/subscriptions" 200 -H "Authorization: Bearer $token"
jq -e --arg id "$flight" '.items | length == 1 and .[0].id == $id' "$response"

login '{"email":"admin@airline.test","password":"TestAdmin123!"}'
update=$(jq -nc --arg plan "$marker" '{flight:{status:"delayed",plan:$plan}}')
request PATCH "$api/flights/$flight" 200 -H "Authorization: Bearer $token" --json "$update"
request GET "$api/flights/$flight" 200
jq -e --arg id "$flight" --arg plan "$marker" \
    '.flight | .id == $id and .status == "delayed" and .plan == $plan' "$response"

deadline=$(($(date +%s) + 30))
while [ "$(date +%s)" -lt "$deadline" ]; do
    request GET "$mailpit/api/v1/messages?limit=50" 200
    messages=$(jq -r --arg email "$email" \
        '.messages[] | select(any(.To[]; .Address == $email)) | .ID' "$response")
    for id in $messages; do
        request GET "$mailpit/api/v1/message/$id" 200
        if jq -e --arg marker "$marker" --arg status "New flight's status is delayed" '
            .Subject == "Flight departure" and
            (.Text | contains($marker) and
                contains("from Sheremetyevo to Pulkovo has changed") and
                contains($status))
        ' "$response"; then
            echo "Scenario passed: $email, $marker"
            exit 0
        fi
    done
    sleep 0.2
done
echo 'Flight change email was not received within 30 seconds' >&2
exit 1
