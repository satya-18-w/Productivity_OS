import { describe, it, expect, beforeEach } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Routes, Route } from "react-router-dom";
import { AppShell } from "./AppShell";
import { ScreenLayout } from "./ScreenLayout";
import { renderShell } from "../test/renderShell";
import { mockViewport } from "../test/mockViewport";

function shellAt(route: string) {
  return renderShell(
    <Routes>
      <Route element={<AppShell />}>
        <Route path="/timeline" element={<ScreenLayout><h1>Timeline</h1></ScreenLayout>} />
      </Route>
    </Routes>,
    { route },
  );
}

describe("AppShell", () => {
  beforeEach(() => mockViewport(1440));

  it("renders the skip link, primary nav and main landmark", () => {
    shellAt("/timeline");
    expect(screen.getByRole("link", { name: "Skip to content" })).toBeDefined();
    expect(screen.getByRole("navigation", { name: "Primary" })).toBeDefined();
    const main = document.querySelector("main#main");
    expect(main).not.toBeNull();
  });

  it("uses the expanded sidebar on a wide viewport (no mobile top bar)", () => {
    shellAt("/timeline");
    expect(document.querySelector(".app-shell")?.getAttribute("data-shell-mode")).toBe("expanded");
    expect(screen.queryByRole("button", { name: "Open navigation" })).toBeNull();
  });

  it("collapses to icons between tablet and laptop", () => {
    mockViewport(900);
    shellAt("/timeline");
    expect(document.querySelector(".app-shell")?.getAttribute("data-shell-mode")).toBe("collapsed");
  });

  it("becomes a drawer on a narrow viewport and opens from the top bar", async () => {
    mockViewport(430);
    shellAt("/timeline");
    expect(document.querySelector(".app-shell")?.getAttribute("data-shell-mode")).toBe("drawer");
    const hamburger = screen.getByRole("button", { name: "Open navigation" });
    await userEvent.click(hamburger);
    const drawer = document.querySelector("dialog.sidebar-drawer") as HTMLDialogElement;
    expect(drawer.open).toBe(true);
  });
});

describe("ScreenLayout", () => {
  it("renders the rail region only when rail content is passed", () => {
    const { rerender } = renderShell(<ScreenLayout>main</ScreenLayout>);
    expect(screen.queryByRole("complementary")).toBeNull();
    rerender(<ScreenLayout rail={<div>widgets</div>} railLabel="Timeline details">main</ScreenLayout>);
    expect(screen.getByRole("complementary", { name: "Timeline details" })).toBeDefined();
  });
});
