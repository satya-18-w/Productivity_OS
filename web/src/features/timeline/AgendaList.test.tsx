import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AgendaList } from "./AgendaList";
import type { PositionedBlock } from "../../api";

function b(over: Partial<PositionedBlock>): PositionedBlock {
  return {
    id: "x", kind: "planned", starts_at: "", ends_at: "", category_id: "c1",
    category_name: "Deep Work", task_id: null, start_minute: 540, end_minute: 660,
    from_prev_day: false, to_next_day: false, local_date: "2026-09-04",
    local_start: "09:00", local_end: "11:00", ends_next_day: false, ...over,
  };
}

const PLANNED = [
  b({ id: "p1", start_minute: 540, end_minute: 660 }),
  b({ id: "p2", category_id: "c2", category_name: "Exercise", start_minute: 390, end_minute: 450 }),
];
const ACTUAL = [b({ id: "a1", kind: "actual", start_minute: 720, end_minute: 780 })];

describe("AgendaList", () => {
  it("merges planned + actual into one time-ordered list", () => {
    render(<AgendaList planned={PLANNED} actual={ACTUAL} now={null} onPick={() => {}} />);
    const rows = screen.getAllByRole("button").filter((el) => el.className.includes("agenda__row"));
    expect(rows).toHaveLength(3);
    // first by start time is Exercise at 06:30
    expect(rows[0].getAttribute("aria-label")).toMatch(/Exercise/);
    expect(rows[2].getAttribute("aria-label")).toMatch(/actual/);
  });

  it("filters by category chip and can clear back to All", async () => {
    render(<AgendaList planned={PLANNED} actual={ACTUAL} now={null} onPick={() => {}} />);
    await userEvent.click(screen.getByRole("button", { name: /Exercise \(1\)/ }));
    let rows = screen.getAllByRole("button").filter((el) => el.className.includes("agenda__row"));
    expect(rows).toHaveLength(1);
    await userEvent.click(screen.getByRole("button", { name: /^All \(3\)/ }));
    rows = screen.getAllByRole("button").filter((el) => el.className.includes("agenda__row"));
    expect(rows).toHaveLength(3);
  });

  it("marks a row as past when it ends before `now`", () => {
    render(<AgendaList planned={PLANNED} actual={ACTUAL} now={500} onPick={() => {}} />);
    const exercise = screen
      .getAllByRole("button")
      .find((el) => el.getAttribute("aria-label")?.includes("Exercise"))!;
    expect(exercise.className).toContain("agenda__row--past");
  });

  it("calls onPick with the block", async () => {
    const onPick = vi.fn();
    render(<AgendaList planned={PLANNED} actual={[]} now={null} onPick={onPick} />);
    await userEvent.click(screen.getByRole("button", { name: /Deep Work — planned/ }));
    expect(onPick).toHaveBeenCalledWith(PLANNED[0]);
  });

  it("shows an empty state with no blocks", () => {
    render(<AgendaList planned={[]} actual={[]} now={null} onPick={() => {}} />);
    expect(screen.getByText("Nothing scheduled")).toBeDefined();
  });

  it("offers an add affordance that calls onAdd", async () => {
    const onAdd = vi.fn();
    render(<AgendaList planned={PLANNED} actual={ACTUAL} now={null} onPick={() => {}} onAdd={onAdd} />);
    // The "+" is aria-hidden (decorative), so it's excluded from the accessible name.
    await userEvent.click(screen.getByRole("button", { name: "Add an agenda item" }));
    expect(onAdd).toHaveBeenCalledTimes(1);
  });

  it("omits the add affordance when onAdd is not provided", () => {
    render(<AgendaList planned={PLANNED} actual={ACTUAL} now={null} onPick={() => {}} />);
    expect(screen.queryByRole("button", { name: /Add an agenda item/ })).toBeNull();
  });
});
