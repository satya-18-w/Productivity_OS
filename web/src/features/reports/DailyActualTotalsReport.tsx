import { Card } from "../../components/ui/Card";
import { formatShortDate } from "../../components/date/dateUtils";
import { fmtDuration } from "../timeline/timelineFormat";
import type { DailyTotal } from "./reportsData";

export function DailyActualTotalsReport({ rows }: { rows: DailyTotal[] }) {
  const max = Math.max(1, ...rows.map((r) => r.seconds));
  const total = rows.reduce((s, r) => s + r.seconds, 0);

  return (
    <Card title="Daily actual totals" headingLevel={2}>
      {rows.length === 0 ? (
        <p className="muted">No actual time in this range.</p>
      ) : (
        <>
        <p className="secondary report-caption">
          Total actual time for each day in the range, {fmtDuration(total)} across {rows.length}{" "}
          {rows.length === 1 ? "day" : "days"}.
        </p>
        <div className="report-vbars__scroll">
          <ul className="report-vbars" aria-label="Total actual time per day">
            {rows.map((r) => (
              <li key={r.date} className="report-vbar">
                <span
                  className="report-vbar__bar"
                  role="img"
                  aria-label={`${r.date}: ${fmtDuration(r.seconds)}`}
                  style={{ height: `${Math.max(2, (r.seconds / max) * 100)}%` }}
                  title={`${r.date}: ${fmtDuration(r.seconds)}`}
                />
                <span className="report-vbar__label">{formatShortDate(r.date)}</span>
              </li>
            ))}
          </ul>
        </div>
        </>
      )}
    </Card>
  );
}
