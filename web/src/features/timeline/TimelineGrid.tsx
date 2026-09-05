import { useEffect, useRef, type CSSProperties } from "react";
import type { BlockKind, PositionedBlock } from "../../api";
import { categoryColor } from "../../components/productivity/categoryColor";
import { fmtMinute } from "./timelineFormat";

/** Default landing hour when the date has no blocks and isn't today (06:00). */
const DEFAULT_ANCHOR_MINUTE = 6 * 60;

const HOURS = Array.from({ length: 25 }, (_, i) => i);

/** A task-linked block shows the task's title (the association); otherwise its category. */
function blockText(b: PositionedBlock, taskTitleById: Map<string, string>): string {
  if (b.task_id) return taskTitleById.get(b.task_id) ?? "Linked task";
  return b.category_name ?? "Uncategorized";
}

function blockLabel(b: PositionedBlock, taskTitleById: Map<string, string>): string {
  const text = blockText(b, taskTitleById);
  const kind = b.kind === "planned" ? "planned" : "actual";
  const span = `${fmtMinute(b.start_minute)}–${fmtMinute(b.end_minute)}`;
  const extra = b.from_prev_day
    ? ", continues from the previous day"
    : b.to_next_day
      ? ", continues into the next day"
      : "";
  const linked = b.task_id ? ", linked to a task" : "";
  return `${text} — ${kind}, ${span}${extra}${linked}. Edit.`;
}

function Block({
  b,
  taskTitleById,
  onPick,
}: {
  b: PositionedBlock;
  taskTitleById: Map<string, string>;
  onPick: (b: PositionedBlock) => void;
}) {
  const top = (b.start_minute / 1440) * 100;
  const height = Math.max(2.4, ((b.end_minute - b.start_minute) / 1440) * 100);
  const style = {
    top: `${top}%`,
    height: `${height}%`,
    "--tl-cat": categoryColor(b.category_id),
  } as CSSProperties;

  return (
    <button
      type="button"
      className={`tl2__block${b.kind === "planned" ? " tl2__block--planned" : ""}${b.task_id ? " tl2__block--linked" : ""}`}
      style={style}
      onClick={() => onPick(b)}
      aria-label={blockLabel(b, taskTitleById)}
    >
      <span className="tl2__block-line">
        <span className="tl2__block-time">
          {b.from_prev_day ? "▲ " : ""}
          {fmtMinute(b.start_minute)}–{fmtMinute(b.end_minute)}
          {b.to_next_day ? " ▼" : ""}
        </span>
        <span className="tl2__block-cat">
          {b.task_id ? "↳ " : ""}
          {blockText(b, taskTitleById)}
        </span>
      </span>
    </button>
  );
}

function Lane({
  kind,
  blocks,
  nowPct,
  taskTitleById,
  onPick,
}: {
  kind: BlockKind;
  blocks: PositionedBlock[];
  nowPct: number | null;
  taskTitleById: Map<string, string>;
  onPick: (b: PositionedBlock) => void;
}) {
  const label = kind === "planned" ? "Planned" : "Actual";
  return (
    <div className={`tl2__lane tl2__lane--${kind}`} role="list" aria-label={`${label} blocks`}>
      {HOURS.map((h) => (
        <div key={h} className="tl2__hourline" style={{ top: `${(h / 24) * 100}%` }} />
      ))}
      {nowPct != null && (
        <div
          className="tl2__now"
          style={{ top: `${nowPct}%` }}
          data-time={fmtMinute(Math.round((nowPct / 100) * 1440))}
        />
      )}
      {blocks.length === 0 ? (
        <p className="tl2__empty">Nothing {kind}</p>
      ) : (
        blocks.map((b) => <Block key={b.id} b={b} taskTitleById={taskTitleById} onPick={onPick} />)
      )}
    </div>
  );
}

export interface TimelineGridProps {
  planned: PositionedBlock[];
  actual: PositionedBlock[];
  /** Minutes past midnight for the "now" line, or null to hide it. */
  now: number | null;
  /** Task id → title, for labelling task-linked blocks (empty map if none). */
  taskTitleById?: Map<string, string>;
  onPick: (b: PositionedBlock) => void;
}

export function TimelineGrid({ planned, actual, now, taskTitleById = new Map(), onPick }: TimelineGridProps) {
  const nowPct = now == null ? null : (now / 1440) * 100;
  // Anchored to the axis, not the outer grid — the axis spans exactly the 24h
  // track (no sticky header row), so a percentage of its height maps directly
  // to a minute-of-day.
  const axisRef = useRef<HTMLDivElement>(null);

  // Land on the relevant part of the day, not a fixed empty midnight — the
  // grid spans the full 24h (G1) but nobody wants to scroll past six empty
  // hours to find their first block.
  useEffect(() => {
    const axis = axisRef.current;
    if (!axis) return;
    const starts = [...planned, ...actual].map((b) => b.start_minute);
    const anchor = now ?? (starts.length ? Math.min(...starts) : DEFAULT_ANCHOR_MINUTE);
    const leadInMinutes = 60; // a little context above the anchor
    const targetPct = Math.max(0, anchor - leadInMinutes) / 1440;
    const y = axis.getBoundingClientRect().top + window.scrollY + targetPct * axis.offsetHeight;
    window.scrollTo({ top: Math.max(0, y) });
  }, [planned, actual, now]);

  return (
    <div className="tl2__scroll">
      <div className="tl2__grid">
        <div className="tl2__lane-head tl2__lane-head--axis" aria-hidden="true" />
        <div className="tl2__lane-head">Planned</div>
        <div className="tl2__lane-head">Actual</div>

        <div className="tl2__axis" aria-hidden="true" ref={axisRef}>
          {HOURS.map((h) => (
            <div key={h} className="tl2__tick" style={{ top: `${(h / 24) * 100}%` }}>
              {h % 3 === 0 ? `${String(h).padStart(2, "0")}:00` : ""}
            </div>
          ))}
        </div>
        <Lane kind="planned" blocks={planned} nowPct={nowPct} taskTitleById={taskTitleById} onPick={onPick} />
        <Lane kind="actual" blocks={actual} nowPct={nowPct} taskTitleById={taskTitleById} onPick={onPick} />
      </div>
    </div>
  );
}
