import { useCallback, useEffect, useMemo, useState, type FormEvent } from "react";
import {
  api,
  ApiError,
  type BlockKind,
  type Category,
  type DayComparison,
  type DayTimeline,
  type NewBlock,
  type PositionedBlock,
} from "../api";

function today(): string {
  return new Date().toLocaleDateString("en-CA"); // YYYY-MM-DD
}

function shiftDate(date: string, days: number): string {
  const d = new Date(date + "T00:00:00");
  d.setDate(d.getDate() + days);
  return d.toLocaleDateString("en-CA");
}

function fmtMinute(m: number): string {
  const clamped = Math.max(0, Math.min(1440, m));
  if (clamped === 1440) return "24:00";
  const h = Math.floor(clamped / 60);
  const min = clamped % 60;
  return `${String(h).padStart(2, "0")}:${String(min).padStart(2, "0")}`;
}

function fmtDuration(seconds: number): string {
  const sign = seconds < 0 ? "−" : "";
  const s = Math.abs(seconds);
  const h = Math.floor(s / 3600);
  const m = Math.round((s % 3600) / 60);
  return h === 0 ? `${sign}${m}m` : `${sign}${h}h ${m}m`;
}

type Editing = { mode: "new" } | { mode: "edit"; block: PositionedBlock } | null;

export function Timeline() {
  const [date, setDate] = useState(today());
  const [data, setData] = useState<DayTimeline | null>(null);
  const [comparison, setComparison] = useState<DayComparison | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [error, setError] = useState("");
  const [editing, setEditing] = useState<Editing>(null);

  const load = useCallback(async () => {
    try {
      const [tl, cmp, cats] = await Promise.all([
        api.timeline(date),
        api.comparison(date),
        api.listCategories(),
      ]);
      setData(tl);
      setComparison(cmp);
      setCategories(cats);
      setError("");
    } catch {
      setError("Could not load the timeline.");
    }
  }, [date]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <div className="stack">
      <section className="card">
        <div className="date-nav">
          <button className="link" onClick={() => setDate(shiftDate(date, -1))}>‹ Prev</button>
          <input type="date" value={date} onChange={(e) => setDate(e.target.value || today())} />
          <button className="link" onClick={() => setDate(shiftDate(date, 1))}>Next ›</button>
          <button className="link" onClick={() => setDate(today())}>Today</button>
          <span className="spacer" />
          <button onClick={() => setEditing({ mode: "new" })}>Add block</button>
        </div>
        {error && <p className="error" role="alert">{error}</p>}
      </section>

      {editing && (
        <BlockForm
          date={date}
          categories={categories}
          editing={editing}
          onClose={() => setEditing(null)}
          onSaved={async () => {
            setEditing(null);
            await load();
          }}
        />
      )}

      <section className="card">
        <TimelineGrid
          data={data}
          onPick={(block) => setEditing({ mode: "edit", block })}
        />
      </section>

      <section className="card">
        <h2>Planned vs actual</h2>
        <ComparisonTable comparison={comparison} />
      </section>
    </div>
  );
}

function ComparisonTable({ comparison }: { comparison: DayComparison | null }) {
  if (!comparison) return <p className="muted">Loading…</p>;
  if (comparison.categories.length === 0) return <p className="muted">No time logged or planned for this date.</p>;

  const totals = comparison.categories.reduce(
    (acc, c) => ({
      planned: acc.planned + c.planned_seconds,
      actual: acc.actual + c.actual_seconds,
      diff: acc.diff + c.difference_seconds,
    }),
    { planned: 0, actual: 0, diff: 0 },
  );

  return (
    <div className="tl-scroll">
      <table className="totals">
        <thead>
          <tr>
            <th>Category</th>
            <th>Planned</th>
            <th>Actual</th>
            <th>Difference</th>
          </tr>
        </thead>
        <tbody>
          {comparison.categories.map((c) => (
            <tr key={c.category_id ?? "uncategorized"}>
              <td>{c.category_name}</td>
              <td>{fmtDuration(c.planned_seconds)}</td>
              <td>{fmtDuration(c.actual_seconds)}</td>
              <td className={c.difference_seconds < 0 ? "neg" : c.difference_seconds > 0 ? "pos" : ""}>
                {fmtDuration(c.difference_seconds)}
              </td>
            </tr>
          ))}
        </tbody>
        <tfoot>
          <tr>
            <td>Total</td>
            <td>{fmtDuration(totals.planned)}</td>
            <td>{fmtDuration(totals.actual)}</td>
            <td>{fmtDuration(totals.diff)}</td>
          </tr>
        </tfoot>
      </table>
    </div>
  );
}

function TimelineGrid({ data, onPick }: { data: DayTimeline | null; onPick: (b: PositionedBlock) => void }) {
  const hours = useMemo(() => Array.from({ length: 25 }, (_, i) => i), []);
  if (!data) return <p className="muted">Loading…</p>;

  const lane = (blocks: PositionedBlock[], kind: BlockKind) => (
    <div className="tl-lane">
      <div className="tl-lane-head">{kind === "planned" ? "Planned" : "Actual"}</div>
      <div className="tl-track">
        {hours.map((h) => (
          <div key={h} className="tl-hourline" style={{ top: `${(h / 24) * 100}%` }} />
        ))}
        {blocks.length === 0 && <p className="tl-empty muted">Nothing {kind}</p>}
        {blocks.map((b) => {
          const top = (b.start_minute / 1440) * 100;
          const height = Math.max(2.2, ((b.end_minute - b.start_minute) / 1440) * 100);
          return (
            <button
              key={b.id}
              className={`tl-block tl-${kind}`}
              style={{ top: `${top}%`, height: `${height}%` }}
              onClick={() => onPick(b)}
              title="Edit block"
            >
              <span className="tl-block-time">
                {b.from_prev_day ? "▲ " : ""}
                {fmtMinute(b.start_minute)}–{fmtMinute(b.end_minute)}
                {b.to_next_day ? " ▼" : ""}
              </span>
              <span className="tl-block-cat">{b.category_name ?? "Uncategorized"}</span>
            </button>
          );
        })}
      </div>
    </div>
  );

  return (
    <div className="tl-scroll">
      <div className="tl">
        <div className="tl-axis">
          {hours.map((h) => (
            <div key={h} className="tl-tick" style={{ top: `${(h / 24) * 100}%` }}>
              {h % 3 === 0 ? `${String(h).padStart(2, "0")}:00` : ""}
            </div>
          ))}
        </div>
        {lane(data.planned, "planned")}
        {lane(data.actual, "actual")}
      </div>
    </div>
  );
}

function BlockForm({
  date,
  categories,
  editing,
  onClose,
  onSaved,
}: {
  date: string;
  categories: Category[];
  editing: NonNullable<Editing>;
  onClose: () => void;
  onSaved: () => Promise<void>;
}) {
  const existing = editing.mode === "edit" ? editing.block : null;

  const [kind, setKind] = useState<BlockKind>(existing?.kind ?? "planned");
  const [formDate, setFormDate] = useState(existing?.local_date ?? date);
  const [start, setStart] = useState(existing?.local_start ?? "09:00");
  const [end, setEnd] = useState(existing?.local_end ?? "10:00");
  const [endsNextDay, setEndsNextDay] = useState(existing?.ends_next_day ?? false);
  const [categoryID, setCategoryID] = useState(existing?.category_id ?? "");
  const [error, setError] = useState("");
  const [fieldError, setFieldError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setFieldError("");
    setBusy(true);
    const payload: NewBlock = {
      kind,
      date: formDate,
      start,
      end,
      ends_next_day: endsNextDay,
      category_id: categoryID || null,
    };
    try {
      if (existing) await api.updateBlock(existing.id, payload);
      else await api.createBlock(payload);
      await onSaved();
    } catch (err) {
      if (err instanceof ApiError && err.code === "VALIDATION_ERROR") {
        setFieldError(Object.values(err.fields ?? {})[0] ?? "Check the fields.");
      } else {
        setError("Could not save the block.");
      }
      setBusy(false);
    }
  }

  async function remove() {
    if (!existing || !window.confirm("Delete this block?")) return;
    setBusy(true);
    try {
      await api.deleteBlock(existing.id);
      await onSaved();
    } catch {
      setError("Could not delete the block.");
      setBusy(false);
    }
  }

  return (
    <form className="card" onSubmit={submit}>
      <h2>{existing ? "Edit block" : "Add block"}</h2>
      {error && <p className="error" role="alert">{error}</p>}

      <fieldset className="radio-row">
        <label>
          <input type="radio" name="kind" checked={kind === "planned"} disabled={!!existing}
            onChange={() => setKind("planned")} /> Planned
        </label>
        <label>
          <input type="radio" name="kind" checked={kind === "actual"} disabled={!!existing}
            onChange={() => setKind("actual")} /> Actual
        </label>
      </fieldset>

      <label>Date
        <input type="date" required value={formDate} onChange={(e) => setFormDate(e.target.value)} />
      </label>
      <div className="two-col">
        <label>Start
          <input type="time" required value={start} onChange={(e) => setStart(e.target.value)} />
        </label>
        <label>End
          <input type="time" required value={end} onChange={(e) => setEnd(e.target.value)} />
        </label>
      </div>
      <label className="checkbox-row">
        <input type="checkbox" checked={endsNextDay} onChange={(e) => setEndsNextDay(e.target.checked)} />
        Ends on the next day
      </label>
      <label>Category
        <select value={categoryID} onChange={(e) => setCategoryID(e.target.value)}>
          <option value="">— none —</option>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>{c.name}</option>
          ))}
        </select>
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
