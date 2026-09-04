import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import {
  api,
  type Category,
  type DayComparison,
  type DayTimeline,
  type PositionedBlock,
} from "../../api";
import { ScreenLayout } from "../../shell/ScreenLayout";
import { PageHeader } from "../../components/layout/PageHeader";
import { Card } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { IconButton } from "../../components/ui/IconButton";
import { Input } from "../../components/ui/Input";
import { SegmentedControl } from "../../components/ui/SegmentedControl";
import { ErrorState } from "../../components/productivity/states";
import { MiniCalendar } from "../../components/date/MiniCalendar";
import { ChevronDownIcon } from "../../components/ui/icons";
import { todayISO, shiftDays, formatFullDate, parseISODate } from "../../components/date/dateUtils";
import { TimelineGrid } from "./TimelineGrid";
import { AgendaList } from "./AgendaList";
import { ComparisonCard } from "./ComparisonCard";
import { BlockDialog, type BlockDialogTarget } from "./BlockDialog";
import { nowMinutes } from "./timelineFormat";

const ISO_RE = /^\d{4}-\d{2}-\d{2}$/;
type View = "day" | "agenda";
const VIEW_OPTIONS = [
  { value: "day", label: "Day" },
  { value: "agenda", label: "Agenda" },
] as const;

export function TimelineScreen() {
  const [params, setParams] = useSearchParams();

  const rawDate = params.get("date");
  const date =
    rawDate && ISO_RE.test(rawDate) && !Number.isNaN(parseISODate(rawDate).getTime())
      ? rawDate
      : todayISO();
  const view: View = params.get("view") === "agenda" ? "agenda" : "day";

  const [timeline, setTimeline] = useState<DayTimeline | null>(null);
  const [comparison, setComparison] = useState<DayComparison | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [error, setError] = useState(false);
  const [dialog, setDialog] = useState<BlockDialogTarget | null>(null);

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

  const load = useCallback(async () => {
    setError(false);
    try {
      const [tl, cmp, cats] = await Promise.all([
        api.timeline(date),
        api.comparison(date),
        api.listCategories(),
      ]);
      setTimeline(tl);
      setComparison(cmp);
      setCategories(cats);
    } catch {
      setError(true);
    }
  }, [date]);

  useEffect(() => {
    void load();
  }, [load]);

  const isToday = date === todayISO();
  const pick = (b: PositionedBlock) => setDialog({ mode: "edit", block: b });

  return (
    <ScreenLayout
      railLabel="Timeline calendar"
      rail={
        <Card padding="compact">
          <MiniCalendar value={date} onChange={setDate} />
        </Card>
      }
    >
      <PageHeader
        eyebrow="Timeline"
        title={formatFullDate(date)}
        subtitle="Plan the day as time blocks and log what actually happened."
        actions={<Button onClick={() => setDialog({ mode: "new" })}>Add block</Button>}
      />

      <div className="tl2__toolbar">
        <SegmentedControl label="Timeline view" options={VIEW_OPTIONS} value={view} onChange={setView} />
        <span style={{ flex: 1 }} />
        <IconButton label="Previous day" size="sm" onClick={() => setDate(shiftDays(date, -1))}>
          <ChevronDownIcon style={{ transform: "rotate(90deg)" }} width={16} height={16} />
        </IconButton>
        <Input
          type="date"
          aria-label="Date"
          value={date}
          onChange={(e) => e.target.value && setDate(e.target.value)}
          style={{ width: "auto" }}
        />
        <IconButton label="Next day" size="sm" onClick={() => setDate(shiftDays(date, 1))}>
          <ChevronDownIcon style={{ transform: "rotate(-90deg)" }} width={16} height={16} />
        </IconButton>
        <Button variant="secondary" size="sm" onClick={() => setDate(todayISO())} disabled={isToday}>
          Today
        </Button>
      </div>

      {error ? (
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
              onPick={pick}
            />
          ) : (
            <div className="tl2">
              <TimelineGrid
                planned={timeline.planned}
                actual={timeline.actual}
                now={isToday ? nowMinutes() : null}
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
          onClose={() => setDialog(null)}
          onSaved={async () => {
            setDialog(null);
            await load();
          }}
        />
      )}
    </ScreenLayout>
  );
}
