import { describe, it, expect } from "vitest";
import type { Goal } from "../../api";
import { filterGoals, goalStats } from "./goalHelpers";

function g(over: Partial<Goal> & { id: string }): Goal {
  return {
    title: `Goal ${over.id}`,
    description: "",
    target_date: null,
    progress: "NOT_STARTED",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "",
    ...over,
  };
}

const GOALS: Goal[] = [
  g({ id: "a", progress: "IN_PROGRESS", created_at: "2026-09-01" }),
  g({ id: "b", progress: "ACHIEVED", created_at: "2026-09-03" }),
  g({ id: "c", progress: "IN_PROGRESS", created_at: "2026-09-02" }),
  g({ id: "d", progress: "ABANDONED", created_at: "2026-08-20" }),
];

describe("filterGoals", () => {
  it("returns all goals newest-first when unfiltered", () => {
    expect(filterGoals(GOALS, "all").map((x) => x.id)).toEqual(["b", "c", "a", "d"]);
  });

  it("narrows to a single progress state", () => {
    expect(filterGoals(GOALS, "IN_PROGRESS").map((x) => x.id)).toEqual(["c", "a"]);
    expect(filterGoals(GOALS, "ACHIEVED").map((x) => x.id)).toEqual(["b"]);
  });
});

describe("goalStats", () => {
  it("counts totals per progress state", () => {
    const s = goalStats(GOALS);
    expect(s.total).toBe(4);
    expect(s.byProgress).toEqual({ NOT_STARTED: 0, IN_PROGRESS: 2, ACHIEVED: 1, ABANDONED: 1 });
  });
});
