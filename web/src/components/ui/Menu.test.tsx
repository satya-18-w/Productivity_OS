import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Menu } from "./Menu";
import { IconButton } from "./IconButton";

function setup(onSelect = vi.fn()) {
  render(
    <Menu
      label="Row actions"
      trigger={<IconButton label="Actions">⋯</IconButton>}
      items={[
        { key: "edit", label: "Edit", onSelect },
        { key: "sep", separator: true },
        { key: "del", label: "Delete", danger: true, onSelect },
      ]}
    />,
  );
  return { onSelect };
}

describe("Menu", () => {
  it("is a closed menu-button until activated", () => {
    setup();
    const trigger = screen.getByRole("button", { name: "Actions" });
    expect(trigger.getAttribute("aria-haspopup")).toBe("menu");
    expect(trigger.getAttribute("aria-expanded")).toBe("false");
    expect(screen.queryByRole("menu")).toBeNull();
  });

  it("opens on click and lists its items", async () => {
    setup();
    await userEvent.click(screen.getByRole("button", { name: "Actions" }));
    expect(screen.getByRole("menu", { name: "Row actions" })).toBeDefined();
    expect(screen.getByRole("menuitem", { name: "Edit" })).toBeDefined();
    expect(screen.getByRole("menuitem", { name: "Delete" })).toBeDefined();
  });

  it("selects an item and closes", async () => {
    const { onSelect } = setup();
    await userEvent.click(screen.getByRole("button", { name: "Actions" }));
    await userEvent.click(screen.getByRole("menuitem", { name: "Edit" }));
    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("menu")).toBeNull();
  });

  it("closes on Escape and returns focus to the trigger", async () => {
    setup();
    const trigger = screen.getByRole("button", { name: "Actions" });
    await userEvent.click(trigger);
    await userEvent.keyboard("{Escape}");
    expect(screen.queryByRole("menu")).toBeNull();
    expect(document.activeElement).toBe(trigger);
  });

  it("moves through items with the arrow keys", async () => {
    setup();
    await userEvent.click(screen.getByRole("button", { name: "Actions" }));
    // first item focused on open
    expect(document.activeElement).toBe(screen.getByRole("menuitem", { name: "Edit" }));
    await userEvent.keyboard("{ArrowDown}");
    expect(document.activeElement).toBe(screen.getByRole("menuitem", { name: "Delete" }));
  });
});
