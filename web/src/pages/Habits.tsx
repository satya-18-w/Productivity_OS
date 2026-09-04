import { useCallback, useEffect, useState, type FormEvent } from "react";
import { api, type HabitList, type HabitView } from "../api";

function today(): string {
  return new Date().toLocaleDateString("en-CA");
}
function shiftDate(date: string, days: number): string {
  const d = new Date(date + "T00:00:00");
  d.setDate(d.getDate() + days);
  return d.toLocaleDateString("en-CA");
}

export function Habits() {
  const [date, setDate] = useState(today());
  const [data, setData] = useState<HabitList | null>(null);
  const [error, setError] = useState("");
  const [showArchived, setShowArchived] = useState(false);

  const load = useCallback(async () => {
    try {
      setData(await api.habits(date));
      setError("");
    } catch {
      setError("Could not load habits.");
    }
  }, [date]);

  useEffect(() => {
    void load();
  }, [load]);

  async function toggle(h: HabitView) {
    try {
      if (h.completed_on_date) await api.unmarkHabit(h.id, date);
      else await api.markHabit(h.id, date);
      await load();
    } catch {
      setError("Could not update the habit.");
    }
  }

  return (
    <div className="stack">
      <section className="card">
        <div className="date-nav">
          <button className="link" onClick={() => setDate(shiftDate(date, -1))}>‹ Prev</button>
          <input type="date" value={date} onChange={(e) => setDate(e.target.value || today())} />
          <button className="link" onClick={() => setDate(shiftDate(date, 1))}>Next ›</button>
          <button className="link" onClick={() => setDate(today())}>Today</button>
        </div>
        {error && <p className="error" role="alert">{error}</p>}
        <NewHabitForm onCreated={load} />
      </section>

      <section className="card">
        <h2>Habits</h2>
        {!data ? (
          <p className="muted">Loading…</p>
        ) : data.habits.length === 0 ? (
          <p className="muted">No active habits yet.</p>
        ) : (
          <ul className="rows">
            {data.habits.map((h) => (
              <li key={h.id} className="row habit-row">
                <button
                  className={`habit-check${h.completed_on_date ? " done" : ""}`}
                  onClick={() => toggle(h)}
                  aria-pressed={h.completed_on_date}
                  title={h.completed_on_date ? "Mark not done" : "Mark done"}
                >
                  {h.completed_on_date ? "✓" : ""}
                </button>
                <span className="row-name">{h.name}</span>
                <span className="habit-streak" title="Current streak">🔥 {h.current_streak}</span>
                <span className="muted habit-30">{h.last_30_days}/30d</span>
                <button className="link danger" onClick={() => archive(h.id, load, setError)}>Archive</button>
              </li>
            ))}
          </ul>
        )}
      </section>

      {data && data.archived.length > 0 && (
        <section className="card">
          <button className="link" onClick={() => setShowArchived((v) => !v)}>
            {showArchived ? "Hide" : "Show"} archived ({data.archived.length})
          </button>
          {showArchived && (
            <ul className="rows">
              {data.archived.map((a) => (
                <li key={a.id} className="row">
                  <span className="row-name">{a.name}</span>
                  <button
                    className="link"
                    onClick={async () => {
                      await api.unarchiveHabit(a.id);
                      await load();
                    }}
                  >
                    Unarchive
                  </button>
                </li>
              ))}
            </ul>
          )}
        </section>
      )}
    </div>
  );
}

async function archive(id: string, reload: () => Promise<void>, setError: (s: string) => void) {
  if (!window.confirm("Archive this habit? Its completion history is kept.")) return;
  try {
    await api.archiveHabit(id);
    await reload();
  } catch {
    setError("Could not archive the habit.");
  }
}

function NewHabitForm({ onCreated }: { onCreated: () => Promise<void> }) {
  const [name, setName] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      await api.createHabit(name);
      setName("");
      await onCreated();
    } catch {
      setError("Could not create the habit.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form className="inline-form" onSubmit={submit}>
      <input type="text" placeholder="New habit name" required maxLength={100}
        value={name} onChange={(e) => setName(e.target.value)} />
      <button type="submit" disabled={busy}>Add</button>
      {error && <span className="field-error">{error}</span>}
    </form>
  );
}
