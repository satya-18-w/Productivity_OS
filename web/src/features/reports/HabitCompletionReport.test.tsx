import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { HabitCompletionReport } from "./HabitCompletionReport";
import type { HabitCompletionRow } from "./reportsData";

const ROWS: HabitCompletionRow[] = [
  { habitId: "h1", habitName: "Workout", completedDays: 5, rangeDays: 10 },
];

describe("HabitCompletionReport", () => {
  it("shows the habit name and completed/range figure with a computed rate", () => {
    render(<HabitCompletionReport rows={ROWS} />);
    expect(screen.getByText("Workout")).toBeDefined();
    expect(screen.getByText("5 / 10 days (50%)")).toBeDefined();
  });

  it("shows an empty state with no habits", () => {
    render(<HabitCompletionReport rows={[]} />);
    expect(screen.getByText("No habits in this range.")).toBeDefined();
  });
});
