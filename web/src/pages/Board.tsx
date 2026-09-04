import { useCallback, useEffect, useState, type DragEvent, type FormEvent } from "react";
import { api, ApiError, type Board as BoardData, type NewTask, type Task, type TaskState } from "../api";

const LABELS: Record<TaskState, string> = {
  BACKLOG: "Backlog",
  TODO: "To do",
  IN_PROGRESS: "In progress",
  DONE: "Done",
};
const ORDER: TaskState[] = ["BACKLOG", "TODO", "IN_PROGRESS", "DONE"];

type Editing = { mode: "new" } | { mode: "edit"; task: Task } | null;

export function Board() {
  const [board, setBoard] = useState<BoardData | null>(null);
  const [error, setError] = useState("");
  const [editing, setEditing] = useState<Editing>(null);
  const [dragOver, setDragOver] = useState<TaskState | null>(null);

  const load = useCallback(async () => {
    try {
      setBoard(await api.board());
      setError("");
    } catch {
      setError("Could not load the board.");
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function move(id: string, to: TaskState) {
    try {
      await api.moveTask(id, to);
      await load();
    } catch {
      setError("Could not move the task.");
    }
  }

  function onDrop(e: DragEvent, to: TaskState) {
    e.preventDefault();
    setDragOver(null);
    const id = e.dataTransfer.getData("text/task-id");
    if (id) void move(id, to);
  }

  return (
    <div className="stack">
      <section className="card">
        <div className="date-nav">
          <h2>Board</h2>
          <span className="spacer" />
          <button onClick={() => setEditing({ mode: "new" })}>Add task</button>
        </div>
        {error && <p className="error" role="alert">{error}</p>}
      </section>

      {editing && (
        <TaskForm
          editing={editing}
          onClose={() => setEditing(null)}
          onSaved={async () => {
            setEditing(null);
            await load();
          }}
        />
      )}

      <div className="board-scroll">
        <div className="board">
          {ORDER.map((state) => {
            const col = board?.columns.find((c) => c.state === state);
            const tasks = col?.tasks ?? [];
            return (
              <div
                key={state}
                className={`board-col${dragOver === state ? " drag-over" : ""}`}
                onDragOver={(e) => {
                  e.preventDefault();
                  setDragOver(state);
                }}
                onDragLeave={() => setDragOver((s) => (s === state ? null : s))}
                onDrop={(e) => onDrop(e, state)}
              >
                <div className="board-col-head">
                  <span>{LABELS[state]}</span>
                  <span className="count">{tasks.length}</span>
                </div>
                {tasks.map((task) => (
                  <TaskCard
                    key={task.id}
                    task={task}
                    onEdit={() => setEditing({ mode: "edit", task })}
                    onMove={move}
                  />
                ))}
                {!board && <p className="muted">Loading…</p>}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

function TaskCard({
  task,
  onEdit,
  onMove,
}: {
  task: Task;
  onEdit: () => void;
  onMove: (id: string, to: TaskState) => void;
}) {
  return (
    <article
      className="task-card"
      draggable
      onDragStart={(e) => e.dataTransfer.setData("text/task-id", task.id)}
    >
      <button className="task-title link" onClick={onEdit} title="Edit task">
        {task.title}
      </button>
      {task.due_date && <span className="task-due">Due {task.due_date}</span>}
      {task.description && <p className="task-desc">{task.description}</p>}
      <label className="task-move">
        Move to
        <select value={task.state} onChange={(e) => onMove(task.id, e.target.value as TaskState)}>
          {ORDER.map((s) => (
            <option key={s} value={s}>{LABELS[s]}</option>
          ))}
        </select>
      </label>
    </article>
  );
}

function TaskForm({
  editing,
  onClose,
  onSaved,
}: {
  editing: NonNullable<Editing>;
  onClose: () => void;
  onSaved: () => Promise<void>;
}) {
  const existing = editing.mode === "edit" ? editing.task : null;

  const [title, setTitle] = useState(existing?.title ?? "");
  const [description, setDescription] = useState(existing?.description ?? "");
  const [dueDate, setDueDate] = useState(existing?.due_date ?? "");
  const [error, setError] = useState("");
  const [fieldError, setFieldError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setFieldError("");
    setBusy(true);
    const payload: NewTask = { title, description, due_date: dueDate || null };
    try {
      if (existing) await api.updateTask(existing.id, payload);
      else await api.createTask(payload);
      await onSaved();
    } catch (err) {
      if (err instanceof ApiError && err.code === "VALIDATION_ERROR") {
        setFieldError(Object.values(err.fields ?? {})[0] ?? "Check the fields.");
      } else {
        setError("Could not save the task.");
      }
      setBusy(false);
    }
  }

  async function remove() {
    if (!existing || !window.confirm("Delete this task?")) return;
    setBusy(true);
    try {
      await api.deleteTask(existing.id);
      await onSaved();
    } catch {
      setError("Could not delete the task.");
      setBusy(false);
    }
  }

  return (
    <form className="card" onSubmit={submit}>
      <h2>{existing ? "Edit task" : "Add task"}</h2>
      {error && <p className="error" role="alert">{error}</p>}
      <label>
        Title
        <input type="text" required maxLength={200} value={title} onChange={(e) => setTitle(e.target.value)} autoFocus />
      </label>
      <label>
        Description
        <textarea rows={3} maxLength={5000} value={description} onChange={(e) => setDescription(e.target.value)} />
      </label>
      <label>
        Due date
        <input type="date" value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
      </label>
      {fieldError && <span className="field-error">{fieldError}</span>}
      <div className="form-actions">
        <button type="submit" disabled={busy}>{existing ? "Save" : "Add"}</button>
        {existing && (
          <button type="button" className="link danger" onClick={remove} disabled={busy}>Delete</button>
        )}
        <button type="button" className="link" onClick={onClose}>Cancel</button>
      </div>
    </form>
  );
}
