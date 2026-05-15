import { normalizeRole } from "../domain/auth.js";

const TOKEN_KEY = "airline_tracker_token";
const EMAIL_KEY = "airline_tracker_email";
const ROLE_KEY = "airline_tracker_role";

export function loadSession() {
  return {
    token: localStorage.getItem(TOKEN_KEY) || "",
    userEmail: localStorage.getItem(EMAIL_KEY) || "",
    userRole: normalizeRole(localStorage.getItem(ROLE_KEY)),
  };
}

export function saveSession({ token, userEmail, userRole }) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(EMAIL_KEY, userEmail);
  localStorage.setItem(ROLE_KEY, userRole);
}

export function clearSession() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(EMAIL_KEY);
  localStorage.removeItem(ROLE_KEY);
}
