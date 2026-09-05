import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { WeekView } from "./WeekView";
import { api, type DayTimeline, type RangeTimeline } from "../../api";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return { ...actual, api: { ...actual.api, timelineRange: vi.fn() } };
});

const WEEK_DATES = ["2026-08-31", "2026-09-01", "2026-09-02", "2026-09-03", "2026-09-04", "2026-09-05", "2026-09-06"];

function empty(date: string): DayTimeline {
  return { date, planned: [], actual: [] };
}

function range(overrides: Record<string, DayTimeline> = {}): RangeTimeline {
  return {
    from: WEEK_DATES[0],
    to: WEEK_DATES[6],
    days: WEEK_DATES.map((d) => overrides[d] ?? empty(d)),
  };
}

function pblock(over: Partial<DayTimeline["planned"][number]>) {
  return {
    id: "b1",
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
    local_date: "2026-09-02",
    local_start: "09:00",
    local_end: "10:00",
    ends_next_day: false,
    ...over,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("WeekView", () => {
  it("fetches the ISO week containing `date` in one range call, Monday-first", async () => {
    vi.mocked(api.timelineRange).mockResolvedValue(range());
    render(<WeekView date="2026-09-03" onPick={vi.fn()} onJumpToDay={vi.fn()} />);
    await waitFor(() => expect(api.timelineRange).toHaveBeenCalledWith("2026-08-31", "2026-09-06"));
    expect(api.timelineRange).toHaveBeenCalledTimes(1);
    expect(screen.getByText("Mon")).toBeDefined();
    expect(screen.getByText("Sun")).toBeDefined();
  });

  it("places a block chip under its own day, dashed for planned / solid for actual", async () => {
    vi.mocked(api.timelineRange).mockResolvedValue(
      range({
        "2026-09-02": {
          date: "2026-09-02",
          planned: [pblock({ id: "p1" })],
          actual: [pblock({ id: "a1", kind: "actual" })],
        },
      }),
    );
    const { container } = render(<WeekView date="2026-09-02" onPick={vi.fn()} onJumpToDay={vi.fn()} />);
    await waitFor(() => expect(screen.getAllByText("Deep Work")).toHaveLength(2));
    const planned = container.querySelector(".tl-week__chip--planned");
    expect(planned?.textContent).toContain("09:00");
    const solid = [...container.querySelectorAll(".tl-week__chip")].find(
      (el) => !el.className.includes("--planned"),
    );
    expect(solid).toBeDefined();
  });

  it("shows an empty-day placeholder for a day with no blocks", async () => {
    vi.mocked(api.timelineRange).mockResolvedValue(range());
    render(<WeekView date="2026-09-03" onPick={vi.fn()} onJumpToDay={vi.fn()} />);
    await waitFor(() => expect(screen.getAllByText("—").length).toBe(7));
  });

  it("clicking a block chip calls onPick with that block", async () => {
    const onPick = vi.fn();
    vi.mocked(api.timelineRange).mockResolvedValue(
      range({ "2026-09-02": { date: "2026-09-02", planned: [pblock({ id: "p1" })], actual: [] } }),
    );
    render(<WeekView date="2026-09-02" onPick={onPick} onJumpToDay={vi.fn()} />);
    const chip = await screen.findByText("Deep Work");
    await userEvent.click(chip);
    expect(onPick).toHaveBeenCalledWith(expect.objectContaining({ id: "p1" }));
  });

  it("clicking a day column's header calls onJumpToDay with that date", async () => {
    const onJumpToDay = vi.fn();
    vi.mocked(api.timelineRange).mockResolvedValue(range());
    render(<WeekView date="2026-09-03" onPick={vi.fn()} onJumpToDay={onJumpToDay} />);
    await waitFor(() => expect(api.timelineRange).toHaveBeenCalled());
    await userEvent.click(screen.getByText("Wed").closest("button")!);
    expect(onJumpToDay).toHaveBeenCalledWith("2026-09-02");
  });
});
