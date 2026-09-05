import { ToggleCircle } from "../../components/ui/Toggle";
import { EmptyState } from "../../components/productivity/states";
import { Streak, Last30, HabitActions } from "./HabitBits";
import { weekdayShort, dayOfMonth, type WeekData } from "./habitData";

export interface HabitWeekGridProps {
  week: WeekData;
  onToggle: (habitId: string, date: string, done: boolean) => void;
  onArchive: (habitId: string) => void;
  onAdd: () => void;
}

export function HabitWeekGrid({ week, onToggle, onArchive, onAdd }: HabitWeekGridProps) {
  if (week.habits.length === 0) {
    return (
      <EmptyState
        title="No habits yet"
        message="Add a habit to start tracking your week."
        action={<button className="ui-btn ui-btn--primary" onClick={onAdd}>Add habit</button>}
      />
    );
  }

  return (
    <div className="habit-grid__scroll">
      <table className="habit-grid">
        <thead>
          <tr>
            <th scope="col" className="habit-grid__habit-col">Habit</th>
            {week.days.map((d) => (
              <th
                key={d}
                scope="col"
                className={`habit-grid__day${d === week.today ? " habit-grid__day--today" : ""}`}
              >
                <span className="habit-grid__wd">{weekdayShort(d)}</span>
                <span className="habit-grid__dm">{dayOfMonth(d)}</span>
              </th>
            ))}
            <th scope="col">Streak</th>
            <th scope="col">Last 30</th>
            <th scope="col"><span className="ui-visually-hidden">Actions</span></th>
          </tr>
        </thead>
        <tbody>
          {week.habits.map((h) => (
            <tr key={h.id}>
              <th scope="row" className="habit-grid__habit-col habit-grid__name">{h.name}</th>
              {week.days.map((d) => {
                const done = week.completion[h.id]?.[d] ?? false;
                return (
                  <td key={d} className="habit-grid__cell">
                    <ToggleCircle
                      label={`${h.name} — ${weekdayShort(d)} ${dayOfMonth(d)}, ${done ? "completed" : "not completed"}`}
                      checked={done}
                      onChange={(e) => onToggle(h.id, d, e.target.checked)}
                    />
                  </td>
                );
              })}
              <td className="habit-grid__cell"><Streak value={h.current_streak} /></td>
              <td className="habit-grid__cell"><Last30 value={h.last_30_days} /></td>
              <td className="habit-grid__cell"><HabitActions name={h.name} onArchive={() => onArchive(h.id)} /></td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
