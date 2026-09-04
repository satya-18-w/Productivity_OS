import type { DayComparison } from "../../api";
import { Card } from "../../components/ui/Card";
import { fmtDuration } from "./timelineFormat";

export function ComparisonCard({ comparison }: { comparison: DayComparison | null }) {
  return (
    <Card title="Planned vs actual" headingLevel={2}>
      {!comparison ? (
        <p className="muted">Loading…</p>
      ) : comparison.categories.length === 0 ? (
        <p className="muted">No time planned or logged for this date.</p>
      ) : (
        <ComparisonTable comparison={comparison} />
      )}
    </Card>
  );
}

function ComparisonTable({ comparison }: { comparison: DayComparison }) {
  const totals = comparison.categories.reduce(
    (acc, c) => ({
      planned: acc.planned + c.planned_seconds,
      actual: acc.actual + c.actual_seconds,
      diff: acc.diff + c.difference_seconds,
    }),
    { planned: 0, actual: 0, diff: 0 },
  );

  return (
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
          {comparison.categories.map((c) => (
            <tr key={c.category_id ?? "uncategorized"}>
              <td>{c.category_name}</td>
              <td>{fmtDuration(c.planned_seconds)}</td>
              <td>{fmtDuration(c.actual_seconds)}</td>
              <td className={c.difference_seconds < 0 ? "neg" : c.difference_seconds > 0 ? "pos" : ""}>
                {c.difference_seconds > 0 ? "+" : ""}
                {fmtDuration(c.difference_seconds)}
              </td>
            </tr>
          ))}
        </tbody>
        <tfoot>
          <tr>
            <td>Total</td>
            <td>{fmtDuration(totals.planned)}</td>
            <td>{fmtDuration(totals.actual)}</td>
            <td>
              {totals.diff > 0 ? "+" : ""}
              {fmtDuration(totals.diff)}
            </td>
          </tr>
        </tfoot>
      </table>
    </div>
  );
}
