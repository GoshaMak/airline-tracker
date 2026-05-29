const TIMEZONE_OFFSET_PATTERN = /(?:z|[+-]\d{2}:?\d{2})$/i;

function utcDateValue(value) {
  if (!value) return null;
  const source = typeof value === "string" && !TIMEZONE_OFFSET_PATTERN.test(value) ? `${value}Z` : value;
  const date = new Date(source);
  return Number.isNaN(date.getTime()) ? null : date;
}

function padDatePart(value) {
  return String(value).padStart(2, "0");
}

export function formatDate(value) {
  const date = utcDateValue(value);
  if (!date) return "not set";
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    timeZone: "UTC",
  }).format(date);
}

export function dateTimeInputValue(value) {
  const date = utcDateValue(value);
  if (!date) return "";
  return [
    date.getUTCFullYear(),
    padDatePart(date.getUTCMonth() + 1),
    padDatePart(date.getUTCDate()),
  ].join("-") + `T${padDatePart(date.getUTCHours())}:${padDatePart(date.getUTCMinutes())}`;
}

export function toUtcISOString(value) {
  const date = utcDateValue(value);
  return date ? date.toISOString() : null;
}

export function dateTimeLocalToUtcISOString(value) {
  if (!value) return null;
  const normalized = value.length === 16 ? `${value}:00` : value;
  const date = new Date(`${normalized}Z`);
  if (Number.isNaN(date.getTime())) return null;
  return date.toISOString();
}

export function shortId(value) {
  return String(value || "").slice(0, 8);
}
