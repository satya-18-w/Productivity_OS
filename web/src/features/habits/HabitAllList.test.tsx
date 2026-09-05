import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HabitAllList } from "./HabitAllList";
import type { ArchivedHabit, HabitView } from "../../api";

const HABITS: HabitView[] = [
  { id: "h1", name: "Workout", current_streak: 12, completed_on_date: false, last_30_days: 20 },
  { id: "h2", name: "Read", current_streak: 4, completed_on_date: true, last_30_days: 15 },
];
const ARCHIVED: ArchivedHabit[] = [{ id: "h9", name: "Old habit" }];

function setup() {
  const props = {
    habits: HABITS,
    archived: ARCHIVED,
    onArchive: vi.fn(),
    onUnarchive: vi.fn(),
    onAdd: vi.fn(),
  };
  render(<HabitAllList {...props} />);
  return props;
}

describe("HabitAllList", () => {
  it("shows active habits with streak and last-30 count", () => {
    setup();
    expect(screen.getByText("Workout")).toBeDefined();
    expect(screen.getByText("20 of the last 30 days")).toBeDefined();
    expect(screen.getByText("15 of the last 30 days")).toBeDefined();
  });

  it("lists archived habits with an Unarchive action (history kept)", () => {
    const { onUnarchive } = setup();
    expect(screen.getByText("Old habit")).toBeDefined();
    return userEvent.click(screen.getByRole("button", { name: "Unarchive" })).then(() => {
      expect(onUnarchive).toHaveBeenCalledWith("h9");
    });
  });

  it("kebab on active rows only offers Archive (no rename / delete — not V1)", async () => {
    setup();
    await userEvent.click(screen.getAllByRole("button", { name: /Actions for Workout/ })[0]);
    const items = screen.getAllByRole("menuitem").map((m) => m.textContent);
    expect(items).toEqual(["Archive"]);
  });
});
