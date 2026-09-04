import { describe, it, expect } from "vitest";
import { screen } from "@testing-library/react";
import { Routes, Route } from "react-router-dom";
import { Sidebar } from "./Sidebar";
import { NAV_ITEMS } from "./navItems";
import { renderShell } from "../test/renderShell";

describe("Sidebar", () => {
  it("renders every primary nav destination as a link", () => {
    renderShell(<Sidebar mode="expanded" />);
    for (const item of NAV_ITEMS) {
      const link = screen.getByRole("link", { name: item.label });
      expect(link.getAttribute("href")).toBe(item.to);
    }
  });

  it("does not render excluded destinations", () => {
    renderShell(<Sidebar mode="expanded" />);
    for (const gone of ["Dashboard", "Notes", "Calendar", "Analytics", "Spaces"]) {
      expect(screen.queryByRole("link", { name: gone })).toBeNull();
    }
  });

  it("marks the active route with aria-current", () => {
    renderShell(
      <Routes>
        <Route path="/goals" element={<Sidebar mode="expanded" />} />
      </Routes>,
      { route: "/goals" },
    );
    const active = screen.getByRole("link", { name: "Goals" });
    expect(active.getAttribute("aria-current")).toBe("page");
    expect(active.className).toContain("active");
    expect(screen.getByRole("link", { name: "Tasks" }).getAttribute("aria-current")).toBeNull();
  });

  it("keeps nav labels accessible in the collapsed (icon-only) sidebar", () => {
    renderShell(<Sidebar mode="collapsed" />);
    // label text is visually hidden but the accessible name survives
    expect(screen.getByRole("link", { name: "Habits" })).toBeDefined();
  });

  it("exposes the primary nav landmark", () => {
    renderShell(<Sidebar mode="expanded" />);
    expect(screen.getByRole("navigation", { name: "Primary" })).toBeDefined();
  });
});
