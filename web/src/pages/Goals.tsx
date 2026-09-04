import { useCallback, useEffect, useState, type FormEvent } from "react";
import { api, ApiError, type Goal, type GoalProgress, type NewGoal } from "../api";

const PROGRESS: { value: GoalProgress; label: string }[] = [
  { value: "NOT_STARTED", label: "Not started" },
  { value: "IN_PROGRESS", label: "In progress" },
  { value: "ACHIEVED", label: "Achieved" },
  { value: "ABANDONED", label: "Abandoned" },
];
const label = (p: GoalProgress) => PROGRESS.find((x) => x.value === p)?.label ?? p;

type Editing = { mode: "new" } | { mode: "edit"; goal: Goal } | null;

export function Goals() {
  const [goals, setGoals] = useState<Goal[] | null>(null);
  const [error, setError] = useState("");
  const [editing, setEditing] = useState<Editing>(null);

  const load = useCallback(async () => {
    try {
      setGoals(await api.goals());
      setError("");
    } catch {
      setError("Could not load goals.");
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function setProgress(id: string, progress: GoalProgress) {
    try {
      await api.setGoalProgress(id, progress);
      await load();
    } catch {
      setError("Could not update progress.");
    }
  }

  return (
    <div className="stack">
      <section className="card">
        <div className="date-nav">
          <h2>Goals</h2>
          <span className="spacer" />
          <button onClick={() => setEditing({ mode: "new" })}>New goal</button>
        </div>
        {error && <p className="error" role="alert">{error}</p>}
      </section>

      {editing && (
        <GoalForm
          editing={editing}
          onClose={() => setEditing(null)}
          onSaved={async () => {
            setEditing(null);
            await load();
          }}
        />
      )}

      <section className="card">
        {!goals ? (
          <p className="empty">Loading…</p>
        ) : goals.length === 0 ? (
          <p className="empty">No goals yet. Set one to aim at.</p>
        ) : (
          <div>
            {goals.map((g) => (
              <article key={g.id} className="goal-card">
                <div className="goal-head">
                  <button className="goal-title" onClick={() => setEditing({ mode: "edit", goal: g })}>
                    {g.title}
                  </button>
                  <span className={`progress-chip progress-${g.progress}`}>{label(g.progress)}</span>
                </div>
                {g.description && <p className="goal-desc">{g.description}</p>}
                <div className="goal-meta">
                  {g.target_date && <span>🎯 {g.target_date}</span>}
                  <span className="spacer" style={{ flex: 1 }} />
                  <select
                    className="goal-progress-select"
                    value={g.progress}
                    onChange={(e) => setProgress(g.id, e.target.value as GoalProgress)}
                  >
                    {PROGRESS.map((p) => (
                      <option key={p.value} value={p.value}>{p.label}</option>
                    ))}
                  </select>
                </div>
              </article>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

function GoalForm({
  editing,
  onClose,
  onSaved,
}: {
  editing: NonNullable<Editing>;
  onClose: () => void;
  onSaved: () => Promise<void>;
}) {
  const existing = editing.mode === "edit" ? editing.goal : null;

  const [title, setTitle] = useState(existing?.title ?? "");
  const [description, setDescription] = useState(existing?.description ?? "");
  const [targetDate, setTargetDate] = useState(existing?.target_date ?? "");
  const [error, setError] = useState("");
  const [fieldError, setFieldError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setFieldError("");
    setBusy(true);
    const payload: NewGoal = { title, description, target_date: targetDate || null };
    try {
      if (existing) await api.updateGoal(existing.id, payload);
      else await api.createGoal(payload);
      await onSaved();
    } catch (err) {
      if (err instanceof ApiError && err.code === "VALIDATION_ERROR") {
        setFieldError(Object.values(err.fields ?? {})[0] ?? "Check the fields.");
      } else {
        setError("Could not save the goal.");
      }
      setBusy(false);
    }
  }

  async function remove() {
    if (!existing || !window.confirm("Delete this goal?")) return;
    setBusy(true);
    try {
      await api.deleteGoal(existing.id);
      await onSaved();
    } catch {
      setError("Could not delete the goal.");
      setBusy(false);
    }
  }

  return (
    <form className="card" onSubmit={submit}>
      <h2>{existing ? "Edit goal" : "New goal"}</h2>
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
        Target date
        <input type="date" value={targetDate} onChange={(e) => setTargetDate(e.target.value)} />
      </label>
      {fieldError && <span className="field-error">{fieldError}</span>}
      <div className="form-actions">
        <button type="submit" disabled={busy}>{existing ? "Save" : "Create"}</button>
        {existing && (
          <button type="button" className="link danger" onClick={remove} disabled={busy}>Delete</button>
        )}
        <button type="button" className="link" onClick={onClose}>Cancel</button>
      </div>
    </form>
  );
}
