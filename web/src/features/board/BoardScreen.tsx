import { useCallback, useEffect, useMemo, useState } from "react";
import { api, type Board, type Category, type Task, type TaskState } from "../../api";
import { ScreenLayout } from "../../shell/ScreenLayout";
import { PageHeader } from "../../components/layout/PageHeader";
import { Button } from "../../components/ui/Button";
import { ErrorState } from "../../components/productivity/states";
import { STATE_ORDER } from "../tasks/taskGroups";
import { TaskDialog, type TaskDialogTarget } from "../tasks/TaskDialog";
import { BoardColumn } from "./BoardColumn";

export function BoardScreen() {
  const [board, setBoard] = useState<Board | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [error, setError] = useState(false);
  const [dialog, setDialog] = useState<TaskDialogTarget | null>(null);

  const load = useCallback(async () => {
    setError(false);
    try {
      const [b, cats] = await Promise.all([api.board(), api.listCategories()]);
      setBoard(b);
      setCategories(cats);
    } catch {
      setError(true);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const byState = useMemo(() => {
    const m = new Map<TaskState, Task[]>();
    for (const s of STATE_ORDER) m.set(s, []);
    for (const col of board?.columns ?? []) m.set(col.state, col.tasks);
    return m;
  }, [board]);

  async function moveTo(id: string, to: TaskState) {
    // No-op if the task is already in that column.
    const current = board?.columns.find((c) => c.tasks.some((t) => t.id === id))?.state;
    if (current === to) return;
    try {
      await api.moveTask(id, to);
      await load();
    } catch {
      setError(true);
    }
  }
  const move = (task: Task, state: TaskState) => moveTo(task.id, state);
  const del = (task: Task) => {
    if (!window.confirm(`Delete "${task.title}"?`)) return;
    void (async () => {
      try {
        await api.deleteTask(task.id);
        await load();
      } catch {
        setError(true);
      }
    })();
  };

  return (
    <ScreenLayout>
      <PageHeader
        eyebrow="Board"
        title="Board"
        subtitle="Your tasks in four fixed columns. Drag a card, or use its menu, to change status."
        actions={<Button onClick={() => setDialog({ mode: "new" })}>Add task</Button>}
      />

      {error && (
        <ErrorState message="Something went wrong with the board." action={<Button onClick={load}>Retry</Button>} />
      )}

      {!board ? (
        <p className="muted">Loading…</p>
      ) : (
        <div className="board2__scroll">
          <div className="board2">
            {STATE_ORDER.map((state) => (
              <BoardColumn
                key={state}
                state={state}
                tasks={byState.get(state) ?? []}
                categories={categories}
                onDropTask={moveTo}
                onEdit={(t) => setDialog({ mode: "edit", task: t })}
                onMove={move}
                onDelete={del}
              />
            ))}
          </div>
        </div>
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
