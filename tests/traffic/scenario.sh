#!/bin/sh
set -eu

api=http://api:8080/api/v1
mailpit=http://mailpit:8025
flight=00000000-0000-4000-8000-000000000050
marker="http-$(cat /proc/sys/kernel/random/uuid)"
passenger_email="passenger-$marker@airline.test"
response=/results/response.json

# shellcheck disable=SC2329
finish() {
    status=$?
    trap - EXIT INT TERM
    if [ -n "$capture" ]; then
        kill -INT "$capture" 2>/dev/null || true
        wait "$capture" || status=1
        tcpdump -nn -A -r /results/http.pcap > /results/http.txt 2>&1 || status=1
    fi
    chmod -R a+rX /results
    exit "$status"
}

: > /results/http.pcap
capture=
trap finish EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

request() {
    method=$1 url=$2 expected=$3
    shift 3
    echo "$method $url"
    body='(empty)'
    previous=
    for argument do
        if [ "$previous" = --json ]; then
            body=$(printf '%s' "$argument" | jq -c 'with_entries(if .key == "password" or .key == "code" then .value = "[redacted]" else . end)' 2>/dev/null || printf '[redacted]')
            break
        fi
        previous=$argument
    done
    printf 'Request body: %s\n' "$body"
    code=$(curl --silent --show-error --connect-timeout 5 --max-time 10 \
        -X "$method" -o "$response" -w '%{http_code}' "$url" "$@")
    echo "HTTP $code"
    if jq -e 'has("token")' "$response" >/dev/null 2>&1; then
        jq '.token = "[redacted]"' "$response"
    elif jq -e '.Subject? == "Airline Tracker login code"' "$response" >/dev/null 2>&1; then
        jq '.Text |= sub("verification code: [0-9]{6}"; "verification code: [redacted]")' "$response"
    else
        cat "$response"
    fi
    echo
    [ "$code" = "$expected" ]
}

login() {
    login_email=$(printf '%s' "$1" | jq -er '.email')
    request DELETE "$mailpit/api/v1/messages" 200
    request POST "$api/auth/tokens" 202 --json "$1"
    challenge=$(jq -er '.challenge_id' "$response")
    deadline=$(($(date +%s) + 30))
    otp_code=
    while [ -z "$otp_code" ] && [ "$(date +%s)" -lt "$deadline" ]; do
        request GET "$mailpit/api/v1/messages?limit=50" 200
        ids=$(jq -r --arg email "$login_email" \
            '.messages[] | select(.Subject == "Airline Tracker login code" and any(.To[]; .Address == $email)) | .ID' "$response")
        for id in $ids; do
            request GET "$mailpit/api/v1/message/$id" 200
            otp_code=$(jq -er '.Text | capture("verification code: (?<code>[0-9]{6})").code' "$response" 2>/dev/null || true)
            [ -z "$otp_code" ] || break
        done
        [ -n "$otp_code" ] || sleep 0.2
    done
    [ -n "$otp_code" ] || { echo "No login code received for $login_email" >&2; return 1; }
    verification=$(jq -nc --arg challenge "$challenge" --arg code "$otp_code" \
        '{challenge_id:$challenge,code:$code}')
    request POST "$api/auth/verify" 200 --json "$verification"
    token=$(jq -er '.token | select(length > 0)' "$response")
    jq '.token = "[redacted]"' "$response" > "$response.safe"
    mv "$response.safe" "$response"
}

passenger_password="Aa1!$(tr -d '-' < /proc/sys/kernel/random/uuid)"
credentials=$(jq -nc --arg email "$passenger_email" --arg password "$passenger_password" \
    '{email:$email,password:$password}')
request POST "$api/users" 201 --json "$credentials"
login "$credentials"
passenger_token=$token

admin_credentials=$(jq -nc --arg password "$E2E_ADMIN_PASSWORD" \
    '{email:"admin@airline.test",password:$password}')
login "$admin_credentials"
admin_token=$token

tcpdump --immediate-mode -i any -p -nn -s 0 -U -w /results/http.pcap \
    'tcp port 8080 or tcp port 8025' 2>/results/tcpdump.log &
capture=$!
attempt=0
until [ -s /results/http.pcap ]; do
    kill -0 "$capture"
    attempt=$((attempt + 1))
    [ "$attempt" -lt 50 ]
    sleep 0.1
done

request GET "$api/flights?limit=10" 200
jq -e --arg id "$flight" '.items[] | select(.id == $id and .status == "scheduled")' "$response"

subscription=$(jq -nc --arg id "$flight" '{flight_id:$id,notify_before_minutes:60}')
request POST "$api/subscriptions" 200 -H "Authorization: Bearer $passenger_token" --json "$subscription"
request GET "$api/subscriptions" 200 -H "Authorization: Bearer $passenger_token"
jq -e --arg id "$flight" '.items | length == 1 and .[0].id == $id' "$response"

update=$(jq -nc --arg plan "$marker" '{flight:{status:"delayed",plan:$plan}}')
request PATCH "$api/flights/$flight" 200 -H "Authorization: Bearer $admin_token" --json "$update"
request GET "$api/flights/$flight" 200
jq -e --arg id "$flight" --arg plan "$marker" \
    '.flight | .id == $id and .status == "delayed" and .plan == $plan' "$response"

deadline=$(($(date +%s) + 30))
while [ "$(date +%s)" -lt "$deadline" ]; do
    request GET "$mailpit/api/v1/messages?limit=50" 200
    messages=$(jq -r --arg email "$passenger_email" \
        '.messages[] | select(any(.To[]; .Address == $email)) | .ID' "$response")
    for id in $messages; do
        request GET "$mailpit/api/v1/message/$id" 200
        if jq -e --arg marker "$marker" --arg status "New flight's status is delayed" '
            .Subject == "Flight departure" and
            (.Text | contains($marker) and
                contains("from Sheremetyevo to Pulkovo has changed") and
                contains($status))
        ' "$response"; then
            echo "Scenario passed: $passenger_email, $marker"
            exit 0
        fi
    done
    sleep 0.2
done
echo "Flight change email was not received for $passenger_email within 30 seconds" >&2
exit 1
