import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TimelineScreen } from "./TimelineScreen";
import { api, type DayTimeline, type PositionedBlock } from "../../api";
import { renderShell } from "../../test/renderShell";
import { todayISO } from "../../components/date/dateUtils";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return {
    ...actual,
    api: {
      ...actual.api,
      timeline: vi.fn(),
      comparison: vi.fn(),
      listCategories: vi.fn(),
    },
  };
});

function pblock(over: Partial<PositionedBlock>): PositionedBlock {
  return {
    id: "b1", kind: "planned", starts_at: "", ends_at: "", category_id: "c1",
    category_name: "Deep Work", start_minute: 540, end_minute: 660,
    from_prev_day: false, to_next_day: false, local_date: "2026-09-01",
    local_start: "09:00", local_end: "11:00", ends_next_day: false, ...over,
  };
}

const TL: DayTimeline = {
  date: "",
  planned: [pblock({ id: "p1" })],
  actual: [pblock({ id: "a1", kind: "actual", category_id: "c2", category_name: "Admin", start_minute: 720, end_minute: 780 })],
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.timeline).mockResolvedValue(TL);
  vi.mocked(api.comparison).mockResolvedValue({ date: "", categories: [] });
  vi.mocked(api.listCategories).mockResolvedValue([]);
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
});
