import { describe, it, expect, vi, beforeEach } from "vitest";
import { api, type HabitList, type HabitView } from "../../api";
import { fetchWeek, mockHabitHistory, trailingDays } from "./habitData";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return { ...actual, api: { ...actual.api, habits: vi.fn() } };
});

function hv(over: Partial<HabitView> & { id: string }): HabitView {
  return { name: `Habit ${over.id}`, current_streak: 3, completed_on_date: false, last_30_days: 12, ...over };
}

beforeEach(() => vi.clearAllMocks());

describe("fetchWeek", () => {
  it("builds a Monday-first ISO week and a per-habit per-day completion map", async () => {
    // 2026-09-04 is a Friday → week is Mon 2026-08-31 .. Sun 2026-09-06
    vi.mocked(api.habits).mockImplementation(async (date?: string) => {
      const completed = date === "2026-09-02" || date === "2026-09-04";
      return {
        date: date ?? "2026-09-04",
        habits: [hv({ id: "h1", completed_on_date: completed })],
        archived: [],
      } as HabitList;
    });

    const wd = await fetchWeek("2026-09-04");
    expect(wd.weekStart).toBe("2026-08-31");
    expect(wd.days).toEqual([
      "2026-08-31", "2026-09-01", "2026-09-02", "2026-09-03", "2026-09-04", "2026-09-05", "2026-09-06",
    ]);
    expect(api.habits).toHaveBeenCalledTimes(7);
    expect(wd.completion.h1["2026-09-02"]).toBe(true);
    expect(wd.completion.h1["2026-09-03"]).toBe(false);
  });
});

describe("mockHabitHistory", () => {
  it("is deterministic per habit id", () => {
    const h = hv({ id: "abc", last_30_days: 20 });
    const days = trailingDays(30, "2026-09-04");
    expect([...mockHabitHistory(h, days)]).toEqual([...mockHabitHistory(h, days)]);
  });

  it("scales roughly with last_30_days density", () => {
    const days = trailingDays(30, "2026-09-04");
    const sparse = mockHabitHistory(hv({ id: "x", last_30_days: 3 }), days).size;
    const dense = mockHabitHistory(hv({ id: "x", last_30_days: 28 }), days).size;
    expect(dense).toBeGreaterThan(sparse);
  });
});

describe("trailingDays", () => {
  it("returns n days ending on the given date, oldest first", () => {
    expect(trailingDays(3, "2026-09-04")).toEqual(["2026-09-02", "2026-09-03", "2026-09-04"]);
  });
});
