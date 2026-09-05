import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HabitsScreen } from "./HabitsScreen";
import { api, type HabitList } from "../../api";
import { todayISO } from "../../components/date/dateUtils";
import { renderShell } from "../../test/renderShell";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return {
    ...actual,
    api: {
      ...actual.api,
      habits: vi.fn(),
      markHabit: vi.fn(),
      unmarkHabit: vi.fn(),
      archiveHabit: vi.fn(),
      unarchiveHabit: vi.fn(),
      createHabit: vi.fn(),
    },
  };
});

const LIST: HabitList = {
  date: "2026-09-04",
  habits: [
    { id: "h1", name: "Workout", current_streak: 12, completed_on_date: true, last_30_days: 20 },
    { id: "h2", name: "Read", current_streak: 4, completed_on_date: false, last_30_days: 10 },
  ],
  archived: [{ id: "h9", name: "Old habit" }],
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.habits).mockResolvedValue(LIST);
  vi.mocked(api.markHabit).mockResolvedValue(undefined);
  vi.mocked(api.unmarkHabit).mockResolvedValue(undefined);
  vi.mocked(api.archiveHabit).mockResolvedValue(undefined);
});

describe("HabitsScreen", () => {
  it("defaults to the Today checklist with V1-safe KPIs", async () => {
    renderShell(<HabitsScreen />, { route: "/habits" });
    await waitFor(() => expect(api.habits).toHaveBeenCalled());
    expect(screen.getByRole("radio", { name: "Today" }).getAttribute("aria-checked")).toBe("true");
    expect(screen.getByText("Completed today")).toBeDefined();
    expect(screen.getByText("1 / 2")).toBeDefined();
    expect(screen.getByText("Best current streak")).toBeDefined();
    // no longest-streak / consistency% cards
    expect(screen.queryByText(/longest/i)).toBeNull();
    expect(screen.queryByText(/consistency/i)).toBeNull();
  });

  it("requests the exact viewed date (not the server's account-today default)", async () => {
    renderShell(<HabitsScreen />, { route: "/habits" });
    // Displayed rows/KPI/toggles are for the browser today — the fetch must
    // name that date, otherwise rows disagree with their own completion flags
    // whenever the account timezone's today differs.
    await waitFor(() => expect(api.habits).toHaveBeenCalledWith(todayISO()));
  });

  it("switches to the week grid (7 habit fetches) via the view control", async () => {
    renderShell(<HabitsScreen />, { route: "/habits" });
    await waitFor(() => expect(api.habits).toHaveBeenCalled());
    vi.mocked(api.habits).mockClear();
    await userEvent.click(screen.getByRole("radio", { name: "This week" }));
    await waitFor(() => expect(api.habits).toHaveBeenCalledTimes(7));
    expect(await screen.findByRole("rowheader", { name: "Workout" })).toBeDefined();
  });

  it("month view shows the sample-data notice (history endpoint pending)", async () => {
    renderShell(<HabitsScreen />, { route: "/habits?view=month" });
    await waitFor(() => expect(api.habits).toHaveBeenCalled());
    expect(await screen.findByText(/Sample data/)).toBeDefined();
  });

  it("all-habits view lists active and archived with unarchive", async () => {
    renderShell(<HabitsScreen />, { route: "/habits?view=all" });
    await waitFor(() => expect(api.habits).toHaveBeenCalled());
    expect(screen.getByText("Old habit")).toBeDefined();
    expect(screen.getByRole("button", { name: "Unarchive" })).toBeDefined();
  });

  it("toggling a habit calls markHabit for today", async () => {
    renderShell(<HabitsScreen />, { route: "/habits" });
    await waitFor(() => expect(api.habits).toHaveBeenCalled());
    await userEvent.click(screen.getByRole("checkbox", { name: /Read —.*not completed/ }));
    expect(api.markHabit).toHaveBeenCalledWith("h2", expect.any(String));
  });

  it("opens the add-habit dialog", async () => {
    renderShell(<HabitsScreen />, { route: "/habits" });
    await waitFor(() => expect(api.habits).toHaveBeenCalled());
    await userEvent.click(screen.getByRole("button", { name: "Add habit" }));
    expect(screen.getByRole("dialog", { name: "Add habit" })).toBeDefined();
  });
});
