import { Card } from "../../components/ui/Card";
import { categoryColor } from "../../components/productivity/categoryColor";
import { fmtDuration } from "../timeline/timelineFormat";
import type { CategoryTime } from "./reportsData";

export function TimeByCategoryReport({ rows }: { rows: CategoryTime[] }) {
  const total = rows.reduce((s, r) => s + r.seconds, 0);
  const max = Math.max(1, ...rows.map((r) => r.seconds));

  return (
    <Card title="Time by category" headingLevel={2}>
      {rows.length === 0 || total === 0 ? (
        <p className="muted">No actual time in this range.</p>
      ) : (
      <>
      <p className="secondary report-caption">Total actual time per category, {fmtDuration(total)} overall.</p>
      <ul className="report-hbars" aria-label="Time by category">
        {rows.map((r) => {
          const color = categoryColor(r.categoryId);
          const pct = total === 0 ? 0 : Math.round((r.seconds / total) * 100);
          return (
            <li key={r.categoryId ?? "uncategorized"} className="report-hbar">
              <span className="report-hbar__label">{r.categoryName}</span>
              <span className="report-hbar__track" title={`${r.categoryName}: ${fmtDuration(r.seconds)} (${pct}%)`}>
                <span
                  className="report-hbar__fill"
                  style={{ width: `${(r.seconds / max) * 100}%`, background: color }}
                />
              </span>
              <span className="report-hbar__value">{fmtDuration(r.seconds)}</span>
            </li>
          );
        })}
      </ul>
      </>
      )}
    </Card>
  );
}
