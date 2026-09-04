import { describe, it, expect } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { UserMenu } from "./UserMenu";
import { renderShell } from "../test/renderShell";

describe("UserMenu", () => {
  it("shows the account email on the trigger", () => {
    renderShell(<UserMenu />);
    expect(screen.getByRole("button", { name: /sam@example\.com/ })).toBeDefined();
  });

  it("opens a menu with Account, Export data and Log out", async () => {
    renderShell(<UserMenu />);
    await userEvent.click(screen.getByRole("button", { name: /sam@example\.com/ }));
    const menu = screen.getByRole("menu", { name: "Account menu" });
    expect(menu).toBeDefined();
    expect(screen.getByRole("menuitem", { name: "Account" }).getAttribute("href")).toBe("/account");
    expect(screen.getByRole("menuitem", { name: "Export data" }).getAttribute("href")).toBe("/export");
    expect(screen.getByRole("menuitem", { name: "Log out" })).toBeDefined();
  });

  it("closes on Escape", async () => {
    renderShell(<UserMenu />);
    const trigger = screen.getByRole("button", { name: /sam@example\.com/ });
    await userEvent.click(trigger);
    expect(screen.queryByRole("menu")).not.toBeNull();
    await userEvent.keyboard("{Escape}");
    expect(screen.queryByRole("menu")).toBeNull();
  });

  it("toggles aria-expanded", async () => {
    renderShell(<UserMenu />);
    const trigger = screen.getByRole("button", { name: /sam@example\.com/ });
    expect(trigger.getAttribute("aria-expanded")).toBe("false");
    await userEvent.click(trigger);
    expect(trigger.getAttribute("aria-expanded")).toBe("true");
  });
});
