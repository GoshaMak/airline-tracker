export function friendlyErrorMessage(error, fallback = "Something went wrong. Please try again.") {
  if (error?.isTimeout || error?.name === "AbortError") {
    return "The backend did not respond in time. Please try again.";
  }

  if (error?.name === "TypeError") {
    return "Cannot reach the backend. Check that the server is running and try again.";
  }

  const status = error?.status;
  const raw = String(error?.message || "").trim().toLowerCase();

  if (raw.includes("email is used")) return "This email is already registered.";
  if (raw.includes("user not found")) return "Email or password is incorrect.";
  if (raw.includes("user already subscribed")) return "You are already subscribed to this flight.";
  if (raw.includes("flight not found")) return "This flight is no longer available.";
  if (raw.includes("bad request") || raw.includes("failed to parse args") || raw.includes("invalid")) {
    return "Check the entered values and try again.";
  }
  if (raw.includes("internal error") || status >= 500) {
    return "The backend had a problem. Please try again later.";
  }
  if (status === 401 || status === 403) {
    return "Your session is not allowed to do that. Sign in with a user account.";
  }
  if (status === 404) {
    return "The requested resource was not found.";
  }

  return raw ? error.message : fallback;
}
