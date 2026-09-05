import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { TimeByCategoryReport } from "./TimeByCategoryReport";
import type { CategoryTime } from "./reportsData";

const ROWS: CategoryTime[] = [
  { categoryId: "c1", categoryName: "Deep Work", seconds: 3600 * 10 },
  { categoryId: null, categoryName: "Uncategorized", seconds: 3600 * 2 },
];

describe("TimeByCategoryReport", () => {
  it("shows every row's name and a literal duration value", () => {
    render(<TimeByCategoryReport rows={ROWS} />);
    expect(screen.getByText("Deep Work")).toBeDefined();
    expect(screen.getByText("Uncategorized")).toBeDefined();
    expect(screen.getByText("10h 0m")).toBeDefined();
    expect(screen.getByText("2h 0m")).toBeDefined();
  });

  it("shows the overall total in the caption", () => {
    render(<TimeByCategoryReport rows={ROWS} />);
    expect(screen.getByText(/12h 0m overall/)).toBeDefined();
  });

  it("shows an empty state with no rows or no time", () => {
    const { rerender } = render(<TimeByCategoryReport rows={[]} />);
    expect(screen.getByText("No actual time in this range.")).toBeDefined();
    rerender(
      <TimeByCategoryReport
        rows={[{ categoryId: "c1", categoryName: "Deep Work", seconds: 0 }]}
      />,
    );
    expect(screen.getByText("No actual time in this range.")).toBeDefined();
  });
});
