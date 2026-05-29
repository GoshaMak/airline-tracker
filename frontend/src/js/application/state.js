import { decodeTokenPayload, normalizeRole } from "../domain/auth.js";

export function createState(session = {}) {
  const tokenRole = normalizeRole(decodeTokenPayload(session.token).role);
  return {
    token: session.token || "",
    userEmail: session.userEmail || "",
    userRole: tokenRole || session.userRole || "",
    flightView: "all",
    flights: [],
    airports: [],
    aircrafts: [],
    aircraftModels: {},
    gates: [],
    subscriptions: [],
    pendingSubscriptionIds: new Set(),
  };
}

export function applyToken(state, token, email = state.userEmail) {
  const payload = decodeTokenPayload(token);

  state.token = token;
  state.userEmail = email;
  state.userRole = normalizeRole(payload.role);
}

export function clearToken(state) {
  state.token = "";
  state.userEmail = "";
  state.userRole = "";
  state.aircrafts = [];
  state.aircraftModels = {};
  state.gates = [];
  state.subscriptions = [];
  state.pendingSubscriptionIds.clear();
  state.flightView = "all";
}

export function setFlightView(state, view) {
  state.flightView = view === "subscribed" && state.token && state.userRole === "user" ? "subscribed" : "all";
}
