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
  const normalized = typeof role === "string" ? role.trim().toLowerCase() : role;
  if (role === 0 || role === "0" || role === "user") return "user";
  if (role === 1 || normalized === "1" || normalized === "admin") return "admin";
  if (normalized === "user") return "user";
  return "";
}

export function isUserRole(role) {
  return role === "user";
}

export function isAdminRole(role) {
  return role === "admin";
}
