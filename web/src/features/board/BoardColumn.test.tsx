import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { BoardColumn } from "./BoardColumn";
import type { Task } from "../../api";

const noop = vi.fn();
const handlers = { onEdit: noop, onMove: noop, onDelete: noop, categories: [] };

function task(id: string): Task {
  return {
    id,
    title: `Task ${id}`,
    description: "",
    due_date: null,
    state: "TODO",
    category_id: null,
    created_at: "",
    updated_at: "",
  };
}

function fireDrop(el: Element, taskId: string) {
  const dataTransfer = {
    types: ["text/task-id"],
    getData: (t: string) => (t === "text/task-id" ? taskId : ""),
    dropEffect: "",
  };
  el.dispatchEvent(Object.assign(new Event("dragover", { bubbles: true, cancelable: true }), { dataTransfer }));
  el.dispatchEvent(Object.assign(new Event("drop", { bubbles: true, cancelable: true }), { dataTransfer }));
}

describe("BoardColumn", () => {
  it("labels the column region with its state and count", () => {
    render(<BoardColumn state="IN_PROGRESS" tasks={[task("a"), task("b")]} onDropTask={noop} {...handlers} />);
    expect(screen.getByRole("region", { name: "In progress — 2 tasks" })).toBeDefined();
    expect(screen.getByText("2")).toBeDefined();
  });

  it("shows an empty hint when there are no tasks", () => {
    render(<BoardColumn state="DONE" tasks={[]} onDropTask={noop} {...handlers} />);
    expect(screen.getByText("No tasks")).toBeDefined();
  });

  it("calls onDropTask with the dropped id and this column's state", () => {
    const onDropTask = vi.fn();
    render(<BoardColumn state="DONE" tasks={[]} onDropTask={onDropTask} {...handlers} />);
    fireDrop(screen.getByRole("region", { name: /Done/ }), "task-42");
    expect(onDropTask).toHaveBeenCalledWith("task-42", "DONE");
  });
});
