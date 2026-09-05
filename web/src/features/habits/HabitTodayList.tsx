import type { HabitView } from "../../api";
import { ToggleCircle } from "../../components/ui/Toggle";
import { EmptyState } from "../../components/productivity/states";
import { formatFullDate } from "../../components/date/dateUtils";
import { Streak, Last30, HabitActions } from "./HabitBits";

export interface HabitTodayListProps {
  habits: HabitView[];
  date: string;
  /** completion[habitId][date] */
  completion: Record<string, Record<string, boolean>>;
  onToggle: (habitId: string, date: string, done: boolean) => void;
  onArchive: (habitId: string) => void;
  onAdd: () => void;
}

export function HabitTodayList({ habits, date, completion, onToggle, onArchive, onAdd }: HabitTodayListProps) {
  if (habits.length === 0) {
    return (
      <EmptyState
        title="No habits yet"
        message="Add a habit and check it off each day to build a streak."
        action={<button className="ui-btn ui-btn--primary" onClick={onAdd}>Add habit</button>}
      />
    );
  }

  const label = formatFullDate(date);

  return (
    <ul className="ui-list habit-today">
      {habits.map((h) => {
        const done = completion[h.id]?.[date] ?? h.completed_on_date;
        return (
          <li key={h.id} className="habit-today__row">
            <ToggleCircle
              label={`${h.name} — ${label}, ${done ? "completed" : "not completed"}`}
              checked={done}
              onChange={(e) => onToggle(h.id, date, e.target.checked)}
            />
            <span className="habit-today__name">{h.name}</span>
            <Last30 value={h.last_30_days} />
            <Streak value={h.current_streak} />
            <HabitActions name={h.name} onArchive={() => onArchive(h.id)} />
          </li>
        );
      })}
    </ul>
  );
}
