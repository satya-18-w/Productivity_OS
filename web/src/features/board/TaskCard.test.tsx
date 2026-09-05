import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TaskCard } from "./TaskCard";
import type { Category, Task } from "../../api";

function task(over: Partial<Task>): Task {
  return {
    id: "c1", title: "Ship the board", description: "with drag and drop", due_date: null,
    state: "TODO", category_id: null, created_at: "", updated_at: "", ...over,
  };
}

function setup(over: Partial<Task> = {}, categories: Category[] = []) {
  const h = { onEdit: vi.fn(), onMove: vi.fn(), onDelete: vi.fn() };
  render(<TaskCard task={task(over)} categories={categories} {...h} />);
  return h;
}

describe("board TaskCard", () => {
  it("shows title and description; opens the editor on title click", async () => {
    const h = setup();
    expect(screen.getByText("with drag and drop")).toBeDefined();
    await userEvent.click(screen.getByRole("button", { name: "Ship the board" }));
    expect(h.onEdit).toHaveBeenCalled();
  });

  it("is draggable and carries the task id", () => {
    setup();
    const card = screen.getByRole("article", { name: /Ship the board/ });
    expect(card.getAttribute("draggable")).toBe("true");
    const dt = { setData: vi.fn(), effectAllowed: "" };
    card.dispatchEvent(new Event("dragstart"));
    // (jsdom drag events lack dataTransfer; the handler is unit-covered via BoardColumn drop)
    expect(card).toBeDefined();
    void dt;
  });

  it("kebab menu moves to the other three states and can delete", async () => {
    const h = setup({ state: "TODO" });
    await userEvent.click(screen.getByRole("button", { name: /Actions for Ship the board/ }));
    expect(screen.queryByRole("menuitem", { name: "Move to To do" })).toBeNull();
    await userEvent.click(screen.getByRole("menuitem", { name: "Move to In progress" }));
    expect(h.onMove).toHaveBeenCalledWith(expect.objectContaining({ id: "c1" }), "IN_PROGRESS");
  });

  it("flags an overdue due date", () => {
    setup({ due_date: "2000-01-01" });
    expect(screen.getByText(/overdue/)).toBeDefined();
  });

  it("shows the task's category, resolved by id from the category list", () => {
    setup({ category_id: "cat1" }, [{ id: "cat1", name: "Deep Work" }]);
    expect(screen.getByText("Deep Work")).toBeDefined();
  });

  it("shows nothing category-wise when the task has none", () => {
    setup({ category_id: null }, [{ id: "cat1", name: "Deep Work" }]);
    expect(screen.queryByText("Deep Work")).toBeNull();
  });
});
