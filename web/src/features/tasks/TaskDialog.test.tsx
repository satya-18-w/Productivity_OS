import type { ReactElement } from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import userEvent from "@testing-library/user-event";
import { TaskDialog } from "./TaskDialog";
import { api, ApiError, type Category, type Task, type TaskLinkedBlock } from "../../api";
import { AuthStub } from "../../test/renderShell";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return {
    ...actual,
    api: { ...actual.api, createTask: vi.fn(), updateTask: vi.fn(), deleteTask: vi.fn(), blocksForTask: vi.fn() },
  };
});

const EXISTING: Task = {
  id: "t9", title: "Old title", description: "notes", due_date: "2026-09-10",
  state: "TODO", category_id: null, created_at: "", updated_at: "",
};

const CATEGORIES: Category[] = [
  { id: "cat1", name: "Deep Work" },
  { id: "cat2", name: "Admin" },
];

function renderDialog(ui: ReactElement) {
  return render(
    <MemoryRouter>
      <AuthStub>{ui}</AuthStub>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.blocksForTask).mockResolvedValue([]);
});

describe("TaskDialog", () => {
  it("creates a task with title / description / due date / category", async () => {
    vi.mocked(api.createTask).mockResolvedValue({} as never);
    const onSaved = vi.fn();
    renderDialog(
      <TaskDialog open target={{ mode: "new" }} categories={CATEGORIES} onClose={() => {}} onSaved={onSaved} />,
    );
    await userEvent.type(screen.getByLabelText(/Title/), "Plan the week");
    await userEvent.click(screen.getByRole("button", { name: "Add" }));
    expect(api.createTask).toHaveBeenCalledWith({
      title: "Plan the week",
      description: "",
      due_date: null,
      category_id: null,
    });
    expect(onSaved).toHaveBeenCalled();
  });

  it("offers a category picker, but no priority/assignee/status field (§7 scope)", () => {
    renderDialog(
      <TaskDialog open target={{ mode: "new" }} categories={CATEGORIES} onClose={() => {}} onSaved={() => {}} />,
    );
    expect(screen.getByLabelText("Category")).toBeDefined();
    expect(screen.getByRole("option", { name: "Deep Work" })).toBeDefined();
    expect(screen.queryByLabelText(/priority/i)).toBeNull();
    expect(screen.queryByLabelText(/assignee/i)).toBeNull();
    expect(screen.queryByLabelText(/status/i)).toBeNull();
  });

  it("creates a task with the chosen category", async () => {
    vi.mocked(api.createTask).mockResolvedValue({} as never);
    renderDialog(
      <TaskDialog open target={{ mode: "new" }} categories={CATEGORIES} onClose={() => {}} onSaved={vi.fn()} />,
    );
    await userEvent.type(screen.getByLabelText(/Title/), "Ship it");
    await userEvent.selectOptions(screen.getByLabelText("Category"), "cat1");
    await userEvent.click(screen.getByRole("button", { name: "Add" }));
    expect(api.createTask).toHaveBeenCalledWith(
      expect.objectContaining({ title: "Ship it", category_id: "cat1" }),
    );
  });

  it("pre-fills and updates when editing, including its category", async () => {
    vi.mocked(api.updateTask).mockResolvedValue(undefined);
    renderDialog(
      <TaskDialog
        open
        target={{ mode: "edit", task: { ...EXISTING, category_id: "cat2" } }}
        categories={CATEGORIES}
        onClose={() => {}}
        onSaved={vi.fn()}
      />,
    );
    expect((screen.getByLabelText(/Title/) as HTMLInputElement).value).toBe("Old title");
    expect((screen.getByLabelText("Category") as HTMLSelectElement).value).toBe("cat2");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(api.updateTask).toHaveBeenCalledWith(
      "t9",
      expect.objectContaining({ title: "Old title", category_id: "cat2" }),
    );
  });

  it("deletes the task", async () => {
    vi.mocked(api.deleteTask).mockResolvedValue(undefined);
    const onSaved = vi.fn();
    renderDialog(
      <TaskDialog
        open
        target={{ mode: "edit", task: EXISTING }}
        categories={CATEGORIES}
        onClose={() => {}}
        onSaved={onSaved}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: "Delete" }));
    expect(api.deleteTask).toHaveBeenCalledWith("t9");
    expect(onSaved).toHaveBeenCalled();
  });

  it("surfaces a field validation error", async () => {
    vi.mocked(api.createTask).mockRejectedValue(
      new ApiError(400, "VALIDATION_ERROR", "bad", { title: "Title is required." }),
    );
    renderDialog(
      <TaskDialog open target={{ mode: "new" }} categories={CATEGORIES} onClose={() => {}} onSaved={vi.fn()} />,
    );
    await userEvent.type(screen.getByLabelText(/Title/), "x");
    await userEvent.click(screen.getByRole("button", { name: "Add" }));
    expect(await screen.findByText("Title is required.")).toBeDefined();
  });

  it("does not show a Scheduled blocks section when creating a new task", () => {
    renderDialog(
      <TaskDialog open target={{ mode: "new" }} categories={CATEGORIES} onClose={() => {}} onSaved={() => {}} />,
    );
    expect(screen.queryByText("Scheduled blocks")).toBeNull();
    expect(api.blocksForTask).not.toHaveBeenCalled();
  });

  it("shows an empty state when editing a task with no linked blocks", async () => {
    renderDialog(
      <TaskDialog open target={{ mode: "edit", task: EXISTING }} categories={CATEGORIES} onClose={() => {}} onSaved={() => {}} />,
    );
    expect(api.blocksForTask).toHaveBeenCalledWith("t9");
    expect(await screen.findByText("No time blocks scheduled yet.")).toBeDefined();
  });

  it("lists a task's scheduled blocks with a link back to the Timeline", async () => {
    const BLOCKS: TaskLinkedBlock[] = [
      {
        id: "b1", kind: "planned",
        starts_at: "2026-09-05T14:00:00Z", ends_at: "2026-09-05T15:00:00Z",
        category_id: "cat1", category_name: "Deep Work", task_id: "t9",
      },
    ];
    vi.mocked(api.blocksForTask).mockResolvedValue(BLOCKS);
    renderDialog(
      <TaskDialog open target={{ mode: "edit", task: EXISTING }} categories={CATEGORIES} onClose={() => {}} onSaved={() => {}} />,
    );
    const link = await screen.findByRole("link", { name: "View on Timeline →" });
    expect(link.getAttribute("href")).toBe("/timeline?date=2026-09-05&openBlock=b1");
    expect(screen.getByText("Planned")).toBeDefined();
    expect(screen.getByText(/14:00–15:00/)).toBeDefined();
  });

  it("shows an error state when the scheduled-blocks fetch fails", async () => {
    vi.mocked(api.blocksForTask).mockRejectedValue(new Error("boom"));
    renderDialog(
      <TaskDialog open target={{ mode: "edit", task: EXISTING }} categories={CATEGORIES} onClose={() => {}} onSaved={() => {}} />,
    );
    expect(await screen.findByText("Could not load scheduled blocks.")).toBeDefined();
  });
});
