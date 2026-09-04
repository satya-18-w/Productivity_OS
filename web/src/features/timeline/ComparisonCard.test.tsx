import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { ComparisonCard } from "./ComparisonCard";
import type { DayComparison } from "../../api";

const CMP: DayComparison = {
  date: "2026-09-04",
  categories: [
    { category_id: "c1", category_name: "Deep Work", planned_seconds: 7200, actual_seconds: 5400, difference_seconds: -1800 },
    { category_id: null, category_name: "Uncategorized", planned_seconds: 0, actual_seconds: 3600, difference_seconds: 3600 },
  ],
};

describe("ComparisonCard", () => {
  it("renders a Uncategorized row and a signed difference", () => {
    render(<ComparisonCard comparison={CMP} />);
    expect(screen.getByRole("heading", { name: "Planned vs actual" })).toBeDefined();
    expect(screen.getByText("Uncategorized")).toBeDefined();
    expect(screen.getByText("−30m")).toBeDefined(); // deep work difference
    expect(screen.getByText("+1h 0m")).toBeDefined(); // uncategorized difference
  });

  it("sums the totals row", () => {
    render(<ComparisonCard comparison={CMP} />);
    const foot = screen.getByRole("table").querySelector("tfoot")!;
    expect(foot.textContent).toContain("2h 0m"); // planned total
    expect(foot.textContent).toContain("2h 30m"); // actual total
  });

  it("shows an empty message when there is no data", () => {
    render(<ComparisonCard comparison={{ date: "2026-09-04", categories: [] }} />);
    expect(screen.getByText("No time planned or logged for this date.")).toBeDefined();
  });
});
