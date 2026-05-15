import { REGISTER_ROLE } from "../config.js";
import { isUserRole } from "../domain/auth.js";
import { setFlightView } from "../application/state.js";
import { loginUser, logoutUser, refreshData, registerUser, subscribeToFlight } from "../application/useCases.js";
import { friendlyErrorMessage } from "./errors.js";
import {
  renderAll,
  renderAuthState,
  renderFlightViewTabs,
  renderFlights,
  renderMetrics,
} from "./render.js";

export async function refreshDataAndRender({ api, els, messages, state }) {
  els.refreshButton.disabled = true;
  try {
    await refreshData({ api, state });
    renderAll(els, state);
  } catch (error) {
    messages.show(friendlyErrorMessage(error, "Could not refresh flight data."), "error");
  } finally {
    els.refreshButton.disabled = false;
  }
}

async function handleLogin(event, context) {
  event.preventDefault();
  event.stopPropagation();
  const { api, els, messages, state } = context;
  const formEl = event.currentTarget;
  const form = new FormData(formEl);
  const email = String(form.get("email") || "");
  const password = String(form.get("password") || "");

  try {
    await loginUser({ api, email, password, sessionStore: context.sessionStore, state });
    messages.show(
      isUserRole(state.userRole) ? "Signed in. You can subscribe to flights." : "Signed in. Use a user account to subscribe.",
      isUserRole(state.userRole) ? "success" : ""
    );
    formEl.reset();
    renderAuthState(els, state);
    await refreshDataAndRender(context);
  } catch (error) {
    messages.show(friendlyErrorMessage(error, "Could not sign in."), "error");
  }
}

async function handleRegister(event, context) {
  event.preventDefault();
  event.stopPropagation();
  const { api, els, messages, state } = context;
  const formEl = event.currentTarget;
  const form = new FormData(formEl);
  const email = String(form.get("email") || "");
  const password = String(form.get("password") || "");
  const role = String(form.get("role") || REGISTER_ROLE);

  try {
    await registerUser({ api, email, password, role });
  } catch (error) {
    messages.show(friendlyErrorMessage(error, "Could not create account."), "error");
    return;
  }

  try {
    await loginUser({ api, email, password, sessionStore: context.sessionStore, state });
    messages.show("Account created and signed in. You can subscribe to flights.", "success");
    formEl.reset();
    renderAuthState(els, state);
    await refreshDataAndRender(context);
  } catch (error) {
    messages.show(
      `Account created, but automatic sign in failed. ${friendlyErrorMessage(error, "Try signing in manually.")}`,
      "error"
    );
  }
}

async function handleSubscribe(flightId, context) {
  const { api, els, messages, state } = context;

  if (!state.token) {
    messages.show("Sign in with a user account to subscribe", "error");
    return;
  }

  if (!isUserRole(state.userRole)) {
    messages.show("Only user accounts can subscribe to flights", "error");
    return;
  }

  state.pendingSubscriptionIds.add(flightId);
  renderFlights(els, state);

  try {
    await subscribeToFlight({ api, flightId, state });
    messages.show("Subscribed to flight", "success");
    renderMetrics(els, state);
    renderFlights(els, state);
  } catch (error) {
    messages.show(friendlyErrorMessage(error, "Could not subscribe to this flight."), "error");
  } finally {
    state.pendingSubscriptionIds.delete(flightId);
    renderFlights(els, state);
  }
}

export function bindEvents(context) {
  const { els, messages, state } = context;

  els.loginForm.addEventListener("submit", (event) => handleLogin(event, context));
  els.registerForm.addEventListener("submit", (event) => handleRegister(event, context));
  els.logoutButton.addEventListener("click", () => {
    logoutUser({ sessionStore: context.sessionStore, state });
    messages.show("Signed out");
    renderAuthState(els, state);
    renderMetrics(els, state);
    renderFlightViewTabs(els, state);
    renderFlights(els, state);
  });
  els.refreshButton.addEventListener("click", () => refreshDataAndRender(context));
  els.airportSearch.addEventListener("input", () => renderAll(els, state));
  els.flightSearch.addEventListener("input", () => renderFlights(els, state));
  els.statusFilter.addEventListener("change", () => renderFlights(els, state));
  els.flightViewTabs.addEventListener("click", (event) => {
    const button = event.target.closest("button[data-view]");
    if (!button) return;

    setFlightView(state, button.dataset.view);
    renderFlightViewTabs(els, state);
    renderFlights(els, state);
  });
  els.flightRows.addEventListener("click", (event) => {
    const button = event.target.closest("button[data-flight-id]");
    if (button) handleSubscribe(button.dataset.flightId, context);
  });
  document.addEventListener("click", () => messages.closeIfReady());
  window.addEventListener("error", () => {
    messages.show("The frontend hit an unexpected error. Reload the page and try again.", "error");
  });
  window.addEventListener("unhandledrejection", () => {
    messages.show("The frontend hit an unexpected error. Reload the page and try again.", "error");
  });
}
