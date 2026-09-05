import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TimelineScreen } from "./TimelineScreen";
import { api, type DayTimeline, type PositionedBlock, type RangeTimeline } from "../../api";
import { renderShell } from "../../test/renderShell";
import { todayISO } from "../../components/date/dateUtils";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return {
    ...actual,
    api: {
      ...actual.api,
      timeline: vi.fn(),
      timelineRange: vi.fn(),
      comparison: vi.fn(),
      listCategories: vi.fn(),
      board: vi.fn(),
    },
  };
});

function pblock(over: Partial<PositionedBlock>): PositionedBlock {
  return {
    id: "b1", kind: "planned", starts_at: "", ends_at: "", category_id: "c1",
    category_name: "Deep Work", task_id: null, start_minute: 540, end_minute: 660,
    from_prev_day: false, to_next_day: false, local_date: "2026-09-01",
    local_start: "09:00", local_end: "11:00", ends_next_day: false, ...over,
  };
}

const TL: DayTimeline = {
  date: "",
  planned: [pblock({ id: "p1" })],
  actual: [pblock({ id: "a1", kind: "actual", category_id: "c2", category_name: "Admin", start_minute: 720, end_minute: 780 })],
};

function emptyRange(from: string, to: string): RangeTimeline {
  const days: DayTimeline[] = [];
  const cursor = new Date(`${from}T00:00:00Z`);
  const end = new Date(`${to}T00:00:00Z`);
  while (cursor <= end) {
    days.push({ date: cursor.toISOString().slice(0, 10), planned: [], actual: [] });
    cursor.setUTCDate(cursor.getUTCDate() + 1);
  }
  return { from, to, days };
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.timeline).mockResolvedValue(TL);
  vi.mocked(api.timelineRange).mockImplementation(async (from, to) => emptyRange(from, to));
  vi.mocked(api.comparison).mockResolvedValue({ date: "", categories: [] });
  vi.mocked(api.listCategories).mockResolvedValue([]);
  vi.mocked(api.board).mockResolvedValue({ columns: [] });
});

describe("TimelineScreen", () => {
  it("loads today's timeline and shows the Day view with two lanes", async () => {
    renderShell(<TimelineScreen />, { route: "/timeline" });
    await waitFor(() => expect(api.timeline).toHaveBeenCalledWith(todayISO()));
    expect(screen.getByRole("radio", { name: "Day" }).getAttribute("aria-checked")).toBe("true");
    expect(screen.getByRole("list", { name: "Planned blocks" })).toBeDefined();
    expect(screen.getByRole("list", { name: "Actual blocks" })).toBeDefined();
  });

  it("reads date + view from the URL params", async () => {
    renderShell(<TimelineScreen />, { route: "/timeline?date=2026-09-01&view=agenda" });
    await waitFor(() => expect(api.timeline).toHaveBeenCalledWith("2026-09-01"));
    expect(screen.getByRole("radio", { name: "Agenda" }).getAttribute("aria-checked")).toBe("true");
    // agenda merges planned + actual into one chronological list
    expect(await screen.findByRole("button", { name: /Deep Work — planned/ })).toBeDefined();
    expect(screen.getByRole("button", { name: /Admin — actual/ })).toBeDefined();
    expect(screen.queryByRole("list", { name: "Planned blocks" })).toBeNull();
  });

  it("switches between Day and Agenda via the view switcher", async () => {
    renderShell(<TimelineScreen />, { route: "/timeline" });
    await waitFor(() => expect(api.timeline).toHaveBeenCalled());
    await userEvent.click(screen.getByRole("radio", { name: "Agenda" }));
    expect(await screen.findByRole("group", { name: "Filter by category" })).toBeDefined();
    await userEvent.click(screen.getByRole("radio", { name: "Day" }));
    expect(await screen.findByRole("list", { name: "Planned blocks" })).toBeDefined();
  });

  it("refetches when the day is stepped", async () => {
    renderShell(<TimelineScreen />, { route: "/timeline?date=2026-09-10" });
    await waitFor(() => expect(api.timeline).toHaveBeenCalledWith("2026-09-10"));
    await userEvent.click(screen.getByRole("button", { name: "Previous day" }));
    await waitFor(() => expect(api.timeline).toHaveBeenCalledWith("2026-09-09"));
  });

  it("shows an error state with retry when the load fails", async () => {
    vi.mocked(api.timeline).mockRejectedValueOnce(new Error("boom"));
    renderShell(<TimelineScreen />, { route: "/timeline" });
    expect(await screen.findByText("Could not load the timeline.")).toBeDefined();
    expect(screen.getByRole("button", { name: "Retry" })).toBeDefined();
  });

  it("opens the add-block dialog from the header action", async () => {
    renderShell(<TimelineScreen />, { route: "/timeline" });
    await waitFor(() => expect(api.listCategories).toHaveBeenCalled());
    await userEvent.click(screen.getByRole("button", { name: "Add block" }));
    expect(screen.getByRole("dialog")).toBeDefined();
    expect(screen.getByRole("heading", { name: "Add block" })).toBeDefined();
  });

  it("opens the matching block's dialog when ?openBlock= is present", async () => {
    renderShell(<TimelineScreen />, { route: "/timeline?date=2026-09-01&openBlock=p1" });
    await waitFor(() => expect(api.timeline).toHaveBeenCalledWith("2026-09-01"));
    expect(await screen.findByRole("heading", { name: "Edit block" })).toBeDefined();
    expect((screen.getByLabelText("Start") as HTMLInputElement).value).toBe("09:00");
  });

  it("switches to Week view (G2) and shows the ISO week range as the header title", async () => {
    renderShell(<TimelineScreen />, { route: "/timeline?date=2026-09-03" });
    await waitFor(() => expect(api.timeline).toHaveBeenCalledWith("2026-09-03"));
    await userEvent.click(screen.getByRole("radio", { name: "Week" }));
    expect(await screen.findByRole("heading", { name: "Aug 31 – Sep 6" })).toBeDefined();
    // Week fetches the whole ISO week (Monday-first, D8) in one range call, not per-day.
    await waitFor(() => expect(api.timelineRange).toHaveBeenCalledWith("2026-08-31", "2026-09-06"));
  });

  it("switches to Month view (G2) and shows the month name as the header title", async () => {
    renderShell(<TimelineScreen />, { route: "/timeline?date=2026-09-03" });
    await waitFor(() => expect(api.timeline).toHaveBeenCalledWith("2026-09-03"));
    await userEvent.click(screen.getByRole("radio", { name: "Month" }));
    expect(await screen.findByRole("heading", { name: "September 2026" })).toBeDefined();
  });

  it("Week/Month step `‹`/`›` by their own unit, not a day", async () => {
    renderShell(<TimelineScreen />, { route: "/timeline?date=2026-09-03&view=week" });
    await waitFor(() => expect(api.timelineRange).toHaveBeenCalledWith("2026-08-31", "2026-09-06"));
    await userEvent.click(screen.getByRole("button", { name: "Next week" }));
    await waitFor(() => expect(api.timelineRange).toHaveBeenCalledWith("2026-09-07", "2026-09-13")); // next Monday
  });
});
