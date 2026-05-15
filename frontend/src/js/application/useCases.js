import { REGISTER_ROLE } from "../config.js";
import { isUserRole } from "../domain/auth.js";
import { applyToken, clearToken } from "./state.js";

export async function loadFlights({ api, state }) {
  const response = await api.request("/list_flights");
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

  try {
    const response = await api.request("/user/list_flights");
    state.subscriptions = response.flights || [];
  } catch {
    state.subscriptions = [];
  }
}

export async function refreshData(context) {
  await Promise.all([loadAirports(context), loadFlights(context), loadSubscriptions(context)]);
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
