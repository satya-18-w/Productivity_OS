import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { DailyActualTotalsReport } from "./DailyActualTotalsReport";
import type { DailyTotal } from "./reportsData";

const ROWS: DailyTotal[] = [
  { date: "2026-08-01", seconds: 3600 * 4 },
  { date: "2026-08-02", seconds: 3600 * 8 },
];

describe("DailyActualTotalsReport", () => {
  it("renders one bar per day with a tooltip carrying the literal value", () => {
    const { container } = render(<DailyActualTotalsReport rows={ROWS} />);
    const bars = container.querySelectorAll(".report-vbar__bar");
    expect(bars).toHaveLength(2);
    expect(bars[0].getAttribute("title")).toBe("2026-08-01: 4h 0m");
    expect(bars[1].getAttribute("title")).toBe("2026-08-02: 8h 0m");
    expect(bars[0].getAttribute("aria-label")).toBe("2026-08-01: 4h 0m");
    expect(bars[1].getAttribute("aria-label")).toBe("2026-08-02: 8h 0m");
  });

  it("summarises the range total in the caption", () => {
    render(<DailyActualTotalsReport rows={ROWS} />);
    expect(screen.getByText(/12h 0m across 2 days/)).toBeDefined();
  });

  it("shows an empty state with no range", () => {
    render(<DailyActualTotalsReport rows={[]} />);
    expect(screen.getByText("No actual time in this range.")).toBeDefined();
  });
});
