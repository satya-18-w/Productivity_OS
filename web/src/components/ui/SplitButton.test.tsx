import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SplitButton } from "./SplitButton";

function setup() {
  const onPrimary = vi.fn();
  const onPlanned = vi.fn();
  const onActual = vi.fn();
  render(
    <SplitButton
      onPrimary={onPrimary}
      menuLabel="Add options"
      items={[
        { key: "planned", label: "Add planned block", onSelect: onPlanned },
        { key: "actual", label: "Add actual block", onSelect: onActual },
      ]}
    >
      Add block
    </SplitButton>,
  );
  return { onPrimary, onPlanned, onActual };
}

describe("SplitButton", () => {
  it("fires the primary action from the main segment", async () => {
    const { onPrimary, onPlanned } = setup();
    await userEvent.click(screen.getByRole("button", { name: "Add block" }));
    expect(onPrimary).toHaveBeenCalledTimes(1);
    expect(onPlanned).not.toHaveBeenCalled();
  });

  it("opens the menu from the caret and selects an item", async () => {
    const { onActual } = setup();
    await userEvent.click(screen.getByRole("button", { name: "Add block, more actions" }));
    expect(screen.getByRole("menu", { name: "Add options" })).toBeDefined();
    await userEvent.click(screen.getByRole("menuitem", { name: "Add actual block" }));
    expect(onActual).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("menu")).toBeNull();
  });

  it("disables both segments together", () => {
    render(
      <SplitButton onPrimary={() => {}} menuLabel="Add options" items={[]} disabled>
        Add block
      </SplitButton>,
    );
    expect(screen.getByRole("button", { name: "Add block" }).hasAttribute("disabled")).toBe(true);
    expect(
      screen.getByRole("button", { name: "Add block, more actions" }).hasAttribute("disabled"),
    ).toBe(true);
  });
});
