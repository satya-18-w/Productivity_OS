import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import {
  api,
  type Category,
  type DayComparison,
  type DayTimeline,
  type PositionedBlock,
  type Task,
} from "../../api";
import { ScreenLayout } from "../../shell/ScreenLayout";
import { PageHeader } from "../../components/layout/PageHeader";
import { Card } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { SplitButton } from "../../components/ui/SplitButton";
import { SegmentedControl } from "../../components/ui/SegmentedControl";
import { ErrorState } from "../../components/productivity/states";
import { MiniCalendar } from "../../components/date/MiniCalendar";
import { DateStepper } from "../../components/date/DateStepper";
import {
  todayISO,
  formatFullDate,
  formatMonthLabel,
  formatShortDate,
  parseISODate,
  shiftDays,
  shiftMonths,
  isoWeekRange,
} from "../../components/date/dateUtils";
import { TimelineGrid } from "./TimelineGrid";
import { AgendaList } from "./AgendaList";
import { TodayTasks } from "./TodayTasks";
import { ComparisonCard } from "./ComparisonCard";
import { PomodoroCard } from "./PomodoroCard";
import { WeekView } from "./WeekView";
import { MonthView } from "./MonthView";
import { BlockDialog, type BlockDialogTarget } from "./BlockDialog";
import { nowMinutes } from "./timelineFormat";

const ISO_RE = /^\d{4}-\d{2}-\d{2}$/;
type View = "day" | "agenda" | "week" | "month";
const VIEW_OPTIONS = [
  { value: "day", label: "Day" },
  { value: "agenda", label: "Agenda" },
  { value: "week", label: "Week" },
  { value: "month", label: "Month" },
] as const;

/** `‹`/`›` step size per view (G2 — Week/Month step by their own unit). */
function stepDate(view: View, date: string, direction: -1 | 1): string {
  if (view === "week") return shiftDays(date, direction * 7);
  if (view === "month") return shiftMonths(date, direction);
  return shiftDays(date, direction);
}

export function TimelineScreen() {
  const [params, setParams] = useSearchParams();

  const rawDate = params.get("date");
  const date =
    rawDate && ISO_RE.test(rawDate) && !Number.isNaN(parseISODate(rawDate).getTime())
      ? rawDate
      : todayISO();
  const rawView = params.get("view");
  const view: View = rawView === "agenda" || rawView === "week" || rawView === "month" ? rawView : "day";

  const [timeline, setTimeline] = useState<DayTimeline | null>(null);
  const [comparison, setComparison] = useState<DayComparison | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [error, setError] = useState(false);
  const [dialog, setDialog] = useState<BlockDialogTarget | null>(null);
  // Bumped after a save/delete from a Week/Month block dialog — those views
  // manage their own fetch internally, keyed on their date range, so a
  // fresh key is the simplest way to make them refetch (they don't share
  // `timeline`/`load` below, which only covers Day/Agenda's single date).
  const [refreshKey, setRefreshKey] = useState(0);

  const patchParams = useCallback(
    (mut: (p: URLSearchParams) => void) => {
      setParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          mut(next);
          return next;
        },
        { replace: true },
      );
    },
    [setParams],
  );

  const setDate = useCallback(
    (iso: string) => patchParams((p) => (iso === todayISO() ? p.delete("date") : p.set("date", iso))),
    [patchParams],
  );
  const setView = useCallback(
    (v: View) => patchParams((p) => (v === "day" ? p.delete("view") : p.set("view", v))),
    [patchParams],
  );
  /** Week/Month's day-header "jump to this date" — switches to Day atomically. */
  const jumpToDay = useCallback(
    (iso: string) =>
      patchParams((p) => {
        p.delete("view");
        if (iso === todayISO()) p.delete("date");
        else p.set("date", iso);
      }),
    [patchParams],
  );

  const load = useCallback(async () => {
    setError(false);
    try {
      const [tl, cmp, cats, board] = await Promise.all([
        api.timeline(date),
        api.comparison(date),
        api.listCategories(),
        api.board(),
      ]);
      setTimeline(tl);
      setComparison(cmp);
      setCategories(cats);
      setTasks(board.columns.flatMap((c) => c.tasks));
    } catch {
      setError(true);
    }
  }, [date]);

  useEffect(() => {
    void load();
  }, [load]);

  const taskTitleById = useMemo(() => new Map(tasks.map((t) => [t.id, t.title])), [tasks]);

  const isToday = date === todayISO();
  const pick = (b: PositionedBlock) => setDialog({ mode: "edit", block: b });

  // A task's "Scheduled blocks" list deep-links here as
  // ?date=<local_date>&openBlock=<id> (the reverse of "Open task" on a linked
  // block) — once that date's timeline has loaded, open the matching block's
  // dialog, then drop the param so it doesn't reopen on a later re-render.
  useEffect(() => {
    const openId = params.get("openBlock");
    if (!openId || !timeline) return;
    const found = [...timeline.planned, ...timeline.actual].find((b) => b.id === openId);
    if (found) pick(found);
    patchParams((p) => p.delete("openBlock"));
    // eslint-disable-next-line react-hooks/exhaustive-deps -- pick/patchParams are stable per render intent
  }, [params, timeline]);

  const headerTitle =
    view === "week"
      ? (() => {
          const [mon, sun] = isoWeekRange(date);
          return `${formatShortDate(mon)} – ${formatShortDate(sun)}`;
        })()
      : view === "month"
        ? formatMonthLabel(date)
        : formatFullDate(date);

  return (
    <ScreenLayout
      railLabel="Timeline calendar and tasks"
      rail={
        <>
          <Card padding="compact">
            <MiniCalendar value={date} onChange={setDate} />
          </Card>
          {view === "day" && <PomodoroCard />}
          <TodayTasks date={date} />
        </>
      }
    >
      <PageHeader
        eyebrow="Timeline"
        title={headerTitle}
        subtitle="Plan the day as time blocks and log what actually happened."
        actions={
          <SplitButton
            onPrimary={() => setDialog({ mode: "new" })}
            menuLabel="Add block options"
            items={[
              {
                key: "planned",
                label: "Add planned block",
                onSelect: () => setDialog({ mode: "new", kind: "planned" }),
              },
              {
                key: "actual",
                label: "Add actual block",
                onSelect: () => setDialog({ mode: "new", kind: "actual" }),
              },
            ]}
          >
            Add block
          </SplitButton>
        }
      />

      <div className="tl2__toolbar">
        <SegmentedControl label="Timeline view" options={VIEW_OPTIONS} value={view} onChange={setView} />
        <span style={{ flex: 1 }} />
        <DateStepper
          value={date}
          onChange={setDate}
          onStep={(direction) => setDate(stepDate(view, date, direction))}
          prevLabel={view === "week" ? "Previous week" : view === "month" ? "Previous month" : "Previous day"}
          nextLabel={view === "week" ? "Next week" : view === "month" ? "Next month" : "Next day"}
        />
      </div>

      {view === "week" ? (
        <WeekView key={refreshKey} date={date} onPick={pick} onJumpToDay={jumpToDay} />
      ) : view === "month" ? (
        <MonthView key={refreshKey} date={date} onPick={pick} onJumpToDay={jumpToDay} />
      ) : error ? (
        <ErrorState message="Could not load the timeline." action={<Button onClick={load}>Retry</Button>} />
      ) : !timeline ? (
        <p className="muted">Loading…</p>
      ) : (
        <>
          {view === "agenda" ? (
            <AgendaList
              planned={timeline.planned}
              actual={timeline.actual}
              now={isToday ? nowMinutes() : null}
              taskTitleById={taskTitleById}
              onPick={pick}
              onAdd={() => setDialog({ mode: "new" })}
            />
          ) : (
            <div className="tl2">
              <TimelineGrid
                planned={timeline.planned}
                actual={timeline.actual}
                now={isToday ? nowMinutes() : null}
                taskTitleById={taskTitleById}
                onPick={pick}
              />
            </div>
          )}
          <ComparisonCard comparison={comparison} />
        </>
      )}

      {dialog && (
        <BlockDialog
          open
          target={dialog}
          date={date}
          categories={categories}
          tasks={tasks}
          onClose={() => setDialog(null)}
          onSaved={async () => {
            setDialog(null);
            await load();
            setRefreshKey((k) => k + 1); // Week/Month manage their own fetch; force a remount to refresh
          }}
        />
      )}
    </ScreenLayout>
  );
}
