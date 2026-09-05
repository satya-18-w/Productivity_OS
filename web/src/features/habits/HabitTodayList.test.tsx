import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HabitTodayList } from "./HabitTodayList";
import type { HabitView } from "../../api";

const HABITS: HabitView[] = [
  { id: "h1", name: "Workout", current_streak: 12, completed_on_date: false, last_30_days: 20 },
  { id: "h2", name: "Read", current_streak: 4, completed_on_date: true, last_30_days: 15 },
];

function setup(over: Partial<Parameters<typeof HabitTodayList>[0]> = {}) {
  const props = {
    habits: HABITS,
    date: "2026-09-04",
    completion: {},
    onToggle: vi.fn(),
    onArchive: vi.fn(),
    onAdd: vi.fn(),
    ...over,
  };
  render(<HabitTodayList {...props} />);
  return props;
}

describe("HabitTodayList", () => {
  it("shows a toggle, name and streak per habit", () => {
    setup();
    expect(screen.getByText("Workout")).toBeDefined();
    expect(screen.getByText("12")).toBeDefined();
    expect((screen.getByRole("checkbox", { name: /Read —.*completed/ }) as HTMLInputElement).checked).toBe(true);
  });

  it("shows the last-30 count per habit", () => {
    setup();
    expect(screen.getByText("20 of the last 30 days")).toBeDefined();
    expect(screen.getByText("15 of the last 30 days")).toBeDefined();
  });

  it("toggles a habit for the date", async () => {
    const { onToggle } = setup();
    await userEvent.click(screen.getByRole("checkbox", { name: /Workout —.*not completed/ }));
    expect(onToggle).toHaveBeenCalledWith("h1", "2026-09-04", true);
  });

  it("shows an empty state when there are no habits", () => {
    setup({ habits: [] });
    expect(screen.getByText("No habits yet")).toBeDefined();
  });

  it("kebab only offers Archive (no rename / delete — not V1)", async () => {
    setup();
    await userEvent.click(screen.getAllByRole("button", { name: /Actions for Workout/ })[0]);
    const items = screen.getAllByRole("menuitem").map((m) => m.textContent);
    expect(items).toEqual(["Archive"]);
  });
});
