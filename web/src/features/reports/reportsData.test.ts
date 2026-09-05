import { describe, it, expect } from "vitest";
import { mockReportData, defaultRange } from "./reportsData";

describe("mockReportData", () => {
  it("is deterministic for the same range", () => {
    const a = mockReportData("2026-08-01", "2026-08-07");
    const b = mockReportData("2026-08-01", "2026-08-07");
    expect(a).toEqual(b);
  });

  it("produces one daily total per day in the inclusive range", () => {
    const data = mockReportData("2026-08-01", "2026-08-07");
    expect(data.dailyActualTotals).toHaveLength(7);
    expect(data.dailyActualTotals[0].date).toBe("2026-08-01");
    expect(data.dailyActualTotals[6].date).toBe("2026-08-07");
  });

  it("includes an explicit Uncategorized bucket in time-by-category", () => {
    const data = mockReportData("2026-08-01", "2026-08-07");
    const uncategorized = data.timeByCategory.find((r) => r.categoryId === null);
    expect(uncategorized).toBeDefined();
    expect(uncategorized?.categoryName).toBe("Uncategorized");
  });

  it("only includes named categories (never Uncategorized) in planned vs actual", () => {
    const data = mockReportData("2026-08-01", "2026-08-07");
    expect(data.plannedVsActual.every((r) => r.categoryId !== null)).toBe(true);
    expect(data.plannedVsActual.length).toBeGreaterThan(0);
  });

  it("computes differenceSeconds as actual minus planned", () => {
    const data = mockReportData("2026-08-01", "2026-08-07");
    for (const r of data.plannedVsActual) {
      expect(r.differenceSeconds).toBe(r.actualSeconds - r.plannedSeconds);
    }
  });

  it("gives every habit the same rangeDays as the number of days in range", () => {
    const data = mockReportData("2026-08-01", "2026-08-10");
    for (const r of data.habitCompletion) {
      expect(r.rangeDays).toBe(10);
    }
  });

  it("varies with a different range", () => {
    const a = mockReportData("2026-08-01", "2026-08-07");
    const b = mockReportData("2026-01-01", "2026-01-31");
    expect(a).not.toEqual(b);
  });
});

describe("defaultRange", () => {
  it("spans 30 days ending today", () => {
    const { from, to } = defaultRange();
    const days = (new Date(to).getTime() - new Date(from).getTime()) / 86_400_000;
    expect(Math.round(days)).toBe(29);
  });
});
