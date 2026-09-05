import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { api, type HabitList } from "../../api";
import { ScreenLayout } from "../../shell/ScreenLayout";
import { PageHeader } from "../../components/layout/PageHeader";
import { Button } from "../../components/ui/Button";
import { IconButton } from "../../components/ui/IconButton";
import { SegmentedControl } from "../../components/ui/SegmentedControl";
import { StatCard } from "../../components/productivity/StatCard";
import { ErrorState } from "../../components/productivity/states";
import { ChevronDownIcon } from "../../components/ui/icons";
import {
  todayISO,
  shiftDays,
  isoWeekRange,
  formatMonthLabel,
} from "../../components/date/dateUtils";
import { fetchWeek, type WeekData } from "./habitData";
import { HabitTodayList } from "./HabitTodayList";
import { HabitWeekGrid } from "./HabitWeekGrid";
import { HabitMonthHeatmap } from "./HabitMonthHeatmap";
import { HabitAllList } from "./HabitAllList";
import { HabitDialog } from "./HabitDialog";

type View = "today" | "week" | "month" | "all";
const VIEWS = [
  { value: "today", label: "Today" },
  { value: "week", label: "This week" },
  { value: "month", label: "This month" },
  { value: "all", label: "All habits" },
] as const;

const ISO_RE = /^\d{4}-\d{2}-\d{2}$/;

export function HabitsScreen() {
  const [params, setParams] = useSearchParams();
  const view: View = (VIEWS.find((v) => v.value === params.get("view"))?.value as View) ?? "today";
  const rawWeek = params.get("week");
  const weekAnchor = rawWeek && ISO_RE.test(rawWeek) ? rawWeek : todayISO();

  const [base, setBase] = useState<HabitList | null>(null);
  const [week, setWeek] = useState<WeekData | null>(null);
  const [error, setError] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);

  const setView = useCallback(
    (v: View) =>
      setParams(
        (p) => {
          const next = new URLSearchParams(p);
          if (v === "today") next.delete("view");
          else next.set("view", v);
          if (v !== "week") next.delete("week");
          return next;
        },
        { replace: true },
      ),
    [setParams],
  );
  const setWeekAnchor = useCallback(
    (iso: string) =>
      setParams(
        (p) => {
          const next = new URLSearchParams(p);
          const [thisMonday] = isoWeekRange(todayISO());
          const [targetMonday] = isoWeekRange(iso);
          if (targetMonday === thisMonday) next.delete("week");
          else next.set("week", iso);
          return next;
        },
        { replace: true },
      ),
    [setParams],
  );

  const reload = useCallback(async () => {
    setError(false);
    try {
      if (view === "week") {
        const wd = await fetchWeek(weekAnchor);
        setWeek(wd);
        setBase({ date: wd.today, habits: wd.habits, archived: wd.archived });
      } else {
        setWeek(null);
        // Request the exact date the screen renders (browser today). Omitting
        // `date` would return the *account* timezone's today, which can be a
        // different calendar day — rows, KPI and toggles would disagree.
        setBase(await api.habits(todayISO()));
      }
    } catch {
      setError(true);
    }
  }, [view, weekAnchor]);

  useEffect(() => {
    void reload();
  }, [reload]);

  async function toggle(habitId: string, date: string, done: boolean) {
    // optimistic
    setWeek((w) =>
      w ? { ...w, completion: { ...w.completion, [habitId]: { ...w.completion[habitId], [date]: done } } } : w,
    );
    setBase((b) =>
      b && date === todayISO()
        ? { ...b, habits: b.habits.map((h) => (h.id === habitId ? { ...h, completed_on_date: done } : h)) }
        : b,
    );
    try {
      await (done ? api.markHabit(habitId, date) : api.unmarkHabit(habitId, date));
    } catch {
      setError(true);
    }
    await reload();
  }

  async function archive(habitId: string) {
    if (!window.confirm("Archive this habit? Its completion history is kept.")) return;
    try {
      await api.archiveHabit(habitId);
      await reload();
    } catch {
      setError(true);
    }
  }
  async function unarchive(habitId: string) {
    try {
      await api.unarchiveHabit(habitId);
      await reload();
    } catch {
      setError(true);
    }
  }

  const openAdd = () => setDialogOpen(true);
  const today = todayISO();
  const [weekMonday, weekSunday] = isoWeekRange(weekAnchor);
  const isThisWeek = weekMonday === isoWeekRange(today)[0];

  const kpis = useMemo(() => {
    const habits = base?.habits ?? [];
    if (habits.length === 0) return null;
    return {
      completedToday: habits.filter((h) => h.completed_on_date).length,
      total: habits.length,
      bestStreak: habits.reduce((m, h) => Math.max(m, h.current_streak), 0),
    };
  }, [base]);

  return (
    <ScreenLayout>
      <PageHeader
        eyebrow="Habits"
        title="Habits"
        subtitle="Your daily habits and current streaks."
        actions={<Button onClick={openAdd}>Add habit</Button>}
      />

      <div className="habit-toolbar">
        <SegmentedControl label="Habit view" options={VIEWS} value={view} onChange={setView} />
        {view === "week" && (
          <div className="habit-toolbar__week">
            <IconButton label="Previous week" size="sm" onClick={() => setWeekAnchor(shiftDays(weekAnchor, -7))}>
              <ChevronDownIcon style={{ transform: "rotate(90deg)" }} width={16} height={16} />
            </IconButton>
            <span className="habit-toolbar__weeklabel">
              {formatMonthLabel(weekMonday) === formatMonthLabel(weekSunday)
                ? formatMonthLabel(weekMonday)
                : `${formatMonthLabel(weekMonday)} – ${formatMonthLabel(weekSunday)}`}
            </span>
            <IconButton label="Next week" size="sm" onClick={() => setWeekAnchor(shiftDays(weekAnchor, 7))}>
              <ChevronDownIcon style={{ transform: "rotate(-90deg)" }} width={16} height={16} />
            </IconButton>
            <Button variant="secondary" size="sm" onClick={() => setWeekAnchor(today)} disabled={isThisWeek}>
              This week
            </Button>
          </div>
        )}
      </div>

      {kpis && view !== "all" && (
        <div className="habit-kpis">
          <StatCard label="Completed today" value={`${kpis.completedToday} / ${kpis.total}`} tint="success" />
          <StatCard label="Active habits" value={kpis.total} />
          <StatCard label="Best current streak" value={kpis.bestStreak} sublabel="days" />
        </div>
      )}

      {error && (
        <ErrorState message="Could not load your habits." action={<Button onClick={reload}>Retry</Button>} />
      )}

      {!base ? (
        <p className="muted">Loading…</p>
      ) : view === "week" ? (
        week ? (
          <HabitWeekGrid week={week} onToggle={toggle} onArchive={archive} onAdd={openAdd} />
        ) : (
          <p className="muted">Loading…</p>
        )
      ) : view === "month" ? (
        <HabitMonthHeatmap habits={base.habits} today={today} onArchive={archive} onAdd={openAdd} />
      ) : view === "all" ? (
        <HabitAllList
          habits={base.habits}
          archived={base.archived}
          onArchive={archive}
          onUnarchive={unarchive}
          onAdd={openAdd}
        />
      ) : (
        <HabitTodayList
          habits={base.habits}
          date={today}
          completion={{}}
          onToggle={toggle}
          onArchive={archive}
          onAdd={openAdd}
        />
      )}

      {dialogOpen && (
        <HabitDialog
          open
          onClose={() => setDialogOpen(false)}
          onSaved={async () => {
            setDialogOpen(false);
            await reload();
          }}
        />
      )}
    </ScreenLayout>
  );
}
