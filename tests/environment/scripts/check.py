import json
import smtplib
import sys
import urllib.request


def request(url, method="GET", data=None):
    payload = None if data is None else json.dumps(data).encode()
    req = urllib.request.Request(
        url, data=payload, method=method, headers={"Content-Type": "application/json"}
    )
    with urllib.request.urlopen(req, timeout=10) as response:
        body = response.read()
        if not body:
            return None
        if "application/json" in response.headers.get("Content-Type", ""):
            return json.loads(body)
        return body.decode()


if sys.argv[1:] == ["clear-mail"]:
    request("http://mailpit:8025/api/v1/messages", method="DELETE")
    assert request("http://mailpit:8025/api/v1/messages")["total"] == 0
else:
    request("http://publisher:9090/metrics")
    request("http://notifier:9090/metrics")
    assert request("http://api:8080/api/v1/status") == "OK"
    flights = request("http://api:8080/api/v1/flights?limit=10")["items"]
    assert any(f["id"] == "00000000-0000-4000-8000-000000000050" for f in flights)
    token = request(
        "http://api:8080/api/v1/auth/tokens",
        method="POST",
        data={
            "email": "admin@airline.test",
            "password": "TestAdmin123!",
        },
    )["token"]
    assert token

    with smtplib.SMTP("mailpit", 1025, timeout=10) as smtp:
        assert smtp.noop()[0] == 250
    request("http://mailpit:8025/api/v1/messages")
    print("Ready: API, fixture flight, administrator login, SMTP and Mailpit API")
