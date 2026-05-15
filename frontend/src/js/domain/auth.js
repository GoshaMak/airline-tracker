export function decodeTokenPayload(token) {
  try {
    const payload = token.split(".")[1];
    const normalized = payload.replaceAll("-", "+").replaceAll("_", "/");
    const padded = normalized.padEnd(normalized.length + ((4 - (normalized.length % 4)) % 4), "=");
    return JSON.parse(atob(padded));
  } catch {
    return {};
  }
}

export function normalizeRole(role) {
  if (role === 0 || role === "0" || role === "user") return "user";
  if (role === 1 || role === "1" || role === "admin") return "admin";
  return "";
}

export function isUserRole(role) {
  return role === "user";
}
