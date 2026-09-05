import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { GoalRow } from "./GoalRow";
import type { Goal } from "../../api";

function goal(over: Partial<Goal>): Goal {
  return {
    id: "g1", title: "Run a half marathon", description: "Train 4x a week", target_date: null,
    progress: "IN_PROGRESS", created_at: "", updated_at: "", ...over,
  };
}

function setup(over: Partial<Goal> = {}) {
  const h = { onSetProgress: vi.fn(), onEdit: vi.fn(), onDelete: vi.fn() };
  return { ...h, ...render(<GoalRow goal={goal(over)} {...h} />) };
}

describe("GoalRow", () => {
  it("shows the title, description and the V1 status label", () => {
    setup();
    expect(screen.getByRole("button", { name: "Run a half marathon" })).toBeDefined();
    expect(screen.getByText("Train 4x a week")).toBeDefined();
    expect(screen.getByText("In progress")).toBeDefined();
  });

  it("has no percentage, progress bar or linked-task affordance", () => {
    const { container } = setup();
    expect(container.querySelector(".ui-progress")).toBeNull();
    expect(container.textContent).not.toMatch(/%|\/\s*\d+\s*tasks/);
  });

  it("flags a past-due target date for an unfinished goal", () => {
    setup({ target_date: "2000-01-01", progress: "IN_PROGRESS" });
    expect(screen.getByText(/past due/)).toBeDefined();
  });

  it("does not flag past-due once achieved", () => {
    setup({ target_date: "2000-01-01", progress: "ACHIEVED" });
    expect(screen.queryByText(/past due/)).toBeNull();
  });

  it("kebab sets progress (excluding current), edits and deletes", async () => {
    const h = setup({ progress: "IN_PROGRESS" });
    await userEvent.click(screen.getByRole("button", { name: /Actions for Run a half marathon/ }));
    expect(screen.queryByRole("menuitem", { name: "Set to In progress" })).toBeNull();
    expect(screen.getByRole("menuitem", { name: "Set to Not started" })).toBeDefined();
    await userEvent.click(screen.getByRole("menuitem", { name: "Set to Achieved" }));
    expect(h.onSetProgress).toHaveBeenCalledWith(expect.objectContaining({ id: "g1" }), "ACHIEVED");
  });
});
