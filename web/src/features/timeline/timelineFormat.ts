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
  // Round to whole minutes first so e.g. 3599s reads "1h 0m", never "0h 60m".
  const totalMin = Math.round(Math.abs(seconds) / 60);
  const h = Math.floor(totalMin / 60);
  const m = totalMin % 60;
  return h === 0 ? `${sign}${m}m` : `${sign}${h}h ${m}m`;
}

/** Current local time as minutes past midnight. */
export function nowMinutes(): number {
  const d = new Date();
  return d.getHours() * 60 + d.getMinutes();
}
