/**
 * PLACEHOLDER data layer for Daily Review (`v1.md §11`, Q1).
 * No reviews backend exists yet — see docs/left.md ("Phase 10 — Daily Review") for the
 * endpoints this stands in for. Swap `fetchDailyReview` / `saveDailyReview` for real
 * `api.dailyReview(date)` / `api.saveDailyReview(date, answers)` calls in
 * `DailyReviewScreen.tsx` when they land. The in-memory store below only persists for
 * the current page session (it resets on reload) — enough to demo create/edit/view.
 */

/** Fixed, non-editable prompt set (Q1 — placeholder wording, replace freely). */
export const DAILY_REVIEW_PROMPTS = [
  { key: "wentWell", label: "What went well today?" },
  { key: "notPlanned", label: "What didn't go as planned?" },
  { key: "differently", label: "What will you do differently tomorrow?" },
  { key: "grateful", label: "One thing you're grateful for." },
] as const;

export type DailyReviewPromptKey = (typeof DAILY_REVIEW_PROMPTS)[number]["key"];

export type DailyReviewAnswers = Record<DailyReviewPromptKey, string>;

export interface DailyReview {
  date: string;
  answers: DailyReviewAnswers;
  updated_at: string;
}

export function emptyAnswers(): DailyReviewAnswers {
  return Object.fromEntries(DAILY_REVIEW_PROMPTS.map((p) => [p.key, ""])) as DailyReviewAnswers;
}

const store = new Map<string, DailyReview>();

/** Look up a saved review for `date`, or null if none exists yet. */
export async function fetchDailyReview(date: string): Promise<DailyReview | null> {
  return store.get(date) ?? null;
}

/** Create or overwrite the review for `date`. */
export async function saveDailyReview(date: string, answers: DailyReviewAnswers): Promise<DailyReview> {
  const record: DailyReview = { date, answers, updated_at: new Date().toISOString() };
  store.set(date, record);
  return record;
}

/* ------------------------------------------------------------------ Weekly */

const WEEK_STORE_KEY = "weekly:";

/** Fixed, non-editable prompt set (Q2 — placeholder wording, replace freely). */
export const WEEKLY_REVIEW_PROMPTS = [
  { key: "highlights", label: "What were the highlights of this week?" },
  { key: "struggles", label: "What did you struggle with?" },
  { key: "timeIntended", label: "Did your time go where you intended?" },
  { key: "nextPriority", label: "What is the one priority for next week?" },
] as const;

export type WeeklyReviewPromptKey = (typeof WEEKLY_REVIEW_PROMPTS)[number]["key"];

export type WeeklyReviewAnswers = Record<WeeklyReviewPromptKey, string>;

export interface WeeklyReview {
  /** ISO Monday of the reviewed week. */
  weekStart: string;
  answers: WeeklyReviewAnswers;
  updated_at: string;
}

export function emptyWeeklyAnswers(): WeeklyReviewAnswers {
  return Object.fromEntries(WEEKLY_REVIEW_PROMPTS.map((p) => [p.key, ""])) as WeeklyReviewAnswers;
}

const weeklyStore = new Map<string, WeeklyReview>();

/** Look up a saved review for the week starting `weekStart`, or null. */
export async function fetchWeeklyReview(weekStart: string): Promise<WeeklyReview | null> {
  return weeklyStore.get(WEEK_STORE_KEY + weekStart) ?? null;
}

/** Create or overwrite the review for the week starting `weekStart`. */
export async function saveWeeklyReview(weekStart: string, answers: WeeklyReviewAnswers): Promise<WeeklyReview> {
  const record: WeeklyReview = { weekStart, answers, updated_at: new Date().toISOString() };
  weeklyStore.set(WEEK_STORE_KEY + weekStart, record);
  return record;
}

/**
 * PLACEHOLDER weekly reference totals (`v1.md §12`): that ISO week's actual
 * time per category, habit completion counts, and tasks-entered-DONE count.
 * No weekly range backend exists yet — deterministic mock keyed by weekStart so
 * the same week always shows the same figures. Swap for
 * `api.comparisonRange` / `api.habits/range` / `api.tasks/throughput` calls
 * when they land (see docs/left.md).
 */
export interface WeekReference {
  weekStart: string;
  weekEnd: string;
  categorySeconds: { name: string; seconds: number }[];
  habitCounts: { name: string; count: number }[];
  tasksDone: number;
}

function hashStr(s: string): number {
  let h = 0;
  for (let i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) | 0;
  return Math.abs(h);
}

export async function mockWeekReference(weekStart: string, weekEnd: string): Promise<WeekReference> {
  const h = hashStr(weekStart);
  const cats = ["Deep Work", "Study", "Personal"];
  const habits = ["Morning workout", "Read 20 pages", "Meditate"];
  return {
    weekStart,
    weekEnd,
    categorySeconds: cats.map((name, i) => ({
      name,
      seconds: 3600 * (2 + ((h >> (i * 3)) & 7)),
    })),
    habitCounts: habits.map((name, i) => ({ name, count: (h >> (i * 2)) % 8 })),
    tasksDone: h % 12,
  };
}
