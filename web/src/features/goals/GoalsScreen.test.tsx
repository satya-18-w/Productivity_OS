import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { GoalsScreen } from "./GoalsScreen";
import { api, type Goal } from "../../api";
import { renderShell } from "../../test/renderShell";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return { ...actual, api: { ...actual.api, goals: vi.fn(), setGoalProgress: vi.fn(), deleteGoal: vi.fn() } };
});

function g(over: Partial<Goal> & { id: string; title: string }): Goal {
  return { description: "", target_date: null, progress: "NOT_STARTED", created_at: "2026-01-01", updated_at: "", ...over };
}

const GOALS: Goal[] = [
  g({ id: "g1", title: "Ship V1", progress: "IN_PROGRESS", created_at: "2026-09-03" }),
  g({ id: "g2", title: "Read more", progress: "ACHIEVED", created_at: "2026-09-02" }),
  g({ id: "g3", title: "Learn piano", progress: "NOT_STARTED", created_at: "2026-09-01" }),
];

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.goals).mockResolvedValue(GOALS);
  vi.mocked(api.setGoalProgress).mockResolvedValue(undefined);
  vi.mocked(api.deleteGoal).mockResolvedValue(undefined);
});

describe("GoalsScreen", () => {
  it("loads goals and shows a flat list with KPI counts (no %)", async () => {
    const { container } = renderShell(<GoalsScreen />, { route: "/goals" });
    await waitFor(() => expect(api.goals).toHaveBeenCalled());
    expect(screen.getByRole("button", { name: "Ship V1" })).toBeDefined();
    const kpi = (label: string) =>
      [...container.querySelectorAll(".ui-stat")]
        .find((c) => c.querySelector(".ui-stat__label")?.textContent === label)
        ?.querySelector(".ui-stat__value")?.textContent;
    expect(kpi("Total")).toBe("3");
    expect(kpi("In progress")).toBe("1");
    expect(kpi("Achieved")).toBe("1");
    expect(container.textContent).not.toMatch(/%/);
    expect(container.querySelector(".ui-progress")).toBeNull();
  });

  it("filters by progress state via the tab + URL", async () => {
    renderShell(<GoalsScreen />, { route: "/goals?filter=ACHIEVED" });
    await waitFor(() => expect(api.goals).toHaveBeenCalled());
    expect(screen.getByRole("radio", { name: "Achieved" }).getAttribute("aria-checked")).toBe("true");
    expect(screen.getByRole("button", { name: "Read more" })).toBeDefined();
    expect(screen.queryByRole("button", { name: "Ship V1" })).toBeNull();
  });

  it("offers only All + the four V1 states in the filter (no categories)", async () => {
    renderShell(<GoalsScreen />, { route: "/goals" });
    await waitFor(() => expect(api.goals).toHaveBeenCalled());
    const radios = screen.getAllByRole("radio");
    expect(radios.map((r) => r.textContent)).toEqual([
      "All",
      "Not started",
      "In progress",
      "Achieved",
      "Abandoned",
    ]);
    expect(screen.queryByRole("radio", { name: "Personal" })).toBeNull();
    expect(screen.queryByRole("radio", { name: "Work" })).toBeNull();
  });

  it("sets a goal's progress from the row menu", async () => {
    renderShell(<GoalsScreen />, { route: "/goals" });
    await waitFor(() => expect(api.goals).toHaveBeenCalled());
    await userEvent.click(screen.getByRole("button", { name: /Actions for Ship V1/ }));
    await userEvent.click(screen.getByRole("menuitem", { name: "Set to Achieved" }));
    expect(api.setGoalProgress).toHaveBeenCalledWith("g1", "ACHIEVED");
  });

  it("opens the new-goal dialog", async () => {
    renderShell(<GoalsScreen />, { route: "/goals" });
    await waitFor(() => expect(api.goals).toHaveBeenCalled());
    await userEvent.click(screen.getAllByRole("button", { name: "New goal" })[0]);
    expect(screen.getByRole("dialog", { name: "New goal" })).toBeDefined();
  });

  it("shows an empty state when there are no goals", async () => {
    vi.mocked(api.goals).mockResolvedValueOnce([]);
    renderShell(<GoalsScreen />, { route: "/goals" });
    expect(await screen.findByText("No goals yet")).toBeDefined();
  });

  it("shows an error state with retry on load failure", async () => {
    vi.mocked(api.goals).mockRejectedValueOnce(new Error("x"));
    renderShell(<GoalsScreen />, { route: "/goals" });
    expect(await screen.findByText("Could not load your goals.")).toBeDefined();
  });
});
