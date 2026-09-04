// Typed client for the Productivity OS API. Same-origin JSON; the session cookie
// rides along automatically. State-changing requests carry the double-submit CSRF
// token read from the csrf_token cookie (ADR-0004).

export interface Account {
  email: string;
  timezone: string;
}

export interface FieldErrors {
  [field: string]: string;
}

export class ApiError extends Error {
  code: string;
  status: number;
  fields?: FieldErrors;

  constructor(status: number, code: string, message: string, fields?: FieldErrors) {
    super(message);
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

function csrfToken(): string {
  const m = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]+)/);
  return m ? decodeURIComponent(m[1]) : "";
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {};
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (method !== "GET" && method !== "HEAD") headers["X-CSRF-Token"] = csrfToken();

  const res = await fetch(path, {
    method,
    headers,
    credentials: "same-origin",
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  if (res.status === 204) return undefined as T;

  const text = await res.text();
  const data = text ? JSON.parse(text) : {};

  if (!res.ok) {
    const e = data.error ?? {};
    throw new ApiError(res.status, e.code ?? "UNKNOWN", e.message ?? "Request failed", e.fields);
  }
  return data as T;
}

export const api = {
  getAccount: () => request<Account>("GET", "/api/account"),
  register: (email: string, password: string, timezone: string) =>
    request<Account>("POST", "/api/accounts", { email, password, timezone }),
  login: (email: string, password: string) =>
    request<void>("POST", "/api/sessions", { email, password }),
  logout: () => request<void>("DELETE", "/api/sessions/current"),
  setTimezone: (timezone: string) =>
    request<void>("PUT", "/api/account/timezone", { timezone }),
  changePassword: (current_password: string, new_password: string) =>
    request<void>("PUT", "/api/account/password", { current_password, new_password }),
};

export function browserTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
  } catch {
    return "UTC";
  }
}
