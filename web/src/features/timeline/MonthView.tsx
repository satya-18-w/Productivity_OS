import { useEffect, useRef, useState, type CSSProperties } from "react";
import { api, type DayTimeline, type PositionedBlock } from "../../api";
import { ErrorState } from "../../components/productivity/states";
import { Button } from "../../components/ui/Button";
import { categoryColor } from "../../components/productivity/categoryColor";
import { cx } from "../../components/cx";
import { monthGrid, toISODate, parseISODate, todayISO, WEEKDAYS_MON_FIRST } from "../../components/date/dateUtils";
import { fmtMinute } from "./timelineFormat";

const MAX_CHIPS = 3;

function catName(b: PositionedBlock): string {
  return b.category_name ?? "Uncategorized";
}

/**
 * Timeline — Month (G2, `design-system.md §6.1`; `screens/timeline-month.md`).
 * A calendar-month grid, up to 3 block chips per day + "+N more". Blocks
 * only — this is not the excluded, separate Calendar/"event" screen
 * (`screens/timeline-month.md` "Overlap with calendar.md"). Fetches the
 * whole visible grid (up to 42 days) in one `api.timelineRange` call (real
 * data) — see `docs/left.md`, since resolved.
 */
export function MonthView({
  date,
  onPick,
  onJumpToDay,
}: {
  /** Any date within the target month. */
  date: string;
  onPick: (b: PositionedBlock) => void;
  onJumpToDay: (iso: string) => void;
}) {
  const monthKey = date.slice(0, 7); // "YYYY-MM" — only the month identity should refetch
  const grid = monthGrid(date);
  const currentMonth = parseISODate(date).getMonth();
  const [byDate, setByDate] = useState<Map<string, DayTimeline> | null>(null);
  const [error, setError] = useState(false);
  const todayCellRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let live = true;
    setByDate(null);
    setError(false);
    const days = monthGrid(`${monthKey}-01`).map(toISODate);
    api
      .timelineRange(days[0], days[days.length - 1])
      .then((range) => {
        if (!live) return;
        const m = new Map<string, DayTimeline>();
        for (const d of range.days) m.set(d.date, d);
        setByDate(m);
      })
      .catch(() => {
        if (live) setError(true);
      });
    return () => {
      live = false;
    };
  }, [monthKey]);

  // Same reasoning as Week's horizontal auto-scroll: bring today's cell into
  // view rather than defaulting to the 1st of the month.
  useEffect(() => {
    if (!byDate) return;
    todayCellRef.current?.scrollIntoView({ inline: "center", block: "nearest" });
  }, [byDate]);

  if (error) {
    return (
      <ErrorState
        message="Could not load this month."
        action={<Button onClick={() => setError(false)}>Retry</Button>}
      />
    );
  }
  if (!byDate) return <p className="muted">Loading…</p>;

  const today = todayISO();

  return (
    <div className="tl-month__scroll">
      <div className="tl-month">
        <div className="tl-month__weekdays" aria-hidden="true">
          {WEEKDAYS_MON_FIRST.map((d) => (
            <span key={d}>{d}</span>
          ))}
        </div>
        <div className="tl-month__grid">
          {grid.map((d) => {
            const iso = toISODate(d);
          const outside = d.getMonth() !== currentMonth;
          const isToday = iso === today;
          const tl = byDate.get(iso);
          const items = tl ? [...tl.planned, ...tl.actual].sort((a, b) => a.start_minute - b.start_minute) : [];
          const shown = items.slice(0, MAX_CHIPS);
          const overflow = items.length - shown.length;
          return (
            <div
              key={iso}
              className={cx("tl-month__cell", outside && "tl-month__cell--outside")}
              ref={isToday ? todayCellRef : undefined}
            >
              <button
                type="button"
                className="tl-month__daynum"
                onClick={() => onJumpToDay(iso)}
                aria-label={d.toLocaleDateString(undefined, { weekday: "long", day: "numeric", month: "long", year: "numeric" })}
              >
                <span className={isToday ? "tl-month__today-ring" : undefined}>{d.getDate()}</span>
              </button>
              <ul className="tl-month__chips">
                {shown.map((b) => (
                  <li key={`${b.kind}-${b.id}`}>
                    <button
                      type="button"
                      className={`tl-month__chip${b.kind === "planned" ? " tl-month__chip--planned" : ""}`}
                      style={{ "--tl-cat": categoryColor(b.category_id) } as CSSProperties}
                      onClick={() => onPick(b)}
                      aria-label={`${catName(b)} — ${b.kind}, ${fmtMinute(b.start_minute)}–${fmtMinute(b.end_minute)}. Edit.`}
                    >
                      {fmtMinute(b.start_minute)} {catName(b)}
                    </button>
                  </li>
                ))}
                {overflow > 0 && (
                  <li>
                    <button type="button" className="tl-month__more" onClick={() => onJumpToDay(iso)}>
                      +{overflow} more
                    </button>
                  </li>
                )}
              </ul>
            </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
