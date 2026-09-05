import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DateStepper } from "./DateStepper";
import { todayISO, shiftDays } from "./dateUtils";

describe("DateStepper", () => {
  it("steps back and forward a day", async () => {
    const onChange = vi.fn();
    render(<DateStepper value="2026-08-15" onChange={onChange} />);
    await userEvent.click(screen.getByRole("button", { name: "Previous day" }));
    expect(onChange).toHaveBeenLastCalledWith("2026-08-14");
    await userEvent.click(screen.getByRole("button", { name: "Next day" }));
    expect(onChange).toHaveBeenLastCalledWith("2026-08-16");
  });

  it("jumps to today and disables Today when already there", async () => {
    const onChange = vi.fn();
    const { rerender } = render(<DateStepper value={shiftDays(todayISO(), -1)} onChange={onChange} />);
    const todayBtn = screen.getByRole("button", { name: "Today" }) as HTMLButtonElement;
    expect(todayBtn.disabled).toBe(false);
    await userEvent.click(todayBtn);
    expect(onChange).toHaveBeenCalledWith(todayISO());

    rerender(<DateStepper value={todayISO()} onChange={onChange} />);
    expect((screen.getByRole("button", { name: "Today" }) as HTMLButtonElement).disabled).toBe(true);
  });

  it("lets the date input jump directly, with a custom accessible label", () => {
    render(<DateStepper value="2026-08-15" onChange={vi.fn()} label="Review date" />);
    expect(screen.getByLabelText("Review date")).toBeDefined();
  });

  it("lets a caller override the step unit (e.g. Timeline Week/Month, G2)", async () => {
    const onStep = vi.fn();
    render(
      <DateStepper
        value="2026-09-05"
        onChange={vi.fn()}
        onStep={onStep}
        prevLabel="Previous week"
        nextLabel="Next week"
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: "Previous week" }));
    expect(onStep).toHaveBeenCalledWith(-1);
    await userEvent.click(screen.getByRole("button", { name: "Next week" }));
    expect(onStep).toHaveBeenCalledWith(1);
  });
});
