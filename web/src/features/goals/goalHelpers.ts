import type { Goal } from "../../api";
import type { GoalProgress } from "../../components/productivity/StatusBadge";

export const PROGRESS_ORDER: GoalProgress[] = ["NOT_STARTED", "IN_PROGRESS", "ACHIEVED", "ABANDONED"];

export type GoalFilter = "all" | GoalProgress;

export const FILTER_OPTIONS: { value: GoalFilter; label: string }[] = [
  { value: "all", label: "All" },
  { value: "NOT_STARTED", label: "Not started" },
  { value: "IN_PROGRESS", label: "In progress" },
  { value: "ACHIEVED", label: "Achieved" },
  { value: "ABANDONED", label: "Abandoned" },
];

/** Newest first (matches the other list screens). */
function byCreatedDesc(a: Goal, b: Goal): number {
  return b.created_at.localeCompare(a.created_at);
}

export function filterGoals(goals: Goal[], filter: GoalFilter): Goal[] {
  const list = filter === "all" ? goals : goals.filter((g) => g.progress === filter);
  return [...list].sort(byCreatedDesc);
}

export interface GoalStats {
  total: number;
  byProgress: Record<GoalProgress, number>;
}

export function goalStats(goals: Goal[]): GoalStats {
  const byProgress: Record<GoalProgress, number> = {
    NOT_STARTED: 0,
    IN_PROGRESS: 0,
    ACHIEVED: 0,
    ABANDONED: 0,
  };
  for (const g of goals) byProgress[g.progress] += 1;
  return { total: goals.length, byProgress };
}
