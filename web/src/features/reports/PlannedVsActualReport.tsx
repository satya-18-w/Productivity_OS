import { Card } from "../../components/ui/Card";
import { fmtDuration } from "../timeline/timelineFormat";
import type { PlannedVsActualRow } from "./reportsData";

export function PlannedVsActualReport({ rows }: { rows: PlannedVsActualRow[] }) {
  const totals = rows.reduce(
    (acc, r) => ({
      planned: acc.planned + r.plannedSeconds,
      actual: acc.actual + r.actualSeconds,
      diff: acc.diff + r.differenceSeconds,
    }),
    { planned: 0, actual: 0, diff: 0 },
  );

  return (
    <Card title="Planned vs actual by category" headingLevel={2}>
      <p className="secondary report-caption">Planned, actual, and difference per category.</p>
      {rows.length === 0 ? (
        <p className="muted">No planned or actual time in this range.</p>
      ) : (
        <div className="tl-scroll">
          <table className="totals">
            <thead>
              <tr>
                <th scope="col">Category</th>
                <th scope="col">Planned</th>
                <th scope="col">Actual</th>
                <th scope="col">Difference</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((r) => (
                <tr key={r.categoryId ?? "uncategorized"}>
                  <td>{r.categoryName}</td>
                  <td>{fmtDuration(r.plannedSeconds)}</td>
                  <td>{fmtDuration(r.actualSeconds)}</td>
                  <td className={r.differenceSeconds < 0 ? "neg" : r.differenceSeconds > 0 ? "pos" : ""}>
                    {r.differenceSeconds > 0 ? "+" : ""}
                    {fmtDuration(r.differenceSeconds)}
                  </td>
                </tr>
              ))}
            </tbody>
            <tfoot>
              <tr>
                <td>Total</td>
                <td>{fmtDuration(totals.planned)}</td>
                <td>{fmtDuration(totals.actual)}</td>
                <td className={totals.diff < 0 ? "neg" : totals.diff > 0 ? "pos" : ""}>
                  {totals.diff > 0 ? "+" : ""}
                  {fmtDuration(totals.diff)}
                </td>
              </tr>
            </tfoot>
          </table>
        </div>
      )}
    </Card>
  );
}
