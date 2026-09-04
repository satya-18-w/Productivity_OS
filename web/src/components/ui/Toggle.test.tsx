import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Switch, ToggleCircle } from "./Toggle";

describe("Switch", () => {
  it("exposes the switch role and toggles", async () => {
    const onChange = vi.fn();
    render(<Switch label="Compact mode" onChange={onChange} />);
    const sw = screen.getByRole("switch", { name: "Compact mode" });
    await userEvent.click(sw);
    expect(onChange).toHaveBeenCalledTimes(1);
    expect((sw as HTMLInputElement).checked).toBe(true);
  });
});

describe("ToggleCircle", () => {
  it("has an accessible name and toggles on click", async () => {
    render(<ToggleCircle label="Meditation — Mon 1 Sep" />);
    const t = screen.getByRole("checkbox", { name: "Meditation — Mon 1 Sep" }) as HTMLInputElement;
    expect(t.checked).toBe(false);
    await userEvent.click(t);
    expect(t.checked).toBe(true);
  });

  it("reflects the controlled checked prop", () => {
    render(<ToggleCircle label="Read — Tue 2 Sep" checked readOnly />);
    expect(
      (screen.getByRole("checkbox", { name: "Read — Tue 2 Sep" }) as HTMLInputElement).checked,
    ).toBe(true);
  });
});
