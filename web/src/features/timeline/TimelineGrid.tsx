import { type CSSProperties } from "react";
import type { BlockKind, PositionedBlock } from "../../api";
import { categoryColor } from "../../components/productivity/categoryColor";
import { fmtMinute } from "./timelineFormat";

const HOURS = Array.from({ length: 25 }, (_, i) => i);

function blockLabel(b: PositionedBlock): string {
  const cat = b.category_name ?? "Uncategorized";
  const kind = b.kind === "planned" ? "planned" : "actual";
  const span = `${fmtMinute(b.start_minute)}–${fmtMinute(b.end_minute)}`;
  const extra = b.from_prev_day
    ? ", continues from the previous day"
    : b.to_next_day
      ? ", continues into the next day"
      : "";
  return `${cat} — ${kind}, ${span}${extra}. Edit.`;
}

function Block({ b, onPick }: { b: PositionedBlock; onPick: (b: PositionedBlock) => void }) {
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
      className={`tl2__block${b.kind === "planned" ? " tl2__block--planned" : ""}`}
      style={style}
      onClick={() => onPick(b)}
      aria-label={blockLabel(b)}
    >
      <span className="tl2__block-time">
        {b.from_prev_day ? "▲ " : ""}
        {fmtMinute(b.start_minute)}–{fmtMinute(b.end_minute)}
        {b.to_next_day ? " ▼" : ""}
      </span>
      <span className="tl2__block-cat">{b.category_name ?? "Uncategorized"}</span>
    </button>
  );
}

function Lane({
  kind,
  blocks,
  nowPct,
  onPick,
}: {
  kind: BlockKind;
  blocks: PositionedBlock[];
  nowPct: number | null;
  onPick: (b: PositionedBlock) => void;
}) {
  const label = kind === "planned" ? "Planned" : "Actual";
  return (
    <div className="tl2__lane" role="list" aria-label={`${label} blocks`}>
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
        blocks.map((b) => <Block key={b.id} b={b} onPick={onPick} />)
      )}
    </div>
  );
}

export interface TimelineGridProps {
  planned: PositionedBlock[];
  actual: PositionedBlock[];
  /** Minutes past midnight for the "now" line, or null to hide it. */
  now: number | null;
  onPick: (b: PositionedBlock) => void;
}

export function TimelineGrid({ planned, actual, now, onPick }: TimelineGridProps) {
  const nowPct = now == null ? null : (now / 1440) * 100;
  return (
    <div className="tl2__scroll">
      <div className="tl2__grid">
        <div className="tl2__lane-head tl2__lane-head--axis" aria-hidden="true" />
        <div className="tl2__lane-head">Planned</div>
        <div className="tl2__lane-head">Actual</div>

        <div className="tl2__axis" aria-hidden="true">
          {HOURS.map((h) => (
            <div key={h} className="tl2__tick" style={{ top: `${(h / 24) * 100}%` }}>
              {h % 3 === 0 ? `${String(h).padStart(2, "0")}:00` : ""}
            </div>
          ))}
        </div>
        <Lane kind="planned" blocks={planned} nowPct={nowPct} onPick={onPick} />
        <Lane kind="actual" blocks={actual} nowPct={nowPct} onPick={onPick} />
      </div>
    </div>
  );
}
