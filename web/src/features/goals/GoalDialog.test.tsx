import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { GoalDialog } from "./GoalDialog";
import { api, ApiError, type Goal } from "../../api";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return { ...actual, api: { ...actual.api, createGoal: vi.fn(), updateGoal: vi.fn() } };
});

const EXISTING: Goal = {
  id: "g9", title: "Learn Go", description: "for the backend", target_date: "2026-12-01",
  progress: "IN_PROGRESS", created_at: "", updated_at: "",
};

beforeEach(() => vi.clearAllMocks());

describe("GoalDialog", () => {
  it("creates a goal with title / description / target date only", async () => {
    vi.mocked(api.createGoal).mockResolvedValue({} as never);
    render(<GoalDialog open target={{ mode: "new" }} onClose={() => {}} onSaved={vi.fn()} />);
    // no progress / % / linked-tasks field
    expect(screen.queryByLabelText(/progress|status|%|tasks/i)).toBeNull();
    await userEvent.type(screen.getByLabelText(/Title/), "Read 12 books");
    await userEvent.click(screen.getByRole("button", { name: "Create" }));
    expect(api.createGoal).toHaveBeenCalledWith({ title: "Read 12 books", description: "", target_date: null });
  });

  it("pre-fills and updates on edit", async () => {
    vi.mocked(api.updateGoal).mockResolvedValue(undefined);
    render(<GoalDialog open target={{ mode: "edit", goal: EXISTING }} onClose={() => {}} onSaved={vi.fn()} />);
    expect((screen.getByLabelText(/Title/) as HTMLInputElement).value).toBe("Learn Go");
    expect((screen.getByLabelText(/Target date/) as HTMLInputElement).value).toBe("2026-12-01");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(api.updateGoal).toHaveBeenCalledWith("g9", expect.objectContaining({ title: "Learn Go" }));
  });

  it("surfaces a field validation error", async () => {
    vi.mocked(api.createGoal).mockRejectedValue(
      new ApiError(400, "VALIDATION_ERROR", "bad", { title: "Title is required." }),
    );
    render(<GoalDialog open target={{ mode: "new" }} onClose={() => {}} onSaved={vi.fn()} />);
    await userEvent.type(screen.getByLabelText(/Title/), "x");
    await userEvent.click(screen.getByRole("button", { name: "Create" }));
    expect(await screen.findByText("Title is required.")).toBeDefined();
  });
});
