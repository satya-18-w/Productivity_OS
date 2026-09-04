import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TimelineGrid } from "./TimelineGrid";
import type { PositionedBlock } from "../../api";

function block(over: Partial<PositionedBlock>): PositionedBlock {
  return {
    id: "b1",
    kind: "planned",
    starts_at: "",
    ends_at: "",
    category_id: "c1",
    category_name: "Deep Work",
    start_minute: 540, // 09:00
    end_minute: 660, // 11:00
    from_prev_day: false,
    to_next_day: false,
    local_date: "2026-09-04",
    local_start: "09:00",
    local_end: "11:00",
    ends_next_day: false,
    ...over,
  };
}

describe("TimelineGrid", () => {
  it("places a planned block in the Planned lane, dashed, time-proportional", () => {
    render(<TimelineGrid planned={[block({})]} actual={[]} now={null} onPick={() => {}} />);
    const btn = screen.getByRole("button", { name: /Deep Work — planned, 09:00–11:00/ });
    expect(btn.className).toContain("tl2__block--planned");
    expect(btn.style.top).toBe(`${(540 / 1440) * 100}%`);
    expect(btn.style.height).toBe(`${(120 / 1440) * 100}%`);
    expect(screen.getByRole("list", { name: "Planned blocks" }).contains(btn)).toBe(true);
  });

  it("renders an actual block solid in the Actual lane", () => {
    render(
      <TimelineGrid planned={[]} actual={[block({ id: "b2", kind: "actual" })]} now={null} onPick={() => {}} />,
    );
    const btn = screen.getByRole("button", { name: /Deep Work — actual/ });
    expect(btn.className).not.toContain("tl2__block--planned");
    expect(screen.getByRole("list", { name: "Actual blocks" }).contains(btn)).toBe(true);
  });

  it("shows the ▼ marker and 'continues into the next day' for a midnight block", () => {
    render(
      <TimelineGrid
        planned={[block({ to_next_day: true, end_minute: 1440, local_end: "24:00" })]}
        actual={[]}
        now={null}
        onPick={() => {}}
      />,
    );
    expect(screen.getByRole("button", { name: /continues into the next day/ })).toBeDefined();
  });

  it("shows the empty hint per lane", () => {
    render(<TimelineGrid planned={[]} actual={[]} now={null} onPick={() => {}} />);
    expect(screen.getByText("Nothing planned")).toBeDefined();
    expect(screen.getByText("Nothing actual")).toBeDefined();
  });

  it("renders the now-line only when `now` is provided", () => {
    const { container, rerender } = render(
      <TimelineGrid planned={[]} actual={[]} now={null} onPick={() => {}} />,
    );
    expect(container.querySelector(".tl2__now")).toBeNull();
    rerender(<TimelineGrid planned={[]} actual={[]} now={600} onPick={() => {}} />);
    expect(container.querySelectorAll(".tl2__now").length).toBe(2); // one per lane
  });

  it("calls onPick with the block when clicked", async () => {
    const onPick = vi.fn();
    const b = block({});
    render(<TimelineGrid planned={[b]} actual={[]} now={null} onPick={onPick} />);
    await userEvent.click(screen.getByRole("button", { name: /Deep Work — planned/ }));
    expect(onPick).toHaveBeenCalledWith(b);
  });
});
