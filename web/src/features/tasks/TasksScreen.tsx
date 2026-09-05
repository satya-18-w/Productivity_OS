import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { api, type Category, type Task, type TaskState } from "../../api";
import { ScreenLayout } from "../../shell/ScreenLayout";
import { PageHeader } from "../../components/layout/PageHeader";
import { Card } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { SegmentedControl } from "../../components/ui/SegmentedControl";
import { StatCard } from "../../components/productivity/StatCard";
import { ListGroupHeader } from "../../components/productivity/ListRow";
import { EmptyState, ErrorState } from "../../components/productivity/states";
import { todayISO } from "../../components/date/dateUtils";
import { TaskRow } from "./TaskRow";
import { TaskDialog, type TaskDialogTarget } from "./TaskDialog";
import { groupTasks, taskStats, STATE_LABELS, STATE_ORDER, type TaskFilter } from "./taskGroups";

const FILTERS = [
  { value: "all", label: "All" },
  { value: "today", label: "Today" },
  { value: "upcoming", label: "Upcoming" },
  { value: "overdue", label: "Overdue" },
  { value: "completed", label: "Completed" },
] as const;

export function TasksScreen() {
  const [params, setParams] = useSearchParams();
  const filter: TaskFilter =
    (FILTERS.find((f) => f.value === params.get("filter"))?.value as TaskFilter) ?? "all";

  const [tasks, setTasks] = useState<Task[] | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [error, setError] = useState(false);
  const [dialog, setDialog] = useState<TaskDialogTarget | null>(null);
  // Last non-DONE state per task, so unchecking the "done" box restores where
  // the task came from (spec: DONE ⇄ previous state). Unknown → TODO.
  const prevState = useRef(new Map<string, TaskState>());

  const setFilter = useCallback(
    (v: TaskFilter) =>
      setParams(
        (prev) => {
          const next = new URLSearchParams(prev);
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
      const [board, cats] = await Promise.all([api.board(), api.listCategories()]);
      setTasks(board.columns.flatMap((c) => c.tasks));
      setCategories(cats);
    } catch {
      setError(true);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  // A Timeline block linked to a task deep-links here as ?openTask=<id> (the
  // "clear path to Task" requirement) — open its edit dialog once, then drop
  // the param so it doesn't reopen on a later re-render or back-navigation.
  useEffect(() => {
    const openId = params.get("openTask");
    if (!openId || !tasks) return;
    const target = tasks.find((t) => t.id === openId);
    if (target) setDialog({ mode: "edit", task: target });
    setParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        next.delete("openTask");
        return next;
      },
      { replace: true },
    );
  }, [params, tasks, setParams]);

  const today = todayISO();
  const stats = useMemo(() => (tasks ? taskStats(tasks, today) : null), [tasks, today]);
  const groups = useMemo(
    () => (tasks ? groupTasks(tasks, today, filter) : []),
    [tasks, today, filter],
  );

  async function withReload(fn: () => Promise<unknown>) {
    try {
      await fn();
      await load();
    } catch {
      setError(true);
    }
  }
  const toggleDone = (task: Task, done: boolean) =>
    withReload(async () => {
      if (done) {
        if (task.state !== "DONE") prevState.current.set(task.id, task.state);
        await api.moveTask(task.id, "DONE");
      } else {
        await api.moveTask(task.id, prevState.current.get(task.id) ?? "TODO");
        prevState.current.delete(task.id);
      }
    });
  const move = (task: Task, state: TaskState) => withReload(() => api.moveTask(task.id, state));
  const del = (task: Task) => {
    if (window.confirm(`Delete "${task.title}"?`)) void withReload(() => api.deleteTask(task.id));
  };

  return (
    <ScreenLayout
      railLabel="Task summary"
      rail={
        stats && (
          <Card padding="compact" title="By status" headingLevel={2}>
            <ul className="ui-list">
              {STATE_ORDER.map((s) => (
                <li key={s} className="tasks-rail__row">
                  <span>{STATE_LABELS[s]}</span>
                  <span className="tasks-rail__count">{stats.byState[s]}</span>
                </li>
              ))}
            </ul>
          </Card>
        )
      }
    >
      <PageHeader
        eyebrow="Tasks"
        title="Tasks"
        subtitle="Everything on your plate, grouped by when it's due."
        actions={<Button onClick={() => setDialog({ mode: "new" })}>Add task</Button>}
      />

      {stats && (
        <div className="tasks-kpis">
          <StatCard label="Total" value={stats.total} sublabel={`${stats.completed} completed`} tint="none" />
          <StatCard label="In progress" value={stats.inProgress} tint="info" />
          <StatCard label="Overdue" value={stats.overdue} tint={stats.overdue ? "warning" : "none"} />
          <StatCard label="Due this week" value={stats.dueThisWeek} tint="none" />
        </div>
      )}

      <SegmentedControl
        label="Filter tasks"
        options={FILTERS}
        value={filter}
        onChange={setFilter}
      />

      {error ? (
        <ErrorState message="Could not load your tasks." action={<Button onClick={load}>Retry</Button>} />
      ) : !tasks ? (
        <p className="muted">Loading…</p>
      ) : groups.length === 0 ? (
        <EmptyState
          title={filter === "all" ? "No tasks yet" : "Nothing here"}
          message={filter === "all" ? "Add a task to get started." : "Try another filter."}
          action={filter === "all" ? <Button onClick={() => setDialog({ mode: "new" })}>Add task</Button> : undefined}
        />
      ) : (
        groups.map((g) => (
          <section key={g.key} className="tasks-group">
            <ListGroupHeader count={g.tasks.length} tone={g.tone}>
              {g.label}
            </ListGroupHeader>
            <ul className="ui-list">
              {g.tasks.map((task) => (
                <TaskRow
                  key={task.id}
                  task={task}
                  today={today}
                  categories={categories}
                  onToggleDone={toggleDone}
                  onMove={move}
                  onEdit={(t) => setDialog({ mode: "edit", task: t })}
                  onDelete={del}
                />
              ))}
            </ul>
          </section>
        ))
      )}

      {dialog && (
        <TaskDialog
          open
          target={dialog}
          categories={categories}
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
