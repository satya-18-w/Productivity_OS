import { shiftDays, todayISO } from "../../components/date/dateUtils";

/**
 * PLACEHOLDER data layer for the five fixed V1 reports (`requirements` §13).
 * No reports backend exists yet — see docs/left.md ("Phase 9 — Reports") for the
 * endpoint this is standing in for. Swap `mockReportData` for a real
 * `api.reports(from, to)` call in `ReportsScreen.tsx` when it lands.
 */

export interface CategoryTime {
  categoryId: string | null; // null = the explicit "Uncategorized" bucket (Q8)
  categoryName: string;
  seconds: number;
}

export interface PlannedVsActualRow {
  categoryId: string | null;
  categoryName: string;
  plannedSeconds: number;
  actualSeconds: number;
  differenceSeconds: number;
}

export interface HabitCompletionRow {
  habitId: string;
  habitName: string;
  completedDays: number;
  rangeDays: number;
}

export interface DailyTotal {
  date: string;
  seconds: number;
}

export interface ReportData {
  from: string;
  to: string;
  timeByCategory: CategoryTime[];
  plannedVsActual: PlannedVsActualRow[];
  habitCompletion: HabitCompletionRow[];
  taskThroughput: number;
  dailyActualTotals: DailyTotal[];
}

const SAMPLE_CATEGORIES = [
  { id: "c1", name: "Deep Work" },
  { id: "c2", name: "Admin" },
  { id: "c3", name: "Exercise" },
  { id: "c4", name: "Learning" },
  { id: null, name: "Uncategorized" },
];

const SAMPLE_HABITS = [
  { id: "h1", name: "Workout" },
  { id: "h2", name: "Read" },
  { id: "h3", name: "Meditation" },
];

function seededRandom(seed: string): () => number {
  let s = 0;
  for (const c of seed) s = (Math.imul(31, s) + c.charCodeAt(0)) | 0;
  return () => {
    s = (Math.imul(1103515245, s) + 12345) & 0x7fffffff;
    return s / 0x7fffffff;
  };
}

function dateRange(from: string, to: string): string[] {
  const days: string[] = [];
  let d = from;
  while (d <= to) {
    days.push(d);
    d = shiftDays(d, 1);
  }
  return days;
}

/** Deterministic sample data for [from, to] (inclusive). Same range → same figures. */
export function mockReportData(from: string, to: string): ReportData {
  const rand = seededRandom(`${from}:${to}`);
  const days = dateRange(from, to);

  const dailyActualTotals: DailyTotal[] = days.map((date) => ({
    date,
    seconds: Math.round((1.5 + rand() * 6) * 3600),
  }));

  const timeByCategory: CategoryTime[] = SAMPLE_CATEGORIES.map((c) => ({
    categoryId: c.id,
    categoryName: c.name,
    seconds: Math.round((c.id ? 4 + rand() * 20 : rand() * 4) * 3600),
  })).sort((a, b) => b.seconds - a.seconds);

  const plannedVsActual: PlannedVsActualRow[] = SAMPLE_CATEGORIES.filter((c) => c.id).map((c) => {
    const planned = Math.round((5 + rand() * 15) * 3600);
    const actual = Math.round(planned * (0.6 + rand() * 0.7));
    return {
      categoryId: c.id,
      categoryName: c.name,
      plannedSeconds: planned,
      actualSeconds: actual,
      differenceSeconds: actual - planned,
    };
  });

  const habitCompletion: HabitCompletionRow[] = SAMPLE_HABITS.map((h) => ({
    habitId: h.id,
    habitName: h.name,
    completedDays: Math.round(days.length * (0.4 + rand() * 0.55)),
    rangeDays: days.length,
  }));

  const taskThroughput = Math.round(3 + rand() * 15);

  return { from, to, timeByCategory, plannedVsActual, habitCompletion, taskThroughput, dailyActualTotals };
}

/** Default range: the trailing 30 days ending today (v1.md states no default). */
export function defaultRange(): { from: string; to: string } {
  const to = todayISO();
  return { from: shiftDays(to, -29), to };
}
