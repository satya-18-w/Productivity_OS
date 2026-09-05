import { useMemo, useState, type CSSProperties } from "react";
import { Link } from "react-router-dom";
import type { PositionedBlock } from "../../api";
import { Chip } from "../../components/ui/Chip";
import { Badge } from "../../components/ui/Badge";
import { EmptyState } from "../../components/productivity/states";
import { categoryColor } from "../../components/productivity/categoryColor";
import { fmtMinute } from "./timelineFormat";

const UNCATEGORIZED = "__uncat__";

function catKey(b: PositionedBlock): string {
  return b.category_id ?? UNCATEGORIZED;
}
function catName(b: PositionedBlock): string {
  return b.category_name ?? "Uncategorized";
}
/** A task-linked block's row label is the task's title, not its (inherited) category. */
function rowText(b: PositionedBlock, taskTitleById: Map<string, string>): string {
  if (b.task_id) return taskTitleById.get(b.task_id) ?? "Linked task";
  return catName(b);
}

export interface AgendaListProps {
  planned: PositionedBlock[];
  actual: PositionedBlock[];
  /** Minutes past midnight, or null — used to mark rows as past. */
  now: number | null;
  /** Task id → title, for labelling task-linked blocks (empty map if none). */
  taskTitleById?: Map<string, string>;
  onPick: (b: PositionedBlock) => void;
  /** Opens the create-block surface (foot "Add an agenda item" row). */
  onAdd?: () => void;
}

export function AgendaList({ planned, actual, now, taskTitleById = new Map(), onPick, onAdd }: AgendaListProps) {
  const [filter, setFilter] = useState<string | null>(null); // null = all

  const items = useMemo(
    () => [...planned, ...actual].sort((a, b) => a.start_minute - b.start_minute || a.end_minute - b.end_minute),
    [planned, actual],
  );

  const facets = useMemo(() => {
    const m = new Map<string, { name: string; count: number }>();
    for (const b of items) {
      const k = catKey(b);
      const e = m.get(k) ?? { name: catName(b), count: 0 };
      e.count += 1;
      m.set(k, e);
    }
    return [...m.entries()];
  }, [items]);

  const shown = filter ? items.filter((b) => catKey(b) === filter) : items;

  if (items.length === 0) {
    return <EmptyState title="Nothing scheduled" message="Add a planned or actual block for this date." />;
  }

  return (
    <div className="agenda">
      <div className="agenda__filters" role="group" aria-label="Filter by category">
        <Chip active={filter === null} onToggle={() => setFilter(null)}>
          All ({items.length})
        </Chip>
        {facets.map(([key, { name, count }]) => (
          <Chip
            key={key}
            active={filter === key}
            onToggle={() => setFilter(filter === key ? null : key)}
            dotColor={categoryColor(key === UNCATEGORIZED ? null : key)}
          >
            {name} ({count})
          </Chip>
        ))}
      </div>

      {shown.length === 0 ? (
        <p className="muted">No blocks in this category for this date.</p>
      ) : (
        <ol className="agenda__list">
          {shown.map((b) => {
            const past = now != null && b.end_minute <= now;
            const catColor = categoryColor(b.category_id);
            const linkedLabel = b.task_id
              ? `, linked to ${taskTitleById.get(b.task_id) ?? "a task"}`
              : "";
            return (
              <li key={`${b.kind}-${b.id}`}>
                <button
                  type="button"
                  className={`agenda__row${past ? " agenda__row--past" : ""}${b.task_id ? " agenda__row--linked" : ""}`}
                  style={{ "--agenda-cat": catColor } as CSSProperties}
                  onClick={() => onPick(b)}
                  aria-label={`${rowText(b, taskTitleById)} — ${b.kind}, ${fmtMinute(b.start_minute)}–${fmtMinute(b.end_minute)}${linkedLabel}. Edit.`}
                >
                  <span className="agenda__time">
                    {b.from_prev_day ? "▲ " : ""}
                    {fmtMinute(b.start_minute)}
                    <span className="agenda__time-sep">–</span>
                    {fmtMinute(b.end_minute)}
                    {b.to_next_day ? " ▼" : ""}
                  </span>
                  <span className="agenda__rail" aria-hidden="true">
                    <span className="agenda__dot" style={{ background: catColor }} />
                  </span>
                  <span className="agenda__cat">
                    {b.task_id ? "↳ " : ""}
                    {rowText(b, taskTitleById)}
                  </span>
                  <Badge tone={b.kind === "actual" ? "brand" : "neutral"}>
                    {b.kind === "actual" ? "Actual" : "Planned"}
                  </Badge>
                </button>
                {b.task_id && (
                  <Link
                    className="agenda__open-task"
                    style={{ "--agenda-cat": catColor } as CSSProperties}
                    to={`/tasks?openTask=${b.task_id}`}
                  >
                    Open task →
                  </Link>
                )}
              </li>
            );
          })}
        </ol>
      )}
      {onAdd && (
        <button type="button" className="agenda__add" onClick={onAdd}>
          <span aria-hidden="true">+</span> Add an agenda item
        </button>
      )}
    </div>
  );
}
