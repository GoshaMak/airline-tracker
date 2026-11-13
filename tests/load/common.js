import http from "k6/http";
import { check, fail, sleep } from "k6";
import exec from "k6/execution";
import crypto from "k6/crypto";
import { Counter, Rate, Trend } from "k6/metrics";

const api = (__ENV.BASE_URL || "http://localhost:8080") + "/api/v1";
const mailpit = __ENV.MAILPIT_URL || "http://localhost:8025";
const runID = __ENV.RUN_ID || "local";
const requests = new Counter("api_requests");
const unexpected = new Rate("unexpected_response");
const business = new Counter("business_operations");
const completed = new Counter("journeys_completed");
const delivered = new Rate("notification_delivered");
const deliveryTime = new Trend("notification_latency", true);
const journeyDuration = new Trend("journey_duration", true);
let operation = 0;
let passengerToken = "";
let requestStarted = 0;

export const stageDuration = __ENV.STAGE_DURATION || "15s";
export const normalVUs = Number(__ENV.VUS || 20);
export const maxVUs = Number(__ENV.MAX_VUS || 200);
if (!Number.isInteger(normalVUs) || normalVUs < 1 ||
    !Number.isInteger(maxVUs) || maxVUs < normalVUs) {
    throw new Error("VUS and MAX_VUS must be integers: 1 <= VUS <= MAX_VUS");
}

export function loadOptions(profile, stages) {
    return {
        scenarios: {
            users: {
                executor: "ramping-vus",
                startVUs: 0,
                stages,
                gracefulRampDown: "40s",
                gracefulStop: "40s",
            },
        },
        tags: {
            profile,
            run_id: runID,
        },
        systemTags: ["status", "method", "name", "scenario", "expected_response"],
        summaryTrendStats: ["min", "avg", "max", "p(50)", "p(75)", "p(90)", "p(95)", "p(99)"],
        thresholds: {
            unexpected_response: [{ threshold: "rate<0.01", abortOnFail: false }],
            notification_delivered: [{ threshold: "rate>0.99", abortOnFail: false }],
            journeys_completed: [{ threshold: "count>0", abortOnFail: false }],
        },
    };
}

function json(response) {
    try {
        return response.json()
    } catch (_) {
        return {}
    }
}

function request(method, path, body, token, expected = 200, kind = "valid") {
    requestStarted = Date.now();
    const response = http.request(
        method,
        api + path,
        body === null ? null : JSON.stringify(body),
        {
            headers: {
                "Content-Type": "application/json",
                Authorization: token ? `Bearer ${token}` : "",
            },
            tags: {
                name: `${method} ${path.split("?")[0].replace(/[0-9a-f]{8}-[0-9a-f-]{27}/g, ":id")}`,
                target: "api",
                kind,
            },
            responseCallback: http.expectedStatuses(expected),
            timeout: "15s",
        },
    );
    requests.add(1, { kind });
    unexpected.add(response.status !== expected);
    check(response, {
        "expected API status": (r) => r.status === expected,
    });
    return response;
}

function loginWithEmailCode(email, password) {
    const started = Date.now();
    const challengeResponse = request("POST", "/auth/tokens", { email, password }, "", 202);
    const challengeId = json(challengeResponse).challenge_id;
    if (!challengeId) {
        fail("Login did not return an email verification challenge");
    }

    const deadline = Date.now() + Number(__ENV.MAIL_TIMEOUT_SECONDS || 30) * 1000;
    let code = "";
    while (Date.now() < deadline && !code) {
        const query = encodeURIComponent(`to:${email} "Airline Tracker login code"`);
        const search = http.get(
            `${mailpit}/api/v1/search?query=${query}&limit=10`,
            {
                tags: {
                    name: "GET Mailpit search",
                    target: "mailpit",
                },
                timeout: "5s",
            });
        for (const summary of (json(search).messages || [])) {
            if (Date.parse(summary.Created) < started) {
                continue;
            }
            const detail = http.get(
                `${mailpit}/api/v1/message/${encodeURIComponent(summary.ID)}`,
                {
                    tags: {
                        name: "GET Mailpit message",
                        target: "mailpit",
                    },
                    timeout: "5s",
                });
            const match = (json(detail).Text || "").match(/verification code:\s*([0-9]{6})/);
            if (match) {
                code = match[1];
                break;
            }
        }
        if (!code) {
            sleep(0.2);
        }
    }
    if (!code) {
        fail(`Email verification code did not arrive for ${email}`);
    }
    const verified = request("POST", "/auth/verify", { challenge_id: challengeId, code }, "", 200);
    const token = json(verified).token;
    if (!token) {
        fail("Email verification did not return a token");
    }
    return token;
}

export function setup() {
    const password = __ENV.E2E_ADMIN_PASSWORD;
    if (!password) {
        fail("Set E2E_ADMIN_PASSWORD before make load-up");
    }
    const adminToken = loginWithEmailCode("admin@airline.test", password);
    if (!adminToken) {
        fail("Run make load-up with the same RUN_ID first");
    }
    return { adminToken };
}

function businessRequest(method, path, body, token, errorPercent) {
    while ((operation + exec.vu.idInTest) % 10 < errorPercent / 10) {
        operation++;
        business.add(1, { kind: "invalid" });
        request(
            "POST",
            "/subscriptions",
            {
                flight_id: "not-a-uuid",
                notify_before_minutes: 60,
            },
            passengerToken,
            400,
            "invalid",
        );
        sleep(Number(__ENV.THINK_SECONDS || 1));
    }
    operation++;
    business.add(1, { kind: "valid" });
    return request(method, path, body, token);
}

function runJourney(errorPercent = 0, setupData = {}) {
    const user = exec.vu.idInTest;
    const email = `load-${user}@airline.test`;
    const hash = crypto.md5(`load-flight-${user}`, "hex");
    const flight = `${hash.slice(0, 8)}-${hash.slice(8, 12)}-${hash.slice(12, 16)}-${hash.slice(16, 20)}-${hash.slice(20)}`;
    const marker = `${runID}-${user}-${exec.scenario.iterationInTest}-${Date.now()}`;
    const pause = () => sleep(Number(__ENV.THINK_SECONDS || 0));
    const token = loginWithEmailCode(email, __ENV.E2E_ADMIN_PASSWORD);
    if (!token) {
        pause();
        return;
    }
    passengerToken = token;

    businessRequest("GET", "/flights?limit=20", null, token, errorPercent);
    pause();
    businessRequest("GET", `/flights/${flight}`, null, token, errorPercent);
    pause();
    const subscription = businessRequest("POST", "/subscriptions", {
        flight_id: flight, notify_before_minutes: 60,
    }, token, errorPercent);
    businessRequest("GET", "/subscriptions", null, token, errorPercent);
    pause();

    const adminToken = setupData.adminToken;
    if (!adminToken || subscription.status !== 200) {
        pause();
        return;
    }

    const updated = businessRequest("PATCH", `/flights/${flight}`, {
        flight: { status: "delayed", plan: marker },
    }, adminToken, errorPercent);
    if (updated.status !== 200) {
        pause();
        return;
    }

    const received = waitForMail(email, marker, requestStarted);
    delivered.add(received);
    check(received, { "flight change email received": (ok) => ok });
    if (received) {
        completed.add(1);
    }
    pause();
}

export function journey(errorPercent = 0, setupData = {}) {
    const started = Date.now();
    try {
        runJourney(errorPercent, setupData);
    } finally {
        journeyDuration.add(Date.now() - started);
    }
}

function waitForMail(email, marker, started) {
    const deadline = Date.now() + Number(__ENV.MAIL_TIMEOUT_SECONDS || 30) * 1000;
    const query = encodeURIComponent(`to:${email} "${marker}"`);
    while (Date.now() < deadline) {
        const response = http.get(`${mailpit}/api/v1/search?query=${query}&limit=1`, {
            tags: { name: "GET Mailpit search", target: "mailpit" }, timeout: "5s",
        });
        const mail = (json(response).messages || [])[0];
        if (response.status === 200 && mail) {
            const received = Date.parse(mail.Created);
            if (!Number.isFinite(received) || received < started) {
                return false;
            }
            deliveryTime.add(received - started);
            return true;
        }
        sleep(1);
    }
    return false;
}

export function handleSummary(data) {
    return {
        "/results/summary.json": JSON.stringify(data, null, 2),
        stdout: `\nSummary saved to /results/summary.json. Completed journeys: ${data.metrics.journeys_completed?.values.count || 0}\n`,
    };
}
