import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { PlannedVsActualReport } from "./PlannedVsActualReport";
import type { PlannedVsActualRow } from "./reportsData";

const ROWS: PlannedVsActualRow[] = [
  { categoryId: "c1", categoryName: "Deep Work", plannedSeconds: 3600 * 10, actualSeconds: 3600 * 8, differenceSeconds: -3600 * 2 },
  { categoryId: "c2", categoryName: "Admin", plannedSeconds: 3600 * 5, actualSeconds: 3600 * 6, differenceSeconds: 3600 },
];

describe("PlannedVsActualReport", () => {
  it("renders a row per category with planned/actual/difference", () => {
    render(<PlannedVsActualReport rows={ROWS} />);
    expect(screen.getByText("Deep Work")).toBeDefined();
    expect(screen.getByText("Admin")).toBeDefined();
    expect(screen.getByText("−2h 0m")).toBeDefined();
    expect(screen.getByText("+1h 0m")).toBeDefined();
  });

  it("shows a totals row that sums the column", () => {
    render(<PlannedVsActualReport rows={ROWS} />);
    expect(screen.getByText("Total")).toBeDefined();
    expect(screen.getByText("15h 0m")).toBeDefined(); // planned total
  });

  it("shows an empty state with no rows", () => {
    render(<PlannedVsActualReport rows={[]} />);
    expect(screen.getByText("No planned or actual time in this range.")).toBeDefined();
  });

  it("colours the totals-row difference negative when the range is net under plan", () => {
    render(<PlannedVsActualReport rows={ROWS} />);
    // Totals: planned 15h, actual 14h → diff −1h.
    expect(screen.getByText("−1h 0m").closest("td")?.className).toMatch(/neg/);
  });

  it("colours the totals-row difference positive when the range is net over plan", () => {
    render(
      <PlannedVsActualReport
        rows={[
          { categoryId: "c1", categoryName: "Deep Work", plannedSeconds: 3600 * 5, actualSeconds: 3600 * 8, differenceSeconds: 3600 * 3 },
        ]}
      />,
    );
    const foot = document.querySelector("tfoot")!;
    expect(within(foot).getByText("+3h 0m").closest("td")?.className).toMatch(/pos/);
  });
});
