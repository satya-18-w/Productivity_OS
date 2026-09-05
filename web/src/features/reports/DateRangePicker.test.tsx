import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DateRangePicker } from "./DateRangePicker";

describe("DateRangePicker", () => {
  it("shows the current from/to values", () => {
    render(<DateRangePicker from="2026-08-01" to="2026-08-07" onChange={vi.fn()} />);
    expect((screen.getByLabelText("From") as HTMLInputElement).value).toBe("2026-08-01");
    expect((screen.getByLabelText("To") as HTMLInputElement).value).toBe("2026-08-07");
  });

  it("calls onChange with the new to date, keeping from", () => {
    const onChange = vi.fn();
    render(<DateRangePicker from="2026-08-01" to="2026-08-07" onChange={onChange} />);
    fireEvent.change(screen.getByLabelText("To"), { target: { value: "2026-08-10" } });
    expect(onChange).toHaveBeenCalledWith({ from: "2026-08-01", to: "2026-08-10" });
  });

  it("does not cap the To date at today (future dates are allowed, Q9)", () => {
    render(<DateRangePicker from="2026-08-01" to="2026-08-07" onChange={vi.fn()} />);
    expect(screen.getByLabelText("To").getAttribute("max")).toBeNull();
  });

  it("applies a preset range on click", async () => {
    const onChange = vi.fn();
    render(<DateRangePicker from="2026-08-01" to="2026-08-07" onChange={onChange} />);
    await userEvent.click(screen.getByRole("button", { name: "Last 7 days" }));
    expect(onChange).toHaveBeenCalled();
    const { from, to } = onChange.mock.calls.at(-1)![0];
    const days = (new Date(to).getTime() - new Date(from).getTime()) / 86_400_000;
    expect(Math.round(days)).toBe(6);
  });
});
