import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, type Task } from "../../api";
import { Card } from "../../components/ui/Card";

const STATE_LABEL: Record<Task["state"], string> = {
  BACKLOG: "Backlog",
  TODO: "To do",
  IN_PROGRESS: "In progress",
  DONE: "Done",
};

/**
 * Right-rail widget for the Timeline: tasks due on the viewed date
 * (read-only; managing tasks happens on /tasks and /board).
 * Hides itself while loading, on error, or when nothing is due —
 * a contextual widget must never break the screen around it.
 */
export function TodayTasks({ date }: { date: string }) {
  const [tasks, setTasks] = useState<Task[] | null>(null);

  useEffect(() => {
    let live = true;
    api
      .board()
      .then((b) => {
        if (!live) return;
        setTasks(b.columns.flatMap((c) => c.tasks).filter((t) => t.due_date === date));
      })
      .catch(() => {
        if (live) setTasks([]);
      });
    return () => {
      live = false;
    };
  }, [date]);

  if (tasks === null || tasks.length === 0) return null;

  const done = tasks.filter((t) => t.state === "DONE").length;
  return (
    <Card padding="compact">
      <div className="today-tasks__head">
        <h2 className="today-tasks__title">Today&apos;s tasks</h2>
        <span className="muted" aria-label={`${done} of ${tasks.length} done`}>
          {done}/{tasks.length}
        </span>
      </div>
      <ul className="today-tasks__list">
        {tasks.map((t) => (
          <li key={t.id} className="today-tasks__row">
            <span
              className={`today-tasks__name${t.state === "DONE" ? " today-tasks__name--done" : ""}`}
            >
              {t.title}
            </span>
            <span className="muted">{STATE_LABEL[t.state]}</span>
          </li>
        ))}
      </ul>
      <Link className="today-tasks__link" to="/tasks">
        View all tasks →
      </Link>
    </Card>
  );
}
