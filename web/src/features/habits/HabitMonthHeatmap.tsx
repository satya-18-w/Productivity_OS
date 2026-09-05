import { useMemo } from "react";
import type { HabitView } from "../../api";
import { EmptyState } from "../../components/productivity/states";
import { Streak, Last30, HabitActions } from "./HabitBits";
import { mockHabitHistory, trailingDays, dayOfMonth } from "./habitData";

export interface HabitMonthHeatmapProps {
  habits: HabitView[];
  today: string;
  onArchive: (habitId: string) => void;
  onAdd: () => void;
}

const DAYS = 35; // 5 weeks

export function HabitMonthHeatmap({ habits, today, onArchive, onAdd }: HabitMonthHeatmapProps) {
  const days = useMemo(() => trailingDays(DAYS, today), [today]);

  if (habits.length === 0) {
    return (
      <EmptyState
        title="No habits yet"
        message="Add a habit to see your monthly consistency."
        action={<button className="ui-btn ui-btn--primary" onClick={onAdd}>Add habit</button>}
      />
    );
  }

  return (
    <div className="habit-month">
      <p className="habit-month__note" role="note">
        ⚠ Sample data — the completion-history endpoint is pending (see docs/left.md).
      </p>
      <div className="habit-grid__scroll">
        <table className="habit-month__table">
          <thead>
            <tr>
              <th scope="col" className="habit-grid__habit-col">Habit</th>
              {days.map((d, i) => (
                <th key={d} scope="col" className="habit-month__dayhead">
                  {i % 7 === 0 || i === days.length - 1 ? dayOfMonth(d) : ""}
                </th>
              ))}
              <th scope="col">Streak</th>
              <th scope="col">Last 30</th>
              <th scope="col"><span className="ui-visually-hidden">Actions</span></th>
            </tr>
          </thead>
          <tbody>
            {habits.map((h) => {
              const done = mockHabitHistory(h, days);
              return (
                <tr key={h.id}>
                  <th scope="row" className="habit-grid__habit-col habit-grid__name">{h.name}</th>
                  {days.map((d) => (
                    <td
                      key={d}
                      className={`habit-month__cell${done.has(d) ? " habit-month__cell--done" : ""}`}
                      title={`${h.name} — ${d}: ${done.has(d) ? "completed" : "not completed"}`}
                    />
                  ))}
                  <td><Streak value={h.current_streak} /></td>
                  <td><Last30 value={h.last_30_days} /></td>
                  <td><HabitActions name={h.name} onArchive={() => onArchive(h.id)} /></td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
