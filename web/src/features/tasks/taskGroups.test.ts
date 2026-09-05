import { describe, it, expect } from "vitest";
import type { Task, TaskState } from "../../api";
import { groupTasks, taskStats, isOverdue, categoryNameFor } from "./taskGroups";

const TODAY = "2026-09-04"; // a Friday

function t(over: Partial<Task> & { id: string }): Task {
  return {
    title: `Task ${over.id}`,
    description: "",
    due_date: null,
    state: "TODO",
    category_id: null,
    created_at: "2026-09-01T00:00:00Z",
    updated_at: "2026-09-01T00:00:00Z",
    ...over,
  };
}

const TASKS: Task[] = [
  t({ id: "a", due_date: "2026-09-02", state: "TODO" }), // overdue
  t({ id: "b", due_date: "2026-09-04", state: "IN_PROGRESS" }), // today
  t({ id: "c", due_date: "2026-09-08", state: "BACKLOG" }), // upcoming (next week)
  t({ id: "d", due_date: null, state: "TODO" }), // no date
  t({ id: "e", due_date: "2026-09-01", state: "DONE" }), // completed (was overdue)
  t({ id: "f", due_date: "2026-09-05", state: "TODO" }), // this week, upcoming
];

describe("isOverdue", () => {
  it("is true only for non-done tasks past their due date", () => {
    expect(isOverdue(TASKS[0], TODAY)).toBe(true);
    expect(isOverdue(TASKS[4], TODAY)).toBe(false); // DONE
    expect(isOverdue(TASKS[1], TODAY)).toBe(false); // due today
  });
});

describe("groupTasks", () => {
  it("buckets by due date and state, omitting empty groups", () => {
    const g = groupTasks(TASKS, TODAY, "all").map((x) => [x.key, x.tasks.map((tt) => tt.id)]);
    expect(g).toEqual([
      ["overdue", ["a"]],
      ["today", ["b"]],
      ["upcoming", ["f", "c"]], // sorted by due date asc
      ["no_date", ["d"]],
      ["completed", ["e"]],
    ]);
  });

  it("respects the filter tab", () => {
    expect(groupTasks(TASKS, TODAY, "overdue").map((x) => x.key)).toEqual(["overdue"]);
    expect(groupTasks(TASKS, TODAY, "completed").flatMap((x) => x.tasks.map((tt) => tt.id))).toEqual(["e"]);
  });

  it("assigns the right accent tones", () => {
    const byKey = Object.fromEntries(groupTasks(TASKS, TODAY, "all").map((x) => [x.key, x.tone]));
    expect(byKey.overdue).toBe("danger");
    expect(byKey.today).toBe("success");
    expect(byKey.completed).toBe("success");
    expect(byKey.upcoming).toBe("neutral");
  });
});

describe("taskStats", () => {
  it("computes the KPI figures", () => {
    const s = taskStats(TASKS, TODAY);
    expect(s.total).toBe(6);
    expect(s.completed).toBe(1);
    expect(s.inProgress).toBe(1);
    expect(s.overdue).toBe(1);
    // ISO week of Fri 2026-09-04 is Mon 08-31 .. Sun 09-06 → a(02), b(04), f(05) due this week, not done
    // (an overdue task can also be "due this week")
    expect(s.dueThisWeek).toBe(3);
    expect(s.byState).toEqual({ BACKLOG: 1, TODO: 3, IN_PROGRESS: 1, DONE: 1 } satisfies Record<TaskState, number>);
  });
});

describe("categoryNameFor", () => {
  const categories = [{ id: "cat1", name: "Deep Work" }];

  it("resolves a category id to its name", () => {
    expect(categoryNameFor("cat1", categories)).toBe("Deep Work");
  });

  it("returns null for no category", () => {
    expect(categoryNameFor(null, categories)).toBeNull();
  });

  it("returns null for a category id no longer in the list (e.g. archived)", () => {
    expect(categoryNameFor("gone", categories)).toBeNull();
  });
});
