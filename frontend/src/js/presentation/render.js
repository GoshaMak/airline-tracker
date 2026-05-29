import { isAdminRole, isUserRole } from "../domain/auth.js";
import { formatDate, shortId } from "../utils/format.js";
import { escapeHtml } from "../utils/html.js";

function airportLabel(state, id) {
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

function filteredFlights(els, state) {
  const needle = els.flightSearch.value.trim().toLowerCase();
  const status = els.statusFilter.value;
  const source = state.flightView === "subscribed" ? state.subscriptions : state.flights;

  return source.filter((flight) => {
    const text = [
      flight.id,
      flight.status,
      flight.plan,
      airportLabel(state, flight.departure_airport_id),
      airportLabel(state, flight.arrival_airport_id),
    ]
      .join(" ")
      .toLowerCase();

    return (!status || flight.status === status) && (!needle || text.includes(needle));
  });
}

function flightLabel(state, flight) {
  return `${shortId(flight.id)} · ${airportLabel(state, flight.departure_airport_id)} to ${airportLabel(
    state,
    flight.arrival_airport_id
  )}`;
}

function aircraftId(aircraft) {
  return aircraft.id || "";
}

function aircraftModelId(aircraft) {
  return aircraft.aircraft_model_id || "";
}

function aircraftLabel(state, aircraft) {
  const model = state.aircraftModels[aircraftModelId(aircraft)];
  const registration = aircraft.registration_number || shortId(aircraftId(aircraft));
  const serial = aircraft.serial_number;
  const manufacturer = model?.manufacturer;
  const modelName = model?.model;
  const modelLabel = model ? [manufacturer, modelName].filter(Boolean).join(" ") : shortId(aircraftModelId(aircraft));
  return [registration, modelLabel, serial ? `S/N ${serial}` : ""].filter(Boolean).join(" · ");
}

function gateId(gate) {
  return gate.id || "";
}

function gateAirportId(gate) {
  return gate.airport_id || "";
}

function gateLabel(state, gate) {
  const number = gate.number || shortId(gateId(gate));
  return `${number} · ${airportLabel(state, gateAirportId(gate))}`;
}

function uniqueOptions(options) {
  const seen = new Set();
  return options.filter((option) => {
    if (!option.value || seen.has(option.value)) return false;
    seen.add(option.value);
    return true;
  });
}

function setSelectOptions(select, options, placeholder) {
  if (!select) return;
  const selectedValue = select.value;
  select.innerHTML = [
    placeholder ? `<option value="">${escapeHtml(placeholder)}</option>` : "",
    ...options.map(
      (option) =>
        `<option value="${escapeHtml(option.value)}">${escapeHtml(option.label)}</option>`
    ),
  ].join("");

  if (selectedValue && options.some((option) => option.value === selectedValue)) {
    select.value = selectedValue;
  }
}

function gateOptionsForAirport(state, airportId) {
  return uniqueOptions(
    state.gates
      .filter((gate) => gateId(gate) && (!airportId || gateAirportId(gate) === airportId))
      .map((gate) => ({
        value: gateId(gate),
        label: gateLabel(state, gate),
      }))
  );
}

export function renderMetrics(els, state) {
  els.totalFlights.textContent = state.flights.length;
  els.activeFlights.textContent = state.flights.filter((f) =>
    ["boarding", "departed", "landed"].includes(f.status)
  ).length;
  els.delayedFlights.textContent = state.flights.filter((f) => f.status === "delayed").length;
  els.cancelledFlights.textContent = state.flights.filter((f) => f.status === "cancelled").length;
  els.subscribedFlights.textContent = state.subscriptions.length;
}

export function renderAirports(els, state) {
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

export function renderFlights(els, state) {
  const flights = filteredFlights(els, state);
  const sourceTotal = state.flightView === "subscribed" ? state.subscriptions.length : state.flights.length;
  els.flightMeta.textContent = `${flights.length} of ${sourceTotal} ${state.flightView} flights shown`;

  els.flightRows.innerHTML =
    flights
      .slice(0, 250)
      .map((flight) => {
        const subscribed = state.subscriptions.some((item) => item.id === flight.id);
        const pending = state.pendingSubscriptionIds.has(flight.id);
        const canSubscribe = Boolean(state.token) && isUserRole(state.userRole) && !subscribed && !pending;
        const canEdit = Boolean(state.token) && isAdminRole(state.userRole);
        const actionLabel = (() => {
          if (canEdit) return "Edit";
          if (pending) return "Subscribing";
          if (subscribed) return "Subscribed";
          if (!state.token) return "Sign in";
          if (!isUserRole(state.userRole)) return "User only";
          return "Subscribe";
        })();

        return `
          <tr>
            <td>
              <strong class="mono">${escapeHtml(shortId(flight.id))}</strong>
              <div class="muted-text mono">${escapeHtml(shortId(flight.aircraft_id))}</div>
            </td>
            <td>
              <strong>${escapeHtml(airportLabel(state, flight.departure_airport_id))}</strong>
              <div class="muted-text">${escapeHtml(airportLabel(state, flight.arrival_airport_id))}</div>
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
              <button class="subscribe-button" type="button" ${
                canEdit ? `data-admin-edit-flight-id="${escapeHtml(flight.id)}"` : `data-flight-id="${escapeHtml(flight.id)}"`
              } ${canEdit || canSubscribe ? "" : "disabled"}>
                ${escapeHtml(actionLabel)}
              </button>
            </td>
          </tr>
        `;
      })
      .join("") || `<tr><td class="empty-row" colspan="7">No flights found</td></tr>`;
}

export function renderAuthState(els, state) {
  els.logoutButton.hidden = !state.token;
  els.sessionState.textContent = state.token
    ? `Signed in${state.userEmail ? ` as ${state.userEmail}` : ""}${state.userRole ? ` (${state.userRole})` : ""}`
    : "Signed out";
  els.sessionState.className = state.token ? "session-state signed-in" : "session-state";
  els.flightViewTabs.querySelector('[data-view="subscribed"]').disabled = !state.token || !isUserRole(state.userRole);
  if (els.adminPanel) {
    els.adminPanel.hidden = !state.token || !isAdminRole(state.userRole);
  }
}

export function renderFlightViewTabs(els, state) {
  els.flightViewTabs.querySelectorAll("button").forEach((button) => {
    button.classList.toggle("active", button.dataset.view === state.flightView);
  });
}

export function renderAdminPanel(els, state) {
  if (!els.adminPanel) return;

  const aircraftOptions = uniqueOptions([
    ...state.aircrafts
      .filter((aircraft) => aircraftId(aircraft))
      .map((aircraft) => ({
        value: aircraftId(aircraft),
        label: aircraftLabel(state, aircraft),
      })),
  ]);
  const airportOptions = state.airports.map((airport) => ({
    value: airport.id,
    label: `${airport.iata_code} · ${airport.city}`,
  }));
  const createDepartureAirportId = els.adminCreateDepartureAirport.value;
  const createArrivalAirportId = els.adminCreateArrivalAirport.value;
  const createDepartureGateOptions = gateOptionsForAirport(state, createDepartureAirportId);
  const createArrivalGateOptions = gateOptionsForAirport(state, createArrivalAirportId);
  setSelectOptions(els.adminCreateAircraft, aircraftOptions, "Select aircraft");
  setSelectOptions(els.adminCreateDepartureAirport, airportOptions, "Select departure airport");
  setSelectOptions(els.adminCreateArrivalAirport, airportOptions, "Select arrival airport");
  setSelectOptions(els.adminCreateDepartureGate, createDepartureGateOptions, "Select departure gate");
  setSelectOptions(els.adminCreateArrivalGate, createArrivalGateOptions, "Select arrival gate");

  const flightOptions = state.flights.map((flight) => ({
    value: flight.id,
    label: flightLabel(state, flight),
  }));
  setSelectOptions(els.adminEditFlightSelect, flightOptions, "Select flight to edit");
}

export function renderAll(els, state) {
  renderMetrics(els, state);
  renderAirports(els, state);
  renderAuthState(els, state);
  renderFlightViewTabs(els, state);
  renderAdminPanel(els, state);
  renderFlights(els, state);
}
