/** Minute-of-day (0–1440) → "HH:MM". */
export function fmtMinute(m: number): string {
  const clamped = Math.max(0, Math.min(1440, m));
  if (clamped === 1440) return "24:00";
  const h = Math.floor(clamped / 60);
  const min = clamped % 60;
  return `${String(h).padStart(2, "0")}:${String(min).padStart(2, "0")}`;
}

/** Seconds → "1h 30m" / "45m" / "−20m". */
export function fmtDuration(seconds: number): string {
  const sign = seconds < 0 ? "−" : "";
  const s = Math.abs(seconds);
  const h = Math.floor(s / 3600);
  const m = Math.round((s % 3600) / 60);
  return h === 0 ? `${sign}${m}m` : `${sign}${h}h ${m}m`;
}

/** Current local time as minutes past midnight. */
export function nowMinutes(): number {
  const d = new Date();
  return d.getHours() * 60 + d.getMinutes();
}
