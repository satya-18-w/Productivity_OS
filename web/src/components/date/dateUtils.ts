/** Local-date helpers. All dates are `YYYY-MM-DD` strings in the viewer's zone. */

export function toISODate(d: Date): string {
  return d.toLocaleDateString("en-CA"); // YYYY-MM-DD, local
}

export function todayISO(): string {
  return toISODate(new Date());
}

export function parseISODate(iso: string): Date {
  return new Date(iso + "T00:00:00");
}

export function shiftDays(iso: string, days: number): string {
  const d = parseISODate(iso);
  d.setDate(d.getDate() + days);
  return toISODate(d);
}

export function shiftMonths(iso: string, months: number): string {
  const d = parseISODate(iso);
  d.setMonth(d.getMonth() + months, 1);
  return toISODate(d);
}

/** 0 = Monday … 6 = Sunday (D8 — ISO / Monday-first). */
export function isoWeekday(d: Date): number {
  return (d.getDay() + 6) % 7;
}

export const WEEKDAYS_MON_FIRST = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"] as const;

/**
 * The 6×7 grid of dates for the month containing `iso`, starting on the Monday
 * on or before the 1st.
 */
export function monthGrid(iso: string): Date[] {
  const anchor = parseISODate(iso);
  const first = new Date(anchor.getFullYear(), anchor.getMonth(), 1);
  const start = new Date(first);
  start.setDate(first.getDate() - isoWeekday(first));
  return Array.from({ length: 42 }, (_, i) => {
    const d = new Date(start);
    d.setDate(start.getDate() + i);
    return d;
  });
}

export function formatMonthLabel(iso: string): string {
  return parseISODate(iso).toLocaleDateString(undefined, { month: "long", year: "numeric" });
}

export function formatFullDate(iso: string): string {
  return parseISODate(iso).toLocaleDateString(undefined, {
    weekday: "long",
    day: "numeric",
    month: "long",
    year: "numeric",
  });
}

export function formatShortDate(iso: string): string {
  return parseISODate(iso).toLocaleDateString(undefined, { day: "numeric", month: "short" });
}

/** [Monday, Sunday] ISO-date strings for the ISO week containing `iso` (D8). */
export function isoWeekRange(iso: string): [string, string] {
  const monday = shiftDays(iso, -isoWeekday(parseISODate(iso)));
  return [monday, shiftDays(monday, 6)];
}

/**
 * Formats a UTC instant (`starts_at`/`ends_at` from the API) as an ISO date +
 * 24h "HH:MM" in a *given* IANA zone — unlike the rest of this file, not the
 * viewer's browser zone. Display-only: every write path already sends
 * wall-clock values the server resolves itself (ADR-0005 — "the client never
 * does tz math"); this exists only because `GET /api/tasks/{id}/blocks`
 * returns raw UTC instants rather than the pre-resolved `local_date`/
 * `local_start`/`local_end` every other block-reading endpoint provides — see
 * `docs/left.md`. Pass the account's own `timezone` (`useAuth().account`),
 * not the browser's.
 */
export function utcInZone(isoUtc: string, timeZone: string): { date: string; time: string } {
  const d = new Date(isoUtc);
  const date = new Intl.DateTimeFormat("en-CA", { timeZone, year: "numeric", month: "2-digit", day: "2-digit" }).format(d);
  const time = new Intl.DateTimeFormat("en-GB", { timeZone, hour: "2-digit", minute: "2-digit", hourCycle: "h23" }).format(d);
  return { date, time };
}
