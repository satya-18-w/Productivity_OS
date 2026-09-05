import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HabitWeekGrid } from "./HabitWeekGrid";
import type { WeekData } from "./habitData";

const WEEK: WeekData = {
  weekStart: "2026-08-31",
  days: ["2026-08-31", "2026-09-01", "2026-09-02", "2026-09-03", "2026-09-04", "2026-09-05", "2026-09-06"],
  today: "2026-09-04",
  habits: [
    { id: "h1", name: "Workout", current_streak: 12, completed_on_date: false, last_30_days: 20 },
    { id: "h2", name: "Read", current_streak: 4, completed_on_date: true, last_30_days: 15 },
  ],
  archived: [],
  completion: {
    h1: { "2026-08-31": true, "2026-09-01": true, "2026-09-04": true },
    h2: { "2026-09-04": true },
  },
};

describe("HabitWeekGrid", () => {
  it("renders a row per habit and 7 dated day columns", () => {
    render(<HabitWeekGrid week={WEEK} onToggle={vi.fn()} onArchive={vi.fn()} onAdd={vi.fn()} />);
    expect(screen.getByRole("rowheader", { name: "Workout" })).toBeDefined();
    expect(screen.getByRole("rowheader", { name: "Read" })).toBeDefined();
    // 7 toggle circles per habit row
    expect(screen.getAllByRole("checkbox", { name: /Workout —/ })).toHaveLength(7);
  });

  it("reflects the completion map and streaks", () => {
    render(<HabitWeekGrid week={WEEK} onToggle={vi.fn()} onArchive={vi.fn()} onAdd={vi.fn()} />);
    const mon = screen.getByRole("checkbox", { name: /Workout — Mon 31, completed/ }) as HTMLInputElement;
    expect(mon.checked).toBe(true);
    const wed = screen.getByRole("checkbox", { name: /Workout — Wed 2, not completed/ }) as HTMLInputElement;
    expect(wed.checked).toBe(false);
    expect(screen.getByText("12")).toBeDefined();
  });

  it("shows a Last 30 column with each habit's last-30 count", () => {
    render(<HabitWeekGrid week={WEEK} onToggle={vi.fn()} onArchive={vi.fn()} onAdd={vi.fn()} />);
    expect(screen.getByRole("columnheader", { name: "Last 30" })).toBeDefined();
    expect(screen.getByText("20 of the last 30 days")).toBeDefined();
    expect(screen.getByText("15 of the last 30 days")).toBeDefined();
  });

  it("toggling a cell calls onToggle with habit id, date and new state", async () => {
    const onToggle = vi.fn();
    render(<HabitWeekGrid week={WEEK} onToggle={onToggle} onArchive={vi.fn()} onAdd={vi.fn()} />);
    await userEvent.click(screen.getByRole("checkbox", { name: /Read — Tue 1, not completed/ }));
    expect(onToggle).toHaveBeenCalledWith("h2", "2026-09-01", true);
  });

  it("archives a habit from its row menu", async () => {
    const onArchive = vi.fn();
    render(<HabitWeekGrid week={WEEK} onToggle={vi.fn()} onArchive={onArchive} onAdd={vi.fn()} />);
    await userEvent.click(screen.getAllByRole("button", { name: /Actions for Workout/ })[0]);
    await userEvent.click(screen.getByRole("menuitem", { name: "Archive" }));
    expect(onArchive).toHaveBeenCalledWith("h1");
  });
});
