import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TasksScreen } from "./TasksScreen";
import { api, type Board, type Task } from "../../api";
import { renderShell } from "../../test/renderShell";
import { todayISO, shiftDays } from "../../components/date/dateUtils";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return {
    ...actual,
    api: { ...actual.api, board: vi.fn(), listCategories: vi.fn(), moveTask: vi.fn(), deleteTask: vi.fn() },
  };
});

const TODAY = todayISO();

function task(over: Partial<Task> & { id: string; title: string }): Task {
  return {
    description: "",
    due_date: null,
    state: "TODO",
    category_id: null,
    created_at: "2026-01-01",
    updated_at: "",
    ...over,
  };
}

const BOARD: Board = {
  columns: [
    { state: "BACKLOG", tasks: [task({ id: "b1", title: "Someday thing" })] },
    { state: "TODO", tasks: [task({ id: "o1", title: "Overdue thing", due_date: shiftDays(TODAY, -2) })] },
    { state: "IN_PROGRESS", tasks: [task({ id: "t1", title: "Today thing", due_date: TODAY, state: "IN_PROGRESS" })] },
    { state: "DONE", tasks: [task({ id: "d1", title: "Done thing", state: "DONE" })] },
  ],
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.board).mockResolvedValue(BOARD);
  vi.mocked(api.listCategories).mockResolvedValue([]);
  vi.mocked(api.moveTask).mockResolvedValue(undefined);
  vi.mocked(api.deleteTask).mockResolvedValue(undefined);
});

describe("TasksScreen", () => {
  it("loads the board, flattens it, and groups by due date", async () => {
    const { container } = renderShell(<TasksScreen />, { route: "/tasks" });
    await waitFor(() => expect(api.board).toHaveBeenCalled());
    const groups = [...container.querySelectorAll(".ui-row-group")].map((g) => g.textContent);
    expect(groups).toEqual(["Overdue(1)", "Today(1)", "No due date(1)", "Completed(1)"]);
    expect(screen.getByRole("button", { name: "Overdue thing" })).toBeDefined();
    expect(screen.getByRole("button", { name: "Done thing" })).toBeDefined();
  });

  it("shows the KPI figures", async () => {
    const { container } = renderShell(<TasksScreen />, { route: "/tasks" });
    await waitFor(() => expect(api.board).toHaveBeenCalled());
    const kpi = (label: string) =>
      [...container.querySelectorAll(".ui-stat")]
        .find((c) => c.querySelector(".ui-stat__label")?.textContent === label)
        ?.querySelector(".ui-stat__value")?.textContent;
    expect(kpi("Total")).toBe("4");
    expect(kpi("In progress")).toBe("1");
    expect(kpi("Overdue")).toBe("1");
    expect(kpi("Due this week")).toBeDefined();
    expect(screen.getByText("1 completed")).toBeDefined(); // sublabel on Total
  });

  it("filters via the tab and the URL", async () => {
    renderShell(<TasksScreen />, { route: "/tasks?filter=completed" });
    await waitFor(() => expect(api.board).toHaveBeenCalled());
    expect(screen.getByRole("radio", { name: "Completed" }).getAttribute("aria-checked")).toBe("true");
    expect(screen.getByRole("button", { name: "Done thing" })).toBeDefined();
    expect(screen.queryByRole("button", { name: "Overdue thing" })).toBeNull();
  });

  it("toggling a checkbox moves the task to DONE / TODO", async () => {
    renderShell(<TasksScreen />, { route: "/tasks" });
    await waitFor(() => expect(api.board).toHaveBeenCalled());
    await userEvent.click(screen.getByRole("checkbox", { name: /Mark "Today thing" done/ }));
    expect(api.moveTask).toHaveBeenCalledWith("t1", "DONE");
  });

  it("unchecking a done task restores its previous state, not always TODO", async () => {
    const DONE_BOARD: Board = {
      columns: [
        { state: "BACKLOG", tasks: [] },
        { state: "TODO", tasks: [] },
        { state: "IN_PROGRESS", tasks: [] },
        {
          state: "DONE",
          tasks: [task({ id: "t1", title: "Today thing", due_date: TODAY, state: "DONE" })],
        },
      ],
    };
    vi.mocked(api.board).mockReset();
    vi.mocked(api.board).mockResolvedValueOnce(BOARD).mockResolvedValue(DONE_BOARD);
    renderShell(<TasksScreen />, { route: "/tasks" });
    await waitFor(() => expect(api.board).toHaveBeenCalled());
    await userEvent.click(screen.getByRole("checkbox", { name: /Mark "Today thing" done/ }));
    await waitFor(() => expect(api.moveTask).toHaveBeenCalledWith("t1", "DONE"));
    await waitFor(() =>
      screen.getByRole("checkbox", { name: /Mark "Today thing" not done/ }),
    );
    await userEvent.click(screen.getByRole("checkbox", { name: /Mark "Today thing" not done/ }));
    await waitFor(() => expect(api.moveTask).toHaveBeenCalledWith("t1", "IN_PROGRESS"));
  });

  it("unchecking a task with no known previous state falls back to TODO", async () => {
    renderShell(<TasksScreen />, { route: "/tasks" });
    await waitFor(() => expect(api.board).toHaveBeenCalled());
    await userEvent.click(screen.getByRole("checkbox", { name: /Mark "Done thing" not done/ }));
    await waitFor(() => expect(api.moveTask).toHaveBeenCalledWith("d1", "TODO"));
  });

  it("opens the add-task dialog", async () => {
    renderShell(<TasksScreen />, { route: "/tasks" });
    await waitFor(() => expect(api.board).toHaveBeenCalled());
    await userEvent.click(screen.getAllByRole("button", { name: "Add task" })[0]);
    expect(screen.getByRole("dialog")).toBeDefined();
  });

  it("shows an error state with retry on load failure", async () => {
    vi.mocked(api.board).mockRejectedValueOnce(new Error("x"));
    renderShell(<TasksScreen />, { route: "/tasks" });
    expect(await screen.findByText("Could not load your tasks.")).toBeDefined();
  });

  it("shows a task's category, resolved from the account's category list", async () => {
    vi.mocked(api.listCategories).mockResolvedValue([{ id: "cat1", name: "Deep Work" }]);
    vi.mocked(api.board).mockResolvedValue({
      columns: [
        { state: "BACKLOG", tasks: [] },
        { state: "TODO", tasks: [task({ id: "t1", title: "Ship it", category_id: "cat1" })] },
        { state: "IN_PROGRESS", tasks: [] },
        { state: "DONE", tasks: [] },
      ],
    });
    renderShell(<TasksScreen />, { route: "/tasks" });
    expect(await screen.findByText("Deep Work")).toBeDefined();
  });
});
