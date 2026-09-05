import { cx } from "../cx";

/**
 * V1 goal progress states — the API values (`requirements` §10). Labels are
 * verbatim from the requirement: Not started / In progress / Achieved / Abandoned.
 * NOT the reference's "On Track" / "At Risk" — those imply a derived health
 * signal V1 does not compute (design-system.md §4.9).
 */
export type GoalProgress = "NOT_STARTED" | "IN_PROGRESS" | "ACHIEVED" | "ABANDONED";

export const GOAL_PROGRESS_LABELS: Record<GoalProgress, string> = {
  NOT_STARTED: "Not started",
  IN_PROGRESS: "In progress",
  ACHIEVED: "Achieved",
  ABANDONED: "Abandoned",
};

const MODIFIER: Record<GoalProgress, string> = {
  NOT_STARTED: "not-started",
  IN_PROGRESS: "in-progress",
  ACHIEVED: "achieved",
  ABANDONED: "abandoned",
};

export interface StatusBadgeProps {
  status: GoalProgress;
  className?: string;
}

/** Goal progress state chip. Dot + text — colour is never the only signal (VP8). */
export function StatusBadge({ status, className }: StatusBadgeProps) {
  return (
    <span className={cx("ui-status", `ui-status--${MODIFIER[status]}`, className)}>
      {GOAL_PROGRESS_LABELS[status]}
    </span>
  );
}
