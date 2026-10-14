import http from "k6/http";
import { check, fail } from "k6";
import exec from "k6/execution";
import { Counter, Rate, Trend } from "k6/metrics";

function integer(name, fallback, minimum) {
    const value = Number(__ENV[name] ?? fallback);
    if (!Number.isInteger(value) || value < minimum) {
        throw new Error(`${name} must be an integer >= ${minimum}`);
    }
    return value;
}

const baseURL = (__ENV.BASE_URL || "http://localhost:8080").replace(/\/$/, "");
const rps = integer("RATE", 100, 2);
if (rps % 2 !== 0) throw new Error("RATE must be even");
const duration = integer("DURATION_SECONDS", 60, 1);
const warmup = integer("WARMUP_SECONDS", 20, 0);
if (__ENV.CHECK_BALANCE && !["true", "false"].includes(__ENV.CHECK_BALANCE)) {
    throw new Error("CHECK_BALANCE must be true or false");
}
const checkBalance = __ENV.CHECK_BALANCE !== "false";
const limit = 50;
const summaryPath = __ENV.SUMMARY_PATH || "tests/load/results/baseline.json";

const airportsLatency = new Trend("airports_latency", true);
const flightsLatency = new Trend("flights_latency", true);
const measuredRequests = new Counter("measured_requests");
const nextPageRequests = new Counter("next_page_requests");
const validResponses = new Rate("valid_responses");
const mainShare = new Rate("backend_main_share");
const reader1Share = new Rate("backend_reader_1_share");
const reader2Share = new Rate("backend_reader_2_share");
const unknownShare = new Rate("backend_unknown_share");
const mainRequests = new Counter("backend_main_requests");
const reader1Requests = new Counter("backend_reader_1_requests");
const reader2Requests = new Counter("backend_reader_2_requests");
const cursors = { airports: null, flights: null };
const sequences = { airports: 0, flights: 0 };

function scenario(endpoint, phase, rps, seconds, startTime) {
    return {
        executor: "constant-arrival-rate",
        exec: endpoint,
        rate: rps,
        timeUnit: "1s",
        duration: `${seconds}s`,
        startTime: `${startTime}s`,
        preAllocatedVUs: phase === "load" ? 50 : 10,
        maxVUs: phase === "load" ? 100 : 20,
        gracefulStop: "10s",
        tags: { phase, endpoint },
    };
}

const scenarios = {
    airports_load: scenario("airports", "load", rps / 2, duration, warmup),
    flights_load: scenario("flights", "load", rps / 2, duration, warmup),
};
if (warmup > 0) {
    scenarios.airports_warmup = scenario("airports", "warmup", 10, warmup, 0);
    scenarios.flights_warmup = scenario("flights", "warmup", 10, warmup, 0);
}

export const options = {
    scenarios,
    summaryTrendStats: ["avg", "min", "med", "max", "p(90)", "p(95)", "p(99)"],
    thresholds: {
        airports_latency: ["p(95)<500", "p(99)<1000"],
        flights_latency: ["p(95)<500", "p(99)<1000"],
        "http_req_failed{phase:load}": ["rate<0.01"],
        "checks{phase:load}": ["rate==1"],
        valid_responses: ["rate==1"],
        measured_requests: [`count>=${rps * duration - 2}`, `count<=${rps * duration + 2}`],
        "measured_requests{endpoint:airports}": [`count>=${rps / 2 * duration - 1}`, `count<=${rps / 2 * duration + 1}`],
        "measured_requests{endpoint:flights}": [`count>=${rps / 2 * duration - 1}`, `count<=${rps / 2 * duration + 1}`],
        dropped_iterations: ["count==0"],
        backend_unknown_share: ["rate==0"],
        ...(checkBalance ? {
            backend_main_share: ["rate>=0.45", "rate<=0.55"],
            backend_reader_1_share: ["rate>=0.20", "rate<=0.30"],
            backend_reader_2_share: ["rate>=0.20", "rate<=0.30"],
        } : {}),
    },
};

function page(response) {
    try {
        const body = response.json();
        if (!body || !Array.isArray(body.items) || body.items.length > limit ||
            typeof body.has_more !== "boolean" ||
            !(body.next_cursor === null || typeof body.next_cursor === "string") ||
            (body.has_more && !body.next_cursor) || (!body.has_more && body.next_cursor !== null) ||
            !body.items.every((item) => item && typeof item.id === "string")) {
            return null;
        }
        return body;
    } catch (_) {
        return null;
    }
}

export function setup() {
    for (const endpoint of ["airports", "flights"]) {
        const response = http.get(`${baseURL}/api/v1/${endpoint}?limit=${limit}`, {
            timeout: "5s", redirects: 0, tags: { phase: "setup", endpoint },
        });
        const body = page(response);
        if (response.status !== 200 || !body || body.items.length === 0) {
            fail(`Preflight failed for ${endpoint}: expected 200 and a nonempty valid page`);
        }
    }
}

function request(endpoint) {
    const phase = exec.scenario.name.endsWith("_load") ? "load" : "warmup";
    const nextPage = sequences[endpoint]++ % 5 === 4 && cursors[endpoint];
    const query = `limit=${limit}${nextPage ? `&cursor=${encodeURIComponent(cursors[endpoint])}` : ""}`;
    const response = http.get(`${baseURL}/api/v1/${endpoint}?${query}`, {
        timeout: "5s", redirects: 0,
        tags: { name: `GET /api/v1/${endpoint}`, phase, endpoint },
    });
    const body = page(response);
    const backend = response.headers["X-Backend-Instance"];
    const knownBackend = ["api", "api-read-1", "api-read-2"].includes(backend);
    check(response, {
        "HTTP 200": (r) => r.status === 200,
        "valid cursor page": () => body !== null,
        "known backend instance": () => knownBackend,
    }, { phase, endpoint });
    cursors[endpoint] = body ? body.next_cursor : null;

    if (phase !== "load") return;
    measuredRequests.add(1, { endpoint });
    nextPageRequests.add(nextPage ? 1 : 0, { endpoint });
    (endpoint === "airports" ? airportsLatency : flightsLatency).add(response.timings.duration);
    validResponses.add(response.status === 200 && body !== null && knownBackend);
    mainShare.add(backend === "api");
    reader1Share.add(backend === "api-read-1");
    reader2Share.add(backend === "api-read-2");
    unknownShare.add(!knownBackend);
    mainRequests.add(backend === "api" ? 1 : 0);
    reader1Requests.add(backend === "api-read-1" ? 1 : 0);
    reader2Requests.add(backend === "api-read-2" ? 1 : 0);
}

export function airports() { request("airports"); }
export function flights() { request("flights"); }

export function handleSummary(data) {
    const result = {
        configuration: {
            base_url: baseURL,
            rate: rps,
            duration_seconds: duration,
            warmup_seconds: warmup,
            limit, check_balance: checkBalance,
            recorded_at: new Date().toISOString(),
        },
        ...data,
    };
    return {
        [summaryPath]: JSON.stringify(result, null, 2),
        stdout: `\nSummary saved to ${summaryPath}\n${JSON.stringify({
            requests: data.metrics.measured_requests.values,
            airports: data.metrics.airports_latency.values,
            flights: data.metrics.flights_latency.values,
            http_failures: data.metrics["http_req_failed{phase:load}"].values,
            backend_shares: {
                api: data.metrics.backend_main_share.values.rate,
                reader_1: data.metrics.backend_reader_1_share.values.rate,
                reader_2: data.metrics.backend_reader_2_share.values.rate,
            },
            failed_thresholds: Object.entries(data.metrics).flatMap(([metric, value]) =>
                Object.entries(value.thresholds || {}).filter(([, result]) => !result.ok)
                    .map(([threshold]) => ({ metric, threshold }))),
        }, null, 2)}\n`,
    };
}
