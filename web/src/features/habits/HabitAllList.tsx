import type { ArchivedHabit, HabitView } from "../../api";
import { Button } from "../../components/ui/Button";
import { EmptyState } from "../../components/productivity/states";
import { ListGroupHeader } from "../../components/productivity/ListRow";
import { Streak, Last30, HabitActions } from "./HabitBits";

export interface HabitAllListProps {
  habits: HabitView[];
  archived: ArchivedHabit[];
  onArchive: (habitId: string) => void;
  onUnarchive: (habitId: string) => void;
  onAdd: () => void;
}

export function HabitAllList({ habits, archived, onArchive, onUnarchive, onAdd }: HabitAllListProps) {
  if (habits.length === 0 && archived.length === 0) {
    return (
      <EmptyState
        title="No habits yet"
        message="Add your first habit to get started."
        action={<Button onClick={onAdd}>Add habit</Button>}
      />
    );
  }

  return (
    <div className="habit-all">
      <section>
        <ListGroupHeader count={habits.length} tone="brand">Active</ListGroupHeader>
        {habits.length === 0 ? (
          <p className="muted">No active habits.</p>
        ) : (
          <ul className="ui-list">
            {habits.map((h) => (
              <li key={h.id} className="habit-all__row">
                <span className="habit-all__name">{h.name}</span>
                <Last30 value={h.last_30_days} />
                <Streak value={h.current_streak} />
                <HabitActions name={h.name} onArchive={() => onArchive(h.id)} />
              </li>
            ))}
          </ul>
        )}
      </section>

      {archived.length > 0 && (
        <section>
          <ListGroupHeader count={archived.length}>Archived</ListGroupHeader>
          <ul className="ui-list">
            {archived.map((a) => (
              <li key={a.id} className="habit-all__row">
                <span className="habit-all__name habit-all__name--muted">{a.name}</span>
                <Button variant="ghost" size="sm" onClick={() => onUnarchive(a.id)}>
                  Unarchive
                </Button>
              </li>
            ))}
          </ul>
          <p className="hint">Archiving keeps a habit's full completion history (Q11).</p>
        </section>
      )}
    </div>
  );
}
