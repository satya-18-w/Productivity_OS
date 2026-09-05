import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MonthView } from "./MonthView";
import { api, type DayTimeline, type RangeTimeline } from "../../api";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return { ...actual, api: { ...actual.api, timelineRange: vi.fn() } };
});

// September 2026's visible 6-week grid: Mon 2026-08-31 .. Sun 2026-10-11 (42 days).
const GRID_FROM = "2026-08-31";
const GRID_TO = "2026-10-11";

function empty(date: string): DayTimeline {
  return { date, planned: [], actual: [] };
}

function range(overrides: Record<string, DayTimeline> = {}): RangeTimeline {
  const days: DayTimeline[] = [];
  const cursor = new Date(2026, 7, 31); // Aug 31, 2026
  for (let i = 0; i < 42; i++) {
    const iso = cursor.toISOString().slice(0, 10);
    days.push(overrides[iso] ?? empty(iso));
    cursor.setDate(cursor.getDate() + 1);
  }
  return { from: GRID_FROM, to: GRID_TO, days };
}

function pblock(id: string, over: Partial<DayTimeline["planned"][number]> = {}) {
  return {
    id,
    kind: "planned" as const,
    starts_at: "",
    ends_at: "",
    category_id: "c1",
    category_name: "Deep Work",
    task_id: null,
    start_minute: 540,
    end_minute: 600,
    from_prev_day: false,
    to_next_day: false,
    local_date: "2026-09-20",
    local_start: "09:00",
    local_end: "10:00",
    ends_next_day: false,
    ...over,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("MonthView", () => {
  it("fetches every visible day (incl. adjacent-month overflow) in one range call and renders a 7-column grid", async () => {
    vi.mocked(api.timelineRange).mockResolvedValue(range());
    render(<MonthView date="2026-09-05" onPick={vi.fn()} onJumpToDay={vi.fn()} />);
    await waitFor(() => expect(api.timelineRange).toHaveBeenCalledWith(GRID_FROM, GRID_TO));
    expect(api.timelineRange).toHaveBeenCalledTimes(1);
    expect(screen.getAllByText("Mon").length).toBeGreaterThan(0);
  });

  it("places a block chip under the right day cell", async () => {
    vi.mocked(api.timelineRange).mockResolvedValue(
      range({ "2026-09-20": { date: "2026-09-20", planned: [pblock("p1")], actual: [] } }),
    );
    render(<MonthView date="2026-09-05" onPick={vi.fn()} onJumpToDay={vi.fn()} />);
    await waitFor(() => expect(screen.getByText(/09:00 Deep Work/)).toBeDefined());
  });

  it("shows a '+N more' overflow link beyond the first 3 blocks in a day", async () => {
    vi.mocked(api.timelineRange).mockResolvedValue(
      range({
        "2026-09-20": {
          date: "2026-09-20",
          planned: [
            pblock("p1", { start_minute: 540, end_minute: 600 }),
            pblock("p2", { start_minute: 660, end_minute: 720 }),
            pblock("p3", { start_minute: 780, end_minute: 840 }),
            pblock("p4", { start_minute: 900, end_minute: 960 }),
          ],
          actual: [],
        },
      }),
    );
    render(<MonthView date="2026-09-05" onPick={vi.fn()} onJumpToDay={vi.fn()} />);
    expect(await screen.findByText("+1 more")).toBeDefined();
  });

  it("clicking a block chip calls onPick", async () => {
    const onPick = vi.fn();
    vi.mocked(api.timelineRange).mockResolvedValue(
      range({ "2026-09-20": { date: "2026-09-20", planned: [pblock("p1")], actual: [] } }),
    );
    render(<MonthView date="2026-09-05" onPick={onPick} onJumpToDay={vi.fn()} />);
    const chip = await screen.findByText(/09:00 Deep Work/);
    await userEvent.click(chip);
    expect(onPick).toHaveBeenCalledWith(expect.objectContaining({ id: "p1" }));
  });

  it("clicking a day number calls onJumpToDay with that date", async () => {
    const onJumpToDay = vi.fn();
    vi.mocked(api.timelineRange).mockResolvedValue(range());
    render(<MonthView date="2026-09-05" onPick={vi.fn()} onJumpToDay={onJumpToDay} />);
    await waitFor(() => expect(api.timelineRange).toHaveBeenCalled());
    await userEvent.click(screen.getByRole("button", { name: /Saturday, September 5, 2026/i }));
    expect(onJumpToDay).toHaveBeenCalledWith("2026-09-05");
  });

  it("only refetches when the month identity changes, not on every day-jump within it", async () => {
    vi.mocked(api.timelineRange).mockResolvedValue(range());
    const { rerender } = render(<MonthView date="2026-09-05" onPick={vi.fn()} onJumpToDay={vi.fn()} />);
    await waitFor(() => expect(api.timelineRange).toHaveBeenCalledTimes(1));
    vi.mocked(api.timelineRange).mockClear();
    rerender(<MonthView date="2026-09-12" onPick={vi.fn()} onJumpToDay={vi.fn()} />);
    await new Promise((r) => setTimeout(r, 10));
    expect(api.timelineRange).not.toHaveBeenCalled();
  });
});
