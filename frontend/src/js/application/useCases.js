import { REGISTER_ROLE } from "../config.js";
import { isAdminRole, isUserRole } from "../domain/auth.js";
import { applyToken, clearToken } from "./state.js";

export async function loadFlights({ api, state }) {
  const response = await api.request("/flight/list");
  state.flights = response.flights || [];
}

export async function loadAirports({ api, state }) {
  const response = await api.request("/airport/list");
  state.airports = response.airports || [];
}

export async function loadSubscriptions({ api, state }) {
  if (!state.token || !isUserRole(state.userRole)) {
    state.subscriptions = [];
    return;
  }

  const response = await api.request("/user/flight/list");
  state.subscriptions = response.flights || [];
}

function aircraftModelId(aircraft) {
  return aircraft.aircraft_model_id || "";
}

export async function loadAircrafts({ api, state }) {
  if (!state.token || !isAdminRole(state.userRole)) {
    state.aircrafts = [];
    state.aircraftModels = {};
    return;
  }

  const response = await api.request("/admin/aircraft/list");
  const aircrafts = response.aircrafts || [];
  state.aircrafts = aircrafts;

  const modelIds = [...new Set(aircrafts.map((aircraft) => aircraftModelId(aircraft)).filter(Boolean))];
  const modelEntries = await Promise.all(
    modelIds.map(async (id) => {
      const model = await api.request(`/admin/aircraft_model/${encodeURIComponent(id)}`);
      return [id, model];
    })
  );
  state.aircraftModels = Object.fromEntries(modelEntries);
}

export async function loadGates({ api, state }) {
  if (!state.token || !isAdminRole(state.userRole)) {
    state.gates = [];
    return;
  }

  const response = await api.request("/gate/list");
  state.gates = response.gates || [];
}

export async function loadAdminResources(context) {
  if (!context.state.token || !isAdminRole(context.state.userRole)) return;
  await Promise.all([loadAircrafts(context), loadGates(context)]);
}

export async function refreshData(context) {
  const results = await Promise.allSettled([
    loadAirports(context),
    loadFlights(context),
    loadSubscriptions(context),
    loadAdminResources(context),
  ]);
  const failed = results.find((result) => result.status === "rejected");
  if (failed) throw failed.reason;
}

export async function loginUser({ api, email, password, sessionStore, state }) {
  const response = await api.request("/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
  applyToken(state, response.token, email);
  sessionStore?.saveSession({
    token: state.token,
    userEmail: state.userEmail,
    userRole: state.userRole,
  });
}

export async function registerUser({ api, email, password, role = REGISTER_ROLE }) {
  await api.request("/register", {
    method: "POST",
    body: JSON.stringify({ email, password, role }),
  });
}

export function logoutUser({ sessionStore, state }) {
  clearToken(state);
  sessionStore?.clearSession();
}

export async function subscribeToFlight({ api, flightId, state }) {
  await api.request(`/user/subscribe?flight_id=${encodeURIComponent(flightId)}`, { method: "POST" });

  const flight = state.flights.find((item) => item.id === flightId);
  if (flight && !state.subscriptions.some((item) => item.id === flightId)) {
    state.subscriptions = [...state.subscriptions, flight];
  }

  await loadSubscriptions({ api, state });
}

export async function createFlight({ api, payload }) {
  await api.request("/flight/create", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateFlight({ api, flightId, payload }) {
  await api.request(`/flight/${encodeURIComponent(flightId)}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}
