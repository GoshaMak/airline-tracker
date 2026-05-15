export function formatDate(value) {
  if (!value) return "not set";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "not set";
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export function shortId(value) {
  return String(value || "").slice(0, 8);
}
