import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { TodayTasks } from "./TodayTasks";
import { api, type Board } from "../../api";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return { ...actual, api: { ...actual.api, board: vi.fn() } };
});

function board(due: string | null, state: Board["columns"][number]["tasks"][number]["state"] = "TODO"): Board {
  return {
    columns: [
      {
        state: "BACKLOG",
        tasks: [
          { id: "t1", title: "Due today", description: "", due_date: due, state, category_id: null, created_at: "", updated_at: "" },
          { id: "t2", title: "Due another day", description: "", due_date: "2026-01-01", state: "TODO", category_id: null, created_at: "", updated_at: "" },
        ],
      },
      { state: "TODO", tasks: [] },
      { state: "IN_PROGRESS", tasks: [] },
      { state: "DONE", tasks: [] },
    ],
  };
}

function renderWidget(date: string) {
  return render(
    <MemoryRouter>
      <TodayTasks date={date} />
    </MemoryRouter>,
  );
}

describe("TodayTasks", () => {
  beforeEach(() => vi.clearAllMocks());

  it("lists only tasks due on the viewed date", async () => {
    vi.mocked(api.board).mockResolvedValue(board("2026-09-04"));
    renderWidget("2026-09-04");
    expect(await screen.findByText("Due today")).toBeDefined();
    expect(screen.queryByText("Due another day")).toBeNull();
    expect(screen.getByText("0/1")).toBeDefined();
  });

  it("counts done tasks and links to /tasks", async () => {
    vi.mocked(api.board).mockResolvedValue(board("2026-09-04", "DONE"));
    renderWidget("2026-09-04");
    await screen.findByText("Due today");
    expect(screen.getByText("1/1")).toBeDefined();
    expect(screen.getByRole("link", { name: /View all tasks/ }).getAttribute("href")).toBe("/tasks");
  });

  it("hides itself when nothing is due or the load fails", async () => {
    vi.mocked(api.board).mockResolvedValue(board(null));
    const { container } = renderWidget("2026-09-04");
    await waitFor(() => expect(api.board).toHaveBeenCalled());
    expect(container.firstChild).toBeNull();

    vi.mocked(api.board).mockRejectedValueOnce(new Error("boom"));
    const r2 = renderWidget("2026-09-04");
    await waitFor(() => expect(api.board).toHaveBeenCalledTimes(2));
    expect(r2.container.firstChild).toBeNull();
  });
});
