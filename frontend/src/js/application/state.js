import { decodeTokenPayload, normalizeRole } from "../domain/auth.js";

export function createState(session = {}) {
  return {
    token: session.token || "",
    userEmail: session.userEmail || "",
    userRole: session.userRole || normalizeRole(decodeTokenPayload(session.token).role),
    flightView: "all",
    flights: [],
    airports: [],
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
  state.subscriptions = [];
  state.pendingSubscriptionIds.clear();
  state.flightView = "all";
}

export function setFlightView(state, view) {
  state.flightView = view === "subscribed" && state.token && state.userRole === "user" ? "subscribed" : "all";
}
