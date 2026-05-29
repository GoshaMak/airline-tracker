import { REGISTER_ROLE } from "../config.js";
import { isAdminRole, isUserRole } from "../domain/auth.js";
import { setFlightView } from "../application/state.js";
import {
  createFlight,
  loadFlights,
  loginUser,
  logoutUser,
  refreshData,
  registerUser,
  subscribeToFlight,
  updateFlight,
} from "../application/useCases.js";
import { friendlyErrorMessage } from "./errors.js";
import { dateTimeInputValue, dateTimeLocalToUtcISOString } from "../utils/format.js";
import {
  renderAll,
  renderAdminPanel,
  renderAuthState,
  renderFlightViewTabs,
  renderFlights,
  renderMetrics,
} from "./render.js";

const EMPTY_VALUE = "";
const DATE_TIME_FIELDS = ["scheduled_departure", "scheduled_arrival", "actual_departure", "actual_arrival"];

function formValue(form, name) {
  return String(form.get(name) || "").trim();
}

function optionalDateTime(value) {
  return value ? dateTimeLocalToUtcISOString(value) : null;
}

function optionalString(value) {
  const trimmed = String(value || "").trim();
  return trimmed ? trimmed : null;
}

function buildBaseFlightPayload(form) {
  return {
    scheduled_departure: dateTimeLocalToUtcISOString(formValue(form, "scheduled_departure")),
    scheduled_arrival: dateTimeLocalToUtcISOString(formValue(form, "scheduled_arrival")),
    actual_departure: optionalDateTime(formValue(form, "actual_departure")),
    actual_arrival: optionalDateTime(formValue(form, "actual_arrival")),
    status: formValue(form, "status"),
    plan: optionalString(formValue(form, "plan")),
  };
}

function buildCreateFlightPayload(formEl) {
  const form = new FormData(formEl);
  return {
    flight: buildBaseFlightPayload(form),
    aircraft_id: formValue(form, "aircraft_id"),
    departure_airport_id: formValue(form, "departure_airport_id"),
    arrival_airport_id: formValue(form, "arrival_airport_id"),
    departure_gate_id: formValue(form, "departure_gate_id"),
    arrival_gate_id: formValue(form, "arrival_gate_id"),
  };
}

function comparableFormValue(form, key) {
  if (DATE_TIME_FIELDS.includes(key)) return formValue(form, key);
  if (key === "plan") return optionalString(formValue(form, key));
  return formValue(form, key);
}

function comparableFlightValue(flight, key) {
  if (!flight) return undefined;
  if (DATE_TIME_FIELDS.includes(key)) return dateTimeInputValue(flight[key]);
  if (key === "plan") return optionalString(flight[key]);
  return flight[key] ?? EMPTY_VALUE;
}

function buildUpdateFlightPayload(formEl, flightId, state) {
  const form = new FormData(formEl);
  const currentFlight = state.flights.find((item) => item.id === flightId);
  const nextFlight = buildBaseFlightPayload(form);
  const flight = {};

  Object.entries(nextFlight).forEach(([key, value]) => {
    if (comparableFormValue(form, key) !== comparableFlightValue(currentFlight, key)) {
      flight[key] = value;
    }
  });

  return {
    flight,
  };
}

function setField(formEl, name, value) {
  const field = formEl.elements[name];
  if (field) field.value = value || EMPTY_VALUE;
}

function populateEditForm(els, state) {
  if (!els.adminEditFlightForm || !els.adminEditFlightSelect) return;
  const flight = state.flights.find((item) => item.id === els.adminEditFlightSelect.value);
  if (!flight) {
    els.adminEditFlightForm.reset();
    return;
  }

  setField(els.adminEditFlightForm, "scheduled_departure", dateTimeInputValue(flight.scheduled_departure));
  setField(els.adminEditFlightForm, "scheduled_arrival", dateTimeInputValue(flight.scheduled_arrival));
  setField(els.adminEditFlightForm, "actual_departure", dateTimeInputValue(flight.actual_departure));
  setField(els.adminEditFlightForm, "actual_arrival", dateTimeInputValue(flight.actual_arrival));
  setField(els.adminEditFlightForm, "status", flight.status);
  setField(els.adminEditFlightForm, "plan", flight.plan);
}

function openFlightEditor(flightId, els, state) {
  renderAdminPanel(els, state);
  els.adminEditFlightSelect.value = flightId;
  populateEditForm(els, state);
  els.adminPanel.scrollIntoView({ behavior: "smooth", block: "start" });
  els.adminEditFlightSelect.focus();
}

export async function refreshDataAndRender({ api, els, messages, state }) {
  els.refreshButton.disabled = true;
  try {
    await refreshData({ api, state });
    renderAll(els, state);
  } catch (error) {
    renderAll(els, state);
    messages.show(friendlyErrorMessage(error, "Could not refresh flight data."), "error");
  } finally {
    els.refreshButton.disabled = false;
  }
}

async function refreshAfterAdminAction(context, fallbackMessage) {
  try {
    await loadFlights(context);
    renderAll(context.els, context.state);
  } catch (error) {
    renderAll(context.els, context.state);
    context.messages.show(friendlyErrorMessage(error, fallbackMessage), "error");
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
      isUserRole(state.userRole)
        ? "Signed in. You can subscribe to flights."
        : "Signed in. Admin panel is available.",
      isUserRole(state.userRole) || isAdminRole(state.userRole) ? "success" : ""
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

async function handleCreateFlight(event, context) {
  event.preventDefault();
  event.stopPropagation();
  const { api, els, messages, state } = context;

  if (!isAdminRole(state.userRole)) {
    messages.show("Sign in as admin to manage flights", "error");
    return;
  }

  const submitButton = event.currentTarget.querySelector('button[type="submit"]');
  submitButton.disabled = true;
  try {
    await createFlight({ api, payload: buildCreateFlightPayload(event.currentTarget) });
    messages.show("Flight created", "success");
    event.currentTarget.reset();
  } catch (error) {
    messages.show(friendlyErrorMessage(error, "Could not create flight."), "error");
    submitButton.disabled = false;
    renderAdminPanel(els, state);
    return;
  }

  await refreshAfterAdminAction(context, "Flight created, but could not refresh flight data.");
  submitButton.disabled = false;
  renderAdminPanel(els, state);
}

async function handleUpdateFlight(event, context) {
  event.preventDefault();
  event.stopPropagation();
  const { api, els, messages, state } = context;
  const flightId = els.adminEditFlightSelect.value;

  if (!isAdminRole(state.userRole)) {
    messages.show("Sign in as admin to manage flights", "error");
    return;
  }

  if (!flightId) {
    messages.show("Select a flight to edit", "error");
    return;
  }

  const submitButton = event.currentTarget.querySelector('button[type="submit"]');
  submitButton.disabled = true;
  try {
    const payload = buildUpdateFlightPayload(event.currentTarget, flightId, state);
    if (Object.keys(payload.flight).length === 0) {
      messages.show("No flight changes to update");
      submitButton.disabled = false;
      return;
    }

    await updateFlight({ api, flightId, payload });
    messages.show("Flight updated", "success");
  } catch (error) {
    messages.show(friendlyErrorMessage(error, "Could not update flight."), "error");
    submitButton.disabled = false;
    renderAdminPanel(els, state);
    els.adminEditFlightSelect.value = flightId;
    populateEditForm(els, state);
    return;
  }

  await refreshAfterAdminAction(context, "Flight updated, but could not refresh flight data.");
  submitButton.disabled = false;
  renderAdminPanel(els, state);
  els.adminEditFlightSelect.value = flightId;
  populateEditForm(els, state);
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
  els.adminCreateDepartureAirport.addEventListener("change", () => renderAdminPanel(els, state));
  els.adminCreateArrivalAirport.addEventListener("change", () => renderAdminPanel(els, state));
  els.adminCreateFlightForm.addEventListener("submit", (event) => handleCreateFlight(event, context));
  els.adminEditFlightForm.addEventListener("submit", (event) => handleUpdateFlight(event, context));
  els.adminEditFlightSelect.addEventListener("change", () => populateEditForm(els, state));
  els.flightViewTabs.addEventListener("click", (event) => {
    const button = event.target.closest("button[data-view]");
    if (!button) return;

    setFlightView(state, button.dataset.view);
    renderFlightViewTabs(els, state);
    renderFlights(els, state);
  });
  els.flightRows.addEventListener("click", (event) => {
    const editButton = event.target.closest("button[data-admin-edit-flight-id]");
    if (editButton) {
      openFlightEditor(editButton.dataset.adminEditFlightId, els, state);
      return;
    }

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
