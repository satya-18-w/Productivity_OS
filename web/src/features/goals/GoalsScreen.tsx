import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { api, type Goal } from "../../api";
import type { GoalProgress } from "../../components/productivity/StatusBadge";
import { ScreenLayout } from "../../shell/ScreenLayout";
import { PageHeader } from "../../components/layout/PageHeader";
import { Card } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { SegmentedControl } from "../../components/ui/SegmentedControl";
import { StatCard } from "../../components/productivity/StatCard";
import { EmptyState, ErrorState } from "../../components/productivity/states";
import { GOAL_PROGRESS_LABELS } from "../../components/productivity/StatusBadge";
import { GoalRow } from "./GoalRow";
import { GoalDialog, type GoalDialogTarget } from "./GoalDialog";
import { FILTER_OPTIONS, PROGRESS_ORDER, filterGoals, goalStats, type GoalFilter } from "./goalHelpers";

export function GoalsScreen() {
  const [params, setParams] = useSearchParams();
  const filter: GoalFilter =
    (FILTER_OPTIONS.find((f) => f.value === params.get("filter"))?.value as GoalFilter) ?? "all";

  const [goals, setGoals] = useState<Goal[] | null>(null);
  const [error, setError] = useState(false);
  const [dialog, setDialog] = useState<GoalDialogTarget | null>(null);

  const setFilter = useCallback(
    (v: GoalFilter) =>
      setParams(
        (p) => {
          const next = new URLSearchParams(p);
          if (v === "all") next.delete("filter");
          else next.set("filter", v);
          return next;
        },
        { replace: true },
      ),
    [setParams],
  );

  const load = useCallback(async () => {
    setError(false);
    try {
      setGoals(await api.goals());
    } catch {
      setError(true);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function withReload(fn: () => Promise<unknown>) {
    try {
      await fn();
      await load();
    } catch {
      setError(true);
    }
  }
  const setProgress = (goal: Goal, progress: GoalProgress) =>
    withReload(() => api.setGoalProgress(goal.id, progress));
  const del = (goal: Goal) => {
    if (window.confirm(`Delete "${goal.title}"?`)) void withReload(() => api.deleteGoal(goal.id));
  };

  const stats = useMemo(() => (goals ? goalStats(goals) : null), [goals]);
  const shown = useMemo(() => (goals ? filterGoals(goals, filter) : []), [goals, filter]);

  return (
    <ScreenLayout
      railLabel="Goal summary"
      rail={
        stats && stats.total > 0 ? (
          <Card padding="compact" title="By status" headingLevel={2}>
            <ul className="ui-list">
              {PROGRESS_ORDER.map((p) => (
                <li key={p} className="goals-rail__row">
                  <span>{GOAL_PROGRESS_LABELS[p]}</span>
                  <span className="goals-rail__count">{stats.byProgress[p]}</span>
                </li>
              ))}
            </ul>
          </Card>
        ) : undefined
      }
    >
      <PageHeader
        eyebrow="Goals"
        title="Goals"
        subtitle="What you're working toward, and where each one stands."
        actions={<Button onClick={() => setDialog({ mode: "new" })}>New goal</Button>}
      />

      {stats && stats.total > 0 && (
        <div className="goals-kpis">
          <StatCard label="Total" value={stats.total} />
          <StatCard label="In progress" value={stats.byProgress.IN_PROGRESS} tint="info" />
          <StatCard label="Achieved" value={stats.byProgress.ACHIEVED} tint="success" />
          <StatCard label="Not started" value={stats.byProgress.NOT_STARTED} />
        </div>
      )}

      <SegmentedControl label="Filter goals" options={FILTER_OPTIONS} value={filter} onChange={setFilter} />

      {error ? (
        <ErrorState message="Could not load your goals." action={<Button onClick={load}>Retry</Button>} />
      ) : !goals ? (
        <p className="muted">Loading…</p>
      ) : shown.length === 0 ? (
        <EmptyState
          title={filter === "all" ? "No goals yet" : "Nothing here"}
          message={filter === "all" ? "Set a goal and track it manually as you go." : "Try another filter."}
          action={filter === "all" ? <Button onClick={() => setDialog({ mode: "new" })}>New goal</Button> : undefined}
        />
      ) : (
        <div className="goal-list">
          {shown.map((goal) => (
            <GoalRow
              key={goal.id}
              goal={goal}
              onSetProgress={setProgress}
              onEdit={(g) => setDialog({ mode: "edit", goal: g })}
              onDelete={del}
            />
          ))}
        </div>
      )}

      {dialog && (
        <GoalDialog
          open
          target={dialog}
          onClose={() => setDialog(null)}
          onSaved={async () => {
            setDialog(null);
            await load();
          }}
        />
      )}
    </ScreenLayout>
  );
}
