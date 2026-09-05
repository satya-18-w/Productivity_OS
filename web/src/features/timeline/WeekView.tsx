import { useEffect, useRef, useState, type CSSProperties } from "react";
import { api, type DayTimeline, type PositionedBlock } from "../../api";
import { ErrorState } from "../../components/productivity/states";
import { Button } from "../../components/ui/Button";
import { categoryColor } from "../../components/productivity/categoryColor";
import { shiftDays, todayISO, isoWeekRange, WEEKDAYS_MON_FIRST } from "../../components/date/dateUtils";
import { fmtMinute } from "./timelineFormat";

function catName(b: PositionedBlock): string {
  return b.category_name ?? "Uncategorized";
}

/**
 * Timeline — Week (G2, `design-system.md §6.1`; `screens/timeline-week.md`).
 * 7 day-columns, each a chronological stack of that day's blocks — not
 * hour-proportional (G2: too dense across 7 columns). Reuses the day view's
 * planned=dashed/actual=solid language (G1) so no new visual system is
 * introduced. Fetches the whole week in one `api.timelineRange` call (real
 * data) — see `docs/left.md`, since resolved.
 */
export function WeekView({
  date,
  onPick,
  onJumpToDay,
}: {
  /** Any date within the target ISO week. */
  date: string;
  onPick: (b: PositionedBlock) => void;
  onJumpToDay: (iso: string) => void;
}) {
  const [monday] = isoWeekRange(date);
  const weekDates = Array.from({ length: 7 }, (_, i) => shiftDays(monday, i));
  const [days, setDays] = useState<DayTimeline[] | null>(null);
  const [error, setError] = useState(false);
  const todayColRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let live = true;
    setDays(null);
    setError(false);
    api
      .timelineRange(weekDates[0], weekDates[6])
      .then((range) => {
        if (!live) return;
        const byDate = new Map(range.days.map((d) => [d.date, d]));
        setDays(weekDates.map((d) => byDate.get(d) ?? { date: d, planned: [], actual: [] }));
      })
      .catch(() => {
        if (live) setError(true);
      });
    return () => {
      live = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- keyed on the week's Monday only
  }, [monday]);

  // 7 columns need more width than the main content area gives on common
  // desktop sizes (the rail eats into it), so the grid scrolls horizontally
  // (`.tl-week__scroll`) — bring today's column into view rather than
  // defaulting to Monday, same reasoning as the Day grid's vertical anchor.
  useEffect(() => {
    if (!days) return;
    todayColRef.current?.scrollIntoView({ inline: "center", block: "nearest" });
  }, [days]);

  if (error) {
    return (
      <ErrorState
        message="Could not load this week."
        action={<Button onClick={() => setError(false)}>Retry</Button>}
      />
    );
  }
  if (!days) return <p className="muted">Loading…</p>;

  const today = todayISO();

  return (
    <div className="tl-week__scroll">
      <div className="tl-week__grid">
        {weekDates.map((d, i) => {
          const items = [...days[i].planned, ...days[i].actual].sort(
            (a, b) => a.start_minute - b.start_minute || a.end_minute - b.end_minute,
          );
          const isToday = d === today;
          const dayNum = Number(d.slice(-2));
          return (
            <div className="tl-week__col" key={d} ref={isToday ? todayColRef : undefined}>
              <button
                type="button"
                className={`tl-week__col-head${isToday ? " is-today" : ""}`}
                onClick={() => onJumpToDay(d)}
              >
                <span className="tl-week__col-weekday">{WEEKDAYS_MON_FIRST[i]}</span>
                <span className="tl-week__col-date">{dayNum}</span>
              </button>
              {items.length === 0 ? (
                <p className="tl-week__empty">—</p>
              ) : (
                <ul className="tl-week__stack">
                  {items.map((b) => (
                    <li key={`${b.kind}-${b.id}`}>
                      <button
                        type="button"
                        className={`tl-week__chip${b.kind === "planned" ? " tl-week__chip--planned" : ""}`}
                        style={{ "--tl-cat": categoryColor(b.category_id) } as CSSProperties}
                        onClick={() => onPick(b)}
                        aria-label={`${catName(b)} — ${b.kind}, ${fmtMinute(b.start_minute)}–${fmtMinute(b.end_minute)}. Edit.`}
                      >
                        <span className="tl-week__chip-time">{fmtMinute(b.start_minute)}</span>
                        <span className="tl-week__chip-cat">{catName(b)}</span>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
