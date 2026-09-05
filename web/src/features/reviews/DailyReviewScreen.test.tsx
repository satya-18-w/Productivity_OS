import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DailyReviewScreen } from "./DailyReviewScreen";
import { api, type DayComparison, type HabitList } from "../../api";
import { renderShell } from "../../test/renderShell";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return { ...actual, api: { ...actual.api, comparison: vi.fn(), habits: vi.fn() } };
});

const COMPARISON: DayComparison = {
  date: "2026-08-15",
  categories: [
    { category_id: "c1", category_name: "Deep Work", planned_seconds: 0, actual_seconds: 3600, difference_seconds: 0 },
    { category_id: "c2", category_name: "Admin", planned_seconds: 0, actual_seconds: 0, difference_seconds: 0 },
  ],
};

const HABITS: HabitList = {
  date: "2026-08-15",
  habits: [
    { id: "h1", name: "Workout", current_streak: 3, completed_on_date: true, last_30_days: 10 },
    { id: "h2", name: "Read", current_streak: 0, completed_on_date: false, last_30_days: 5 },
  ],
  archived: [],
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.comparison).mockResolvedValue(COMPARISON);
  vi.mocked(api.habits).mockResolvedValue(HABITS);
});

describe("DailyReviewScreen", () => {
  it("shows the four fixed prompts and the reference totals, filtering zero-time categories", async () => {
    renderShell(<DailyReviewScreen />, { route: "/reviews/daily?date=2026-08-15" });
    await waitFor(() => expect(api.comparison).toHaveBeenCalledWith("2026-08-15"));

    expect(screen.getByText("What went well today?")).toBeDefined();
    expect(screen.getByText("What didn't go as planned?")).toBeDefined();
    expect(screen.getByText("What will you do differently tomorrow?")).toBeDefined();
    expect(screen.getByText("One thing you're grateful for.")).toBeDefined();

    expect(screen.getByText(/Deep Work/)).toBeDefined();
    expect(screen.queryByText(/Admin/)).toBeNull(); // 0 actual seconds — filtered out

    expect(screen.getByText("Workout")).toBeDefined();
    expect(screen.getByText("Read")).toBeDefined();
  });

  it("starts blank with a 'Save review' button when nothing is saved yet", async () => {
    renderShell(<DailyReviewScreen />, { route: "/reviews/daily?date=2026-08-15" });
    await waitFor(() => expect(api.comparison).toHaveBeenCalled());
    expect((screen.getByLabelText("What went well today?") as HTMLTextAreaElement).value).toBe("");
    expect(screen.getByRole("button", { name: "Save review" })).toBeDefined();
  });

  it("saves an answer and re-shows it after navigating away and back", async () => {
    renderShell(<DailyReviewScreen />, { route: "/reviews/daily?date=2026-08-20" });
    await waitFor(() => expect(api.comparison).toHaveBeenCalled());

    await userEvent.type(screen.getByLabelText("What went well today?"), "Shipped the review screen");
    await userEvent.click(screen.getByRole("button", { name: "Save review" }));
    await waitFor(() => expect(screen.getByText("Saved")).toBeDefined());
    expect(screen.getByRole("button", { name: "Save changes" })).toBeDefined();
  });

  it("navigating the date stepper loads that date's own review state", async () => {
    renderShell(<DailyReviewScreen />, { route: "/reviews/daily?date=2026-09-11" });
    await waitFor(() => expect(api.comparison).toHaveBeenCalledWith("2026-09-11"));
    await userEvent.click(screen.getByRole("button", { name: "Previous day" }));
    await waitFor(() => expect(api.comparison).toHaveBeenCalledWith("2026-09-10"));
  });
});
