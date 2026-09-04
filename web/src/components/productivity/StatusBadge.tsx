import { cx } from "../cx";

/** V1 goal progress states — labels verbatim from requirements §10. */
export type GoalProgress = "not_started" | "in_progress" | "achieved" | "abandoned";

const LABELS: Record<GoalProgress, string> = {
  not_started: "Not started",
  in_progress: "In progress",
  achieved: "Achieved",
  abandoned: "Abandoned",
};

const MODIFIER: Record<GoalProgress, string> = {
  not_started: "not-started",
  in_progress: "in-progress",
  achieved: "achieved",
  abandoned: "abandoned",
};

export interface StatusBadgeProps {
  status: GoalProgress;
  className?: string;
}

/**
 * Goal progress state chip. The four V1 states only — no "On Track" / "At Risk"
 * (those imply a derived health signal V1 does not compute; design-system §4.9).
 * The dot + text mean colour is never the only signal (VP8).
 */
export function StatusBadge({ status, className }: StatusBadgeProps) {
  return (
    <span className={cx("ui-status", `ui-status--${MODIFIER[status]}`, className)}>
      {LABELS[status]}
    </span>
  );
}
