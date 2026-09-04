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
