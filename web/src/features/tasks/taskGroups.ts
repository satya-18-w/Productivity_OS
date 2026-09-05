import type { Category, Task, TaskState } from "../../api";
import { isoWeekRange } from "../../components/date/dateUtils";

/** A task's category has no `category_name` from the API (unlike a time block) —
 * resolve it against the account's category list, shared by TaskRow and TaskCard. */
export function categoryNameFor(categoryId: string | null, categories: Category[]): string | null {
  if (!categoryId) return null;
  return categories.find((c) => c.id === categoryId)?.name ?? null;
}

export const STATE_LABELS: Record<TaskState, string> = {
  BACKLOG: "Backlog",
  TODO: "To do",
  IN_PROGRESS: "In progress",
  DONE: "Done",
};
export const STATE_ORDER: TaskState[] = ["BACKLOG", "TODO", "IN_PROGRESS", "DONE"];

export type TaskFilter = "all" | "today" | "upcoming" | "overdue" | "completed";
export type GroupKey = "overdue" | "today" | "upcoming" | "no_date" | "completed";

export function isOverdue(t: Task, today: string): boolean {
  return t.state !== "DONE" && t.due_date != null && t.due_date < today;
}

function groupOf(t: Task, today: string): GroupKey {
  if (t.state === "DONE") return "completed";
  if (t.due_date == null) return "no_date";
  if (t.due_date < today) return "overdue";
  if (t.due_date === today) return "today";
  return "upcoming";
}

const GROUP_META: Record<GroupKey, { label: string; tone: "neutral" | "danger" | "success" | "brand" }> = {
  overdue: { label: "Overdue", tone: "danger" },
  today: { label: "Today", tone: "success" },
  upcoming: { label: "Upcoming", tone: "neutral" },
  no_date: { label: "No due date", tone: "neutral" },
  completed: { label: "Completed", tone: "success" },
};

const GROUP_ORDER: GroupKey[] = ["overdue", "today", "upcoming", "no_date", "completed"];

const FILTER_GROUPS: Record<TaskFilter, GroupKey[]> = {
  all: GROUP_ORDER,
  today: ["today"],
  upcoming: ["upcoming"],
  overdue: ["overdue"],
  completed: ["completed"],
};

function sortWithin(key: GroupKey, tasks: Task[]): Task[] {
  const byDate = (a: Task, b: Task) => (a.due_date ?? "").localeCompare(b.due_date ?? "");
  const byCreatedDesc = (a: Task, b: Task) => b.created_at.localeCompare(a.created_at);
  return [...tasks].sort(key === "no_date" || key === "completed" ? byCreatedDesc : byDate);
}

export interface TaskGroup {
  key: GroupKey;
  label: string;
  tone: "neutral" | "danger" | "success" | "brand";
  tasks: Task[];
}

/** Group tasks for a given filter tab. Empty groups are omitted. */
export function groupTasks(tasks: Task[], today: string, filter: TaskFilter): TaskGroup[] {
  const wanted = new Set(FILTER_GROUPS[filter]);
  const buckets = new Map<GroupKey, Task[]>();
  for (const t of tasks) {
    const g = groupOf(t, today);
    if (!wanted.has(g)) continue;
    if (!buckets.has(g)) buckets.set(g, []);
    buckets.get(g)!.push(t);
  }
  return GROUP_ORDER.filter((k) => buckets.has(k)).map((k) => ({
    key: k,
    label: GROUP_META[k].label,
    tone: GROUP_META[k].tone,
    tasks: sortWithin(k, buckets.get(k)!),
  }));
}

export interface TaskStats {
  total: number;
  completed: number;
  inProgress: number;
  overdue: number;
  dueThisWeek: number;
  byState: Record<TaskState, number>;
}

export function taskStats(tasks: Task[], today: string): TaskStats {
  const [weekStart, weekEnd] = isoWeekRange(today);
  const byState: Record<TaskState, number> = { BACKLOG: 0, TODO: 0, IN_PROGRESS: 0, DONE: 0 };
  let completed = 0;
  let inProgress = 0;
  let overdue = 0;
  let dueThisWeek = 0;
  for (const t of tasks) {
    byState[t.state] += 1;
    if (t.state === "DONE") completed += 1;
    if (t.state === "IN_PROGRESS") inProgress += 1;
    if (isOverdue(t, today)) overdue += 1;
    if (t.state !== "DONE" && t.due_date != null && t.due_date >= weekStart && t.due_date <= weekEnd) {
      dueThisWeek += 1;
    }
  }
  return { total: tasks.length, completed, inProgress, overdue, dueThisWeek, byState };
}
