import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ThemeToggle } from "./ThemeToggle";
import { setThemePreference } from "../theme";

beforeEach(() => setThemePreference("system"));
afterEach(() => {
  setThemePreference("system");
  document.documentElement.removeAttribute("data-theme");
});

describe("ThemeToggle", () => {
  it("is a labelled radiogroup with three options", () => {
    render(<ThemeToggle />);
    const group = screen.getByRole("radiogroup", { name: "Theme" });
    expect(group).toBeDefined();
    expect(screen.getByRole("radio", { name: "Light" })).toBeDefined();
    expect(screen.getByRole("radio", { name: "Dark" })).toBeDefined();
    expect(screen.getByRole("radio", { name: "System" })).toBeDefined();
  });

  it("selecting Dark stamps data-theme and persists it", async () => {
    render(<ThemeToggle />);
    await userEvent.click(screen.getByRole("radio", { name: "Dark" }));
    expect(document.documentElement.getAttribute("data-theme")).toBe("dark");
    expect(localStorage.getItem("pos-theme")).toBe("dark");
  });

  it("selecting System clears the attribute", async () => {
    render(<ThemeToggle />);
    await userEvent.click(screen.getByRole("radio", { name: "Light" }));
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");
    await userEvent.click(screen.getByRole("radio", { name: "System" }));
    expect(document.documentElement.hasAttribute("data-theme")).toBe(false);
    expect(localStorage.getItem("pos-theme")).toBeNull();
  });

  it("compact mode renders a single cycling button", async () => {
    render(<ThemeToggle compact />);
    const btn = screen.getByRole("button");
    await userEvent.click(btn); // system -> light
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");
  });
});
