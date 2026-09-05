import { useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import { ScreenLayout } from "../../shell/ScreenLayout";
import { PageHeader } from "../../components/layout/PageHeader";
import { DateRangePicker } from "./DateRangePicker";
import { TimeByCategoryReport } from "./TimeByCategoryReport";
import { PlannedVsActualReport } from "./PlannedVsActualReport";
import { HabitCompletionReport } from "./HabitCompletionReport";
import { TaskThroughputReport } from "./TaskThroughputReport";
import { DailyActualTotalsReport } from "./DailyActualTotalsReport";
import { mockReportData, defaultRange } from "./reportsData";

const ISO_RE = /^\d{4}-\d{2}-\d{2}$/;

export function ReportsScreen() {
  const [params, setParams] = useSearchParams();
  // Computed once per mount so "no range in the URL yet" doesn't drift as today changes.
  const fallback = useMemo(defaultRange, []);
  const rawFrom = params.get("from");
  const rawTo = params.get("to");
  const from = rawFrom && ISO_RE.test(rawFrom) ? rawFrom : fallback.from;
  const to = rawTo && ISO_RE.test(rawTo) ? rawTo : fallback.to;
  // Typed input can invert the range; keep URL, inputs, and data canonical.
  const effFrom = from <= to ? from : to;
  const effTo = from <= to ? to : from;

  function setRange(range: { from: string; to: string }) {
    const norm = range.from <= range.to ? range : { from: range.to, to: range.from };
    setParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        next.set("from", norm.from);
        next.set("to", norm.to);
        return next;
      },
      { replace: true },
    );
  }

  // PLACEHOLDER — no reports backend exists yet (docs/left.md, "Phase 9 — Reports").
  const data = useMemo(() => mockReportData(effFrom, effTo), [effFrom, effTo]);

  return (
    <ScreenLayout>
      <PageHeader
        eyebrow="Reports"
        title="Reports"
        subtitle="Five fixed, deterministic reports over a date range you choose."
      />

      <DateRangePicker from={effFrom} to={effTo} onChange={setRange} />

      <p className="hint report-sample-note" role="note">
        ⚠ Sample data — the reports endpoint is pending (see docs/left.md). Figures below
        are illustrative, not computed from your account.
      </p>

      <div className="reports-grid">
        <TimeByCategoryReport rows={data.timeByCategory} />
        <PlannedVsActualReport rows={data.plannedVsActual} />
        <HabitCompletionReport rows={data.habitCompletion} />
        <TaskThroughputReport count={data.taskThroughput} />
        <DailyActualTotalsReport rows={data.dailyActualTotals} />
      </div>
    </ScreenLayout>
  );
}
