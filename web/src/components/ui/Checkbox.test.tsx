import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Checkbox } from "./Checkbox";

describe("Checkbox", () => {
  it("associates the visible label with the input", async () => {
    render(<Checkbox label="Remember me" />);
    const box = screen.getByRole("checkbox", { name: "Remember me" });
    expect(box).toBeDefined();
    await userEvent.click(screen.getByText("Remember me"));
    expect((box as HTMLInputElement).checked).toBe(true);
  });

  it("fires onChange with the toggled state", async () => {
    const onChange = vi.fn();
    render(<Checkbox label="Done" onChange={onChange} />);
    await userEvent.click(screen.getByRole("checkbox", { name: "Done" }));
    expect(onChange).toHaveBeenCalledTimes(1);
  });

  it("is operable with the keyboard", async () => {
    render(<Checkbox label="Done" />);
    const box = screen.getByRole("checkbox", { name: "Done" }) as HTMLInputElement;
    box.focus();
    await userEvent.keyboard(" ");
    expect(box.checked).toBe(true);
  });

  it("respects disabled", async () => {
    const onChange = vi.fn();
    render(<Checkbox label="Done" disabled onChange={onChange} />);
    await userEvent.click(screen.getByRole("checkbox", { name: "Done" }));
    expect(onChange).not.toHaveBeenCalled();
  });
});
