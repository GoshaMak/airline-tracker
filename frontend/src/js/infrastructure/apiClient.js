import { API_BASE, REQUEST_TIMEOUT_MS } from "../config.js";

function readErrorMessage(body) {
  if (!body) return "";
  if (typeof body === "string") return body;
  if (typeof body.msg === "string") return body.msg;
  if (typeof body.err === "string") return body.err;
  return "";
}

export function createApiClient(getToken) {
  return {
    async request(path, options = {}) {
      const token = getToken();
      const controller = new AbortController();
      const timeoutId = window.setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);
      const headers = {
        Accept: "application/json",
        ...(options.body ? { "Content-Type": "application/json" } : {}),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.headers,
      };

      try {
        let response;
        try {
          response = await fetch(`${API_BASE}${path}`, { ...options, headers, signal: controller.signal });
        } catch (error) {
          if (error?.name === "AbortError") {
            error.status = 0;
            error.isTimeout = true;
            throw error;
          }
          error.status = 0;
          throw error;
        }

        const contentType = response.headers.get("content-type") || "";
        let body = "";
        try {
          body = contentType.includes("application/json") ? await response.json() : await response.text();
        } catch {
          body = "";
        }

        if (!response.ok) {
          const error = new Error(readErrorMessage(body) || `Request failed with ${response.status}`);
          error.status = response.status;
          error.body = body;
          throw error;
        }

        return body;
      } finally {
        window.clearTimeout(timeoutId);
      }
    },
  };
}
