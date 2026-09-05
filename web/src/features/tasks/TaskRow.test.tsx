import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TaskRow } from "./TaskRow";
import type { Category, Task } from "../../api";

const TODAY = "2026-09-04";

function task(over: Partial<Task>): Task {
  return {
    id: "t1", title: "Write report", description: "", due_date: null,
    state: "TODO", category_id: null, created_at: "", updated_at: "", ...over,
  };
}

function renderRow(
  over: Partial<Task> = {},
  handlers: Partial<Record<string, ReturnType<typeof vi.fn>>> = {},
  categories: Category[] = [],
) {
  const h = {
    onToggleDone: vi.fn(), onMove: vi.fn(), onEdit: vi.fn(), onDelete: vi.fn(), ...handlers,
  };
  render(
    <ul>
      <TaskRow task={task(over)} today={TODAY} categories={categories} {...h} />
    </ul>,
  );
  return h;
}

describe("TaskRow", () => {
  it("toggles done via the checkbox", async () => {
    const h = renderRow({ state: "TODO" });
    await userEvent.click(screen.getByRole("checkbox", { name: /Mark "Write report" done/ }));
    expect(h.onToggleDone).toHaveBeenCalledWith(expect.objectContaining({ id: "t1" }), true);
  });

  it("shows the checkbox checked and title struck through when done", () => {
    renderRow({ state: "DONE" });
    expect((screen.getByRole("checkbox") as HTMLInputElement).checked).toBe(true);
  });

  it("flags an overdue due date", () => {
    renderRow({ due_date: "2026-09-01" });
    expect(screen.getByText(/overdue/)).toBeDefined();
  });

  it("shows 'Today' for a task due today", () => {
    renderRow({ due_date: TODAY });
    expect(screen.getByText("Today")).toBeDefined();
  });

  it("kebab menu offers the other three states, edit and delete", async () => {
    const h = renderRow({ state: "TODO" });
    await userEvent.click(screen.getByRole("button", { name: /Actions for Write report/ }));
    expect(screen.getByRole("menuitem", { name: "Edit" })).toBeDefined();
    expect(screen.getByRole("menuitem", { name: "Move to Backlog" })).toBeDefined();
    expect(screen.getByRole("menuitem", { name: "Move to In progress" })).toBeDefined();
    expect(screen.queryByRole("menuitem", { name: "Move to To do" })).toBeNull(); // current state hidden
    await userEvent.click(screen.getByRole("menuitem", { name: "Move to Done" }));
    expect(h.onMove).toHaveBeenCalledWith(expect.objectContaining({ id: "t1" }), "DONE");
  });

  it("opens the editor when the title is clicked", async () => {
    const h = renderRow();
    await userEvent.click(screen.getByRole("button", { name: "Write report" }));
    expect(h.onEdit).toHaveBeenCalled();
  });

  it("shows the task's category, resolved by id from the category list", () => {
    renderRow({ category_id: "cat1" }, {}, [{ id: "cat1", name: "Deep Work" }]);
    expect(screen.getByText("Deep Work")).toBeDefined();
  });

  it("shows nothing category-wise when the task has none", () => {
    renderRow({ category_id: null }, {}, [{ id: "cat1", name: "Deep Work" }]);
    expect(screen.queryByText("Deep Work")).toBeNull();
  });
});
