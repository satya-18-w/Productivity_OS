// Typed client for the Productivity OS API. Same-origin JSON; the session cookie
// rides along automatically. State-changing requests carry the double-submit CSRF
// token read from the csrf_token cookie (ADR-0004).

export interface Account {
  email: string;
  timezone: string;
}

export interface Category {
  id: string;
  name: string;
}

export type BlockKind = "planned" | "actual";

export interface Block {
  id: string;
  kind: BlockKind;
  starts_at: string;
  ends_at: string;
  category_id: string | null;
  category_name?: string | null;
}

export interface PositionedBlock extends Block {
  start_minute: number;
  end_minute: number;
  from_prev_day: boolean;
  to_next_day: boolean;
  local_date: string;
  local_start: string;
  local_end: string;
  ends_next_day: boolean;
}

export interface DayTimeline {
  date: string;
  planned: PositionedBlock[];
  actual: PositionedBlock[];
}

export interface ComparisonRow {
  category_id: string | null;
  category_name: string;
  planned_seconds: number;
  actual_seconds: number;
  difference_seconds: number;
}

export interface DayComparison {
  date: string;
  categories: ComparisonRow[];
}

export interface NewBlock {
  kind: BlockKind;
  date: string;
  start: string;
  end: string;
  ends_next_day: boolean;
  category_id: string | null;
}

export type TaskState = "BACKLOG" | "TODO" | "IN_PROGRESS" | "DONE";

export interface Task {
  id: string;
  title: string;
  description: string;
  due_date: string | null;
  state: TaskState;
  created_at: string;
  updated_at: string;
}

export interface BoardColumn {
  state: TaskState;
  tasks: Task[];
}

export interface Board {
  columns: BoardColumn[];
}

export interface NewTask {
  title: string;
  description: string;
  due_date: string | null;
}

export interface HabitView {
  id: string;
  name: string;
  current_streak: number;
  completed_on_date: boolean;
  last_30_days: number;
}

export interface ArchivedHabit {
  id: string;
  name: string;
}

export interface HabitList {
  date: string;
  habits: HabitView[];
  archived: ArchivedHabit[];
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

  listCategories: () =>
    request<{ categories: Category[] }>("GET", "/api/categories").then((r) => r.categories),
  createCategory: (name: string) => request<Category>("POST", "/api/categories", { name }),
  renameCategory: (id: string, name: string) =>
    request<void>("PATCH", `/api/categories/${id}`, { name }),
  archiveCategory: (id: string) => request<void>("POST", `/api/categories/${id}/archive`),

  timeline: (date: string) =>
    request<DayTimeline>("GET", `/api/timeline?date=${encodeURIComponent(date)}`),
  comparison: (date: string) =>
    request<DayComparison>("GET", `/api/comparison?date=${encodeURIComponent(date)}`),
  createBlock: (b: NewBlock) => request<Block>("POST", "/api/blocks", b),
  updateBlock: (id: string, b: NewBlock) => request<void>("PUT", `/api/blocks/${id}`, b),
  deleteBlock: (id: string) => request<void>("DELETE", `/api/blocks/${id}`),

  board: () => request<Board>("GET", "/api/board"),
  createTask: (t: NewTask) => request<Task>("POST", "/api/tasks", t),
  updateTask: (id: string, t: NewTask) => request<void>("PATCH", `/api/tasks/${id}`, t),
  moveTask: (id: string, state: TaskState) =>
    request<void>("PUT", `/api/tasks/${id}/state`, { state }),
  deleteTask: (id: string) => request<void>("DELETE", `/api/tasks/${id}`),

  habits: (date?: string) =>
    request<HabitList>("GET", date ? `/api/habits?date=${encodeURIComponent(date)}` : "/api/habits"),
  createHabit: (name: string) =>
    request<ArchivedHabit>("POST", "/api/habits", { name }),
  archiveHabit: (id: string) => request<void>("POST", `/api/habits/${id}/archive`),
  unarchiveHabit: (id: string) => request<void>("POST", `/api/habits/${id}/unarchive`),
  markHabit: (id: string, date: string) =>
    request<void>("PUT", `/api/habits/${id}/completions/${date}`),
  unmarkHabit: (id: string, date: string) =>
    request<void>("DELETE", `/api/habits/${id}/completions/${date}`),
};

export function browserTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
  } catch {
    return "UTC";
  }
}
