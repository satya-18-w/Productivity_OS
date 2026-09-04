import { cx } from "../cx";

export interface ProgressBarProps {
  /** 0–100. Clamped. */
  value: number;
  /** Accessible name, e.g. "Time logged". */
  label: string;
  tone?: "brand" | "success" | "warning" | "danger" | "goal";
  className?: string;
}

/**
 * Determinate progress / ratio bar (design-system.md §4.11).
 * NOTE: not for goal progress — V1 goals have four manual states, no % (§10).
 */
export function ProgressBar({ value, label, tone = "brand", className }: ProgressBarProps) {
  const pct = Math.max(0, Math.min(100, value));
  return (
    <span
      className={cx("ui-progress", className)}
      role="progressbar"
      aria-label={label}
      aria-valuenow={Math.round(pct)}
      aria-valuemin={0}
      aria-valuemax={100}
    >
      <span
        className={cx("ui-progress__fill", tone !== "brand" && `ui-progress__fill--${tone}`)}
        style={{ width: `${pct}%` }}
      />
    </span>
  );
}
