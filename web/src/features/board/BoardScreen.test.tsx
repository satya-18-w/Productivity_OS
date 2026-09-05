import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { BoardScreen } from "./BoardScreen";
import { api, type Board, type Task } from "../../api";
import { renderShell } from "../../test/renderShell";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return {
    ...actual,
    api: { ...actual.api, board: vi.fn(), listCategories: vi.fn(), moveTask: vi.fn(), deleteTask: vi.fn() },
  };
});

function task(id: string, over: Partial<Task> = {}): Task {
  return {
    id,
    title: `Task ${id}`,
    description: "",
    due_date: null,
    state: "TODO",
    category_id: null,
    created_at: "",
    updated_at: "",
    ...over,
  };
}

const BOARD: Board = {
  columns: [
    { state: "BACKLOG", tasks: [task("b1", { state: "BACKLOG", title: "Idea" })] },
    { state: "TODO", tasks: [task("t1", { title: "Do this" })] },
    { state: "IN_PROGRESS", tasks: [] },
    { state: "DONE", tasks: [task("d1", { state: "DONE", title: "Shipped" })] },
  ],
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.board).mockResolvedValue(BOARD);
  vi.mocked(api.listCategories).mockResolvedValue([]);
  vi.mocked(api.moveTask).mockResolvedValue(undefined);
  vi.mocked(api.deleteTask).mockResolvedValue(undefined);
});

describe("BoardScreen", () => {
  it("renders the four fixed columns in order", async () => {
    renderShell(<BoardScreen />, { route: "/board" });
    await waitFor(() => expect(api.board).toHaveBeenCalled());
    const regions = screen.getAllByRole("region").map((r) => r.getAttribute("aria-label"));
    expect(regions).toEqual([
      "Backlog — 1 task",
      "To do — 1 task",
      "In progress — 0 tasks",
      "Done — 1 task",
    ]);
  });

  it("moves a task via the card kebab menu", async () => {
    renderShell(<BoardScreen />, { route: "/board" });
    await waitFor(() => expect(api.board).toHaveBeenCalled());
    const card = screen.getByRole("article", { name: /Do this/ });
    await userEvent.click(within(card).getByRole("button", { name: /Actions for Do this/ }));
    await userEvent.click(screen.getByRole("menuitem", { name: "Move to Done" }));
    expect(api.moveTask).toHaveBeenCalledWith("t1", "DONE");
  });

  it("does not call the API when moving to the current column", async () => {
    renderShell(<BoardScreen />, { route: "/board" });
    await waitFor(() => expect(api.board).toHaveBeenCalled());
    // drop "t1" (already TODO) onto the To do column
    const todo = screen.getByRole("region", { name: /To do/ });
    const dt = { types: ["text/task-id"], getData: () => "t1", dropEffect: "" };
    todo.dispatchEvent(Object.assign(new Event("dragover", { bubbles: true, cancelable: true }), { dataTransfer: dt }));
    todo.dispatchEvent(Object.assign(new Event("drop", { bubbles: true, cancelable: true }), { dataTransfer: dt }));
    expect(api.moveTask).not.toHaveBeenCalled();
  });

  it("opens the add-task dialog", async () => {
    renderShell(<BoardScreen />, { route: "/board" });
    await waitFor(() => expect(api.board).toHaveBeenCalled());
    await userEvent.click(screen.getByRole("button", { name: "Add task" }));
    expect(screen.getByRole("dialog")).toBeDefined();
  });

  it("shows an error state with retry on load failure", async () => {
    vi.mocked(api.board).mockRejectedValueOnce(new Error("x"));
    renderShell(<BoardScreen />, { route: "/board" });
    expect(await screen.findByText("Something went wrong with the board.")).toBeDefined();
  });
});
