const API_BASE = "/api";

const state = {
  token: localStorage.getItem("airline_tracker_token") || "",
  userEmail: localStorage.getItem("airline_tracker_email") || "",
  userRole: normalizeRole(localStorage.getItem("airline_tracker_role")),
  flightView: "all",
  flights: [],
  airports: [],
  subscriptions: [],
  pendingSubscriptionIds: new Set(),
};

const els = {
  loginForm: document.getElementById("loginForm"),
  registerForm: document.getElementById("registerForm"),
  logoutButton: document.getElementById("logoutButton"),
  sessionState: document.getElementById("sessionState"),
  authMessage: document.getElementById("authMessage"),
  airportSearch: document.getElementById("airportSearch"),
  airportList: document.getElementById("airportList"),
  airportCount: document.getElementById("airportCount"),
  totalFlights: document.getElementById("totalFlights"),
  activeFlights: document.getElementById("activeFlights"),
  delayedFlights: document.getElementById("delayedFlights"),
  cancelledFlights: document.getElementById("cancelledFlights"),
  subscribedFlights: document.getElementById("subscribedFlights"),
  flightRows: document.getElementById("flightRows"),
  flightMeta: document.getElementById("flightMeta"),
  flightSearch: document.getElementById("flightSearch"),
  statusFilter: document.getElementById("statusFilter"),
  refreshButton: document.getElementById("refreshButton"),
  flightViewTabs: document.getElementById("flightViewTabs"),
};

function setMessage(text, type = "") {
  els.authMessage.textContent = text;
  els.authMessage.className = `message ${type}`.trim();
}

function setServerStatus(ok, text) {
}

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function decodeTokenPayload(token) {
  try {
    const payload = token.split(".")[1];
    const normalized = payload.replaceAll("-", "+").replaceAll("_", "/");
    const padded = normalized.padEnd(normalized.length + ((4 - (normalized.length % 4)) % 4), "=");
    return JSON.parse(atob(padded));
  } catch {
    return {};
  }
}

function normalizeRole(role) {
  if (role === 0 || role === "0" || role === "user") return "user";
  if (role === 1 || role === "1" || role === "admin") return "admin";
  return "";
}

function applyToken(token, email = state.userEmail) {
  const payload = decodeTokenPayload(token);
  state.token = token;
  state.userEmail = email;
  state.userRole = normalizeRole(payload.role);

  localStorage.setItem("airline_tracker_token", state.token);
  localStorage.setItem("airline_tracker_email", state.userEmail);
  localStorage.setItem("airline_tracker_role", state.userRole);
}

function clearToken() {
  state.token = "";
  state.userEmail = "";
  state.userRole = "";
  state.subscriptions = [];
  state.pendingSubscriptionIds.clear();
  state.flightView = "all";

  localStorage.removeItem("airline_tracker_token");
  localStorage.removeItem("airline_tracker_email");
  localStorage.removeItem("airline_tracker_role");
}

async function request(path, options = {}) {
  const headers = {
    Accept: "application/json",
    ...(options.body ? { "Content-Type": "application/json" } : {}),
    ...(state.token ? { Authorization: `Bearer ${state.token}` } : {}),
    ...options.headers,
  };

  const response = await fetch(`${API_BASE}${path}`, { ...options, headers });
  const contentType = response.headers.get("content-type") || "";
  const body = contentType.includes("application/json") ? await response.json() : await response.text();

  if (!response.ok) {
    const msg = typeof body === "object" ? body.msg || body.err || JSON.stringify(body) : body;
    throw new Error(msg || `Request failed with ${response.status}`);
  }

  return body;
}

function formatDate(value) {
  if (!value) return "not set";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "not set";
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function shortId(value) {
  return String(value || "").slice(0, 8);
}

function airportLabel(id) {
  const airport = state.airports.find((item) => item.id === id);
  if (!airport) return shortId(id);
  return `${airport.iata_code} · ${airport.city}`;
}

function badgeClass(status) {
  if (["cancelled"].includes(status)) return "stop";
  if (["delayed", "rescheduled"].includes(status)) return "warn";
  if (["departed", "landed", "boarding"].includes(status)) return "info";
  return "ok";
}

function filteredFlights() {
  const needle = els.flightSearch.value.trim().toLowerCase();
  const status = els.statusFilter.value;
  const source = state.flightView === "subscribed" ? state.subscriptions : state.flights;

  return source.filter((flight) => {
    const text = [
      flight.id,
      flight.status,
      flight.plan,
      airportLabel(flight.departure_airport_id),
      airportLabel(flight.arrival_airport_id),
    ]
      .join(" ")
      .toLowerCase();

    return (!status || flight.status === status) && (!needle || text.includes(needle));
  });
}

function renderMetrics() {
  els.totalFlights.textContent = state.flights.length;
  els.activeFlights.textContent = state.flights.filter((f) =>
    ["boarding", "departed", "landed"].includes(f.status)
  ).length;
  els.delayedFlights.textContent = state.flights.filter((f) => f.status === "delayed").length;
  els.cancelledFlights.textContent = state.flights.filter((f) => f.status === "cancelled").length;
  els.subscribedFlights.textContent = state.subscriptions.length;
}

function renderAirports() {
  const needle = els.airportSearch.value.trim().toLowerCase();
  const airports = state.airports
    .filter((airport) =>
      [airport.iata_code, airport.title, airport.city, airport.country].join(" ").toLowerCase().includes(needle)
    )
    .slice(0, 80);

  els.airportCount.textContent = state.airports.length;
  els.airportList.innerHTML =
    airports
      .map(
        (airport) => `
          <article class="airport-item">
            <strong>${escapeHtml(airport.iata_code)} · ${escapeHtml(airport.title)}</strong>
            <span>${escapeHtml(airport.city)}, ${escapeHtml(airport.country)}</span>
          </article>
        `
      )
      .join("") || `<div class="airport-item"><span>No airports found</span></div>`;
}

function renderFlights() {
  const flights = filteredFlights();
  const sourceTotal = state.flightView === "subscribed" ? state.subscriptions.length : state.flights.length;
  els.flightMeta.textContent = `${flights.length} of ${sourceTotal} ${state.flightView} flights shown`;

  els.flightRows.innerHTML =
    flights
      .slice(0, 250)
      .map((flight) => {
        const subscribed = state.subscriptions.some((item) => item.id === flight.id);
        const pending = state.pendingSubscriptionIds.has(flight.id);
        const canSubscribe = Boolean(state.token) && state.userRole === "user" && !subscribed && !pending;
        const actionLabel = (() => {
          if (pending) return "Subscribing";
          if (subscribed) return "Subscribed";
          if (!state.token) return "Sign in";
          if (state.userRole !== "user") return "User only";
          return "Subscribe";
        })();
        return `
          <tr>
            <td>
              <strong class="mono">${escapeHtml(shortId(flight.id))}</strong>
              <div class="muted-text mono">${escapeHtml(shortId(flight.aircraft_id))}</div>
            </td>
            <td>
              <strong>${escapeHtml(airportLabel(flight.departure_airport_id))}</strong>
              <div class="muted-text">${escapeHtml(airportLabel(flight.arrival_airport_id))}</div>
            </td>
            <td><span class="badge ${badgeClass(flight.status)}">${escapeHtml(flight.status)}</span></td>
            <td>
              <strong>${escapeHtml(formatDate(flight.scheduled_departure))}</strong>
              <div class="muted-text">${escapeHtml(formatDate(flight.scheduled_arrival))}</div>
            </td>
            <td>
              <strong>${escapeHtml(formatDate(flight.actual_departure))}</strong>
              <div class="muted-text">${escapeHtml(formatDate(flight.actual_arrival))}</div>
            </td>
            <td>${escapeHtml(flight.plan || "none")}</td>
            <td class="action-cell">
              <button class="subscribe-button" type="button" data-flight-id="${escapeHtml(flight.id)}" ${canSubscribe ? "" : "disabled"}>
                ${escapeHtml(actionLabel)}
              </button>
            </td>
          </tr>
        `;
      })
      .join("") || `<tr><td class="empty-row" colspan="7">No flights found</td></tr>`;
}

function renderAuthState() {
  els.logoutButton.hidden = !state.token;
  els.sessionState.textContent = state.token
    ? `Signed in${state.userEmail ? ` as ${state.userEmail}` : ""}${state.userRole ? ` (${state.userRole})` : ""}`
    : "Signed out";
  els.sessionState.className = state.token ? "session-state signed-in" : "session-state";
  els.flightViewTabs.querySelector('[data-view="subscribed"]').disabled = !state.token || state.userRole !== "user";
}

function renderFlightViewTabs() {
  els.flightViewTabs.querySelectorAll("button").forEach((button) => {
    button.classList.toggle("active", button.dataset.view === state.flightView);
  });
}

function setFlightView(view) {
  state.flightView = view === "subscribed" && state.token && state.userRole === "user" ? "subscribed" : "all";
  renderFlightViewTabs();
  renderFlights();
}

async function loadStatus() {
  try {
    await request("/status");
    setServerStatus(true, "Backend online");
  } catch (error) {
    setServerStatus(false, "Backend offline");
  }
}

async function loadFlights() {
  const response = await request("/list_flights");
  state.flights = response.flights || [];
}

async function loadAirports() {
  const response = await request("/airport/list");
  state.airports = response.airports || [];
}

async function loadSubscriptions() {
  if (!state.token || state.userRole !== "user") {
    state.subscriptions = [];
    return;
  }

  try {
    const response = await request("/user/list_flights");
    state.subscriptions = response.flights || [];
  } catch {
    state.subscriptions = [];
  }
}

async function refreshData() {
  els.refreshButton.disabled = true;
  try {
    await Promise.all([loadStatus(), loadAirports(), loadFlights(), loadSubscriptions()]);
    renderMetrics();
    renderAirports();
    renderAuthState();
    renderFlightViewTabs();
    renderFlights();
  } catch (error) {
    setServerStatus(false, "Backend offline");
  } finally {
    els.refreshButton.disabled = false;
  }
}

async function login(event) {
  event.preventDefault();
  const formEl = event.currentTarget;
  const form = new FormData(formEl);
  const email = String(form.get("email") || "");

  try {
    const response = await request("/login", {
      method: "POST",
      body: JSON.stringify({
        email,
        password: form.get("password"),
      }),
    });
    applyToken(response.token, email);
    setMessage(
      state.userRole === "user" ? "Signed in. You can subscribe to flights." : "Signed in. Use a user account to subscribe.",
      state.userRole === "user" ? "success" : ""
    );
    formEl.reset();
    renderAuthState();
    await refreshData();
  } catch (error) {
    setMessage(error.message, "error");
  }
}

async function register(event) {
  event.preventDefault();
  const formEl = event.currentTarget;
  const form = new FormData(formEl);

  try {
    await request("/register", {
      method: "POST",
      body: JSON.stringify({
        email: form.get("email"),
        password: form.get("password"),
        role: form.get("role"),
      }),
    });
    setMessage("Account created", "success");
    formEl.reset();
  } catch (error) {
    setMessage(error.message, "error");
  }
}

async function subscribe(flightId) {
  if (!state.token) {
    setMessage("Sign in with a user account to subscribe", "error");
    return;
  }

  if (state.userRole !== "user") {
    setMessage("Only user accounts can subscribe to flights", "error");
    return;
  }

  state.pendingSubscriptionIds.add(flightId);
  renderFlights();
  try {
    await request(`/user/subscribe?flight_id=${encodeURIComponent(flightId)}`, { method: "POST" });
    const flight = state.flights.find((item) => item.id === flightId);
    if (flight && !state.subscriptions.some((item) => item.id === flightId)) {
      state.subscriptions = [...state.subscriptions, flight];
    }
    setMessage("Subscribed to flight", "success");
    await loadSubscriptions();
    renderMetrics();
    renderFlights();
  } catch (error) {
    setMessage(error.message, "error");
  } finally {
    state.pendingSubscriptionIds.delete(flightId);
    renderFlights();
  }
}

els.loginForm.addEventListener("submit", login);
els.registerForm.addEventListener("submit", register);
els.logoutButton.addEventListener("click", () => {
  clearToken();
  setMessage("Signed out");
  renderAuthState();
  renderMetrics();
  renderFlightViewTabs();
  renderFlights();
});
els.refreshButton.addEventListener("click", refreshData);
els.airportSearch.addEventListener("input", renderAirports);
els.flightSearch.addEventListener("input", renderFlights);
els.statusFilter.addEventListener("change", renderFlights);
els.flightViewTabs.addEventListener("click", (event) => {
  const button = event.target.closest("button[data-view]");
  if (button) setFlightView(button.dataset.view);
});
els.flightRows.addEventListener("click", (event) => {
  const button = event.target.closest("button[data-flight-id]");
  if (button) subscribe(button.dataset.flightId);
});

if (state.token && !state.userRole) {
  state.userRole = normalizeRole(decodeTokenPayload(state.token).role);
}
renderAuthState();
renderFlightViewTabs();
refreshData();
