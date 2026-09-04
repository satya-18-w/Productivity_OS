import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MiniCalendar } from "./MiniCalendar";

describe("MiniCalendar", () => {
  it("shows the month of the selected date, Monday-first (D8)", () => {
    const { container } = render(<MiniCalendar value="2026-09-04" onChange={() => {}} />);
    expect(screen.getByText("September 2026")).toBeDefined();
    const weekdays = container.querySelector(".ui-minical__weekdays")!;
    expect(weekdays.textContent).toBe("MonTueWedThuFriSatSun");
  });

  it("marks the selected day pressed", () => {
    render(<MiniCalendar value="2026-09-04" onChange={() => {}} />);
    const day = screen.getByRole("button", { name: /September 4, 2026/ });
    expect(day.getAttribute("aria-pressed")).toBe("true");
  });

  it("fires onChange with the clicked day's ISO date", async () => {
    const onChange = vi.fn();
    render(<MiniCalendar value="2026-09-04" onChange={onChange} />);
    await userEvent.click(screen.getByRole("button", { name: /September 10, 2026/ }));
    expect(onChange).toHaveBeenCalledWith("2026-09-10");
  });

  it("navigates months without changing the selection", async () => {
    const onChange = vi.fn();
    render(<MiniCalendar value="2026-09-04" onChange={onChange} />);
    await userEvent.click(screen.getByRole("button", { name: "Next month" }));
    expect(screen.getByText("October 2026")).toBeDefined();
    expect(onChange).not.toHaveBeenCalled();
  });
});
