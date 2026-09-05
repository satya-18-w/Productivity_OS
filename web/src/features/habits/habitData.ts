import { api, type ArchivedHabit, type HabitView } from "../../api";
import { isoWeekRange, shiftDays, toISODate, parseISODate } from "../../components/date/dateUtils";

export interface WeekData {
  weekStart: string; // Monday ISO
  days: string[]; // 7 ISO dates, Mon..Sun
  today: string;
  habits: HabitView[];
  archived: ArchivedHabit[];
  /** completion[habitId][dateISO] === true when marked complete. */
  completion: Record<string, Record<string, boolean>>;
}

/**
 * The ISO week (Monday-first, D8) containing `anyDay`, with per-habit per-day
 * completion. Currently 7 `GET /api/habits?date=` calls — see docs/left.md for the
 * batch endpoint that would replace this.
 */
export async function fetchWeek(anyDay: string): Promise<WeekData> {
  const [weekStart] = isoWeekRange(anyDay);
  const days = Array.from({ length: 7 }, (_, i) => shiftDays(weekStart, i));
  const today = toISODate(new Date());

  const lists = await Promise.all(days.map((d) => api.habits(d)));

  const completion: Record<string, Record<string, boolean>> = {};
  lists.forEach((list, i) => {
    for (const h of list.habits) {
      (completion[h.id] ??= {})[days[i]] = h.completed_on_date;
    }
  });

  // Habit list / streaks / archived: prefer today's response, else the last day's.
  const ref = lists[days.indexOf(today)] ?? lists[lists.length - 1];
  return { weekStart, days, today, habits: ref.habits, archived: ref.archived, completion };
}

/* ------------------------------------------------------------------ mock --- */

/**
 * PLACEHOLDER for `GET /api/habits/history` (see docs/left.md — required for the
 * "This Month" view). Deterministic per habit id so the UI is stable and testable.
 * Density loosely tracks `last_30_days`. Replace with `api.habitHistory(from, to)`.
 */
export function mockHabitHistory(habit: HabitView, days: string[]): Set<string> {
  const density = Math.min(1, Math.max(0.15, habit.last_30_days / 30));
  let seed = 0;
  for (const c of habit.id) seed = (Math.imul(31, seed) + c.charCodeAt(0)) | 0;
  const done = new Set<string>();
  days.forEach((d, i) => {
    seed = (Math.imul(1103515245, seed) + 12345) & 0x7fffffff;
    const r = seed / 0x7fffffff;
    // recent days a touch more likely
    const recencyBoost = i > days.length - 8 ? 0.12 : 0;
    if (r < density + recencyBoost) done.add(d);
  });
  return done;
}

/** N days ending today (inclusive), oldest first. */
export function trailingDays(n: number, endISO = toISODate(new Date())): string[] {
  return Array.from({ length: n }, (_, i) => shiftDays(endISO, -(n - 1 - i)));
}

export function weekdayShort(iso: string): string {
  return parseISODate(iso).toLocaleDateString(undefined, { weekday: "short" });
}
export function dayOfMonth(iso: string): number {
  return parseISODate(iso).getDate();
}
