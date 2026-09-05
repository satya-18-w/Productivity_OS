import { Card } from "../../components/ui/Card";
import { ProgressBar } from "../../components/ui/ProgressBar";
import type { HabitCompletionRow } from "./reportsData";

export function HabitCompletionReport({ rows }: { rows: HabitCompletionRow[] }) {
  return (
    <Card title="Habit completion" headingLevel={2}>
      <p className="secondary report-caption">Completed days and completion rate per habit across the range.</p>
      {rows.length === 0 ? (
        <p className="muted">No habits in this range.</p>
      ) : (
        <ul className="report-habit-list" aria-label="Habit completion">
          {rows.map((r) => {
            const rate = r.rangeDays === 0 ? 0 : Math.round((r.completedDays / r.rangeDays) * 100);
            return (
              <li key={r.habitId} className="report-habit-row">
                <span className="report-habit-row__name">{r.habitName}</span>
                <ProgressBar value={rate} label={`${r.habitName} completion rate`} tone="success" />
                <span className="report-habit-row__figure">
                  {r.completedDays} / {r.rangeDays} days ({rate}%)
                </span>
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}
