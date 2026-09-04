import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { App } from "./App";
import { AuthStub } from "./test/renderShell";
import { mockViewport } from "./test/mockViewport";

// Keep routing tests hermetic — stub the screens that would hit the API on mount.
vi.mock("./features/timeline", () => ({ TimelineScreen: () => <h1>Timeline page</h1> }));
vi.mock("./pages/Board", () => ({ Board: () => <h1>Board page</h1> }));
vi.mock("./pages/Habits", () => ({ Habits: () => <h1>Habits page</h1> }));
vi.mock("./pages/Goals", () => ({ Goals: () => <h1>Goals page</h1> }));
vi.mock("./pages/Categories", () => ({ Categories: () => <h1>Categories page</h1> }));
vi.mock("./pages/Account", () => ({ Account: () => <h1>Account page</h1> }));

function renderApp(route: string, account: { email: string; timezone: string } | null = { email: "a@b.co", timezone: "UTC" }) {
  return render(
    <MemoryRouter initialEntries={[route]}>
      <AuthStub account={account}>
        <App />
      </AuthStub>
    </MemoryRouter>,
  );
}

describe("App routing (D10)", () => {
  beforeEach(() => mockViewport(1440));

  it("lands '/' on the Timeline screen (no dashboard)", () => {
    renderApp("/");
    expect(screen.getByRole("heading", { name: "Timeline page" })).toBeDefined();
    expect(screen.getByRole("navigation", { name: "Primary" })).toBeDefined();
  });

  it("renders placeholders for not-yet-built screens", () => {
    renderApp("/tasks");
    expect(screen.getByRole("heading", { level: 1, name: "Tasks" })).toBeDefined();
  });

  it("redirects unknown routes to the landing screen", () => {
    renderApp("/nope/nope");
    expect(screen.getByRole("heading", { name: "Timeline page" })).toBeDefined();
  });

  it("sends unauthenticated users to /login (guard intact)", () => {
    renderApp("/goals", null);
    expect(screen.getByRole("heading", { name: "Welcome back" })).toBeDefined();
    expect(screen.queryByRole("navigation", { name: "Primary" })).toBeNull();
  });

  it("keeps auth screens outside the shell", () => {
    renderApp("/login", null);
    expect(screen.queryByRole("navigation", { name: "Primary" })).toBeNull();
  });
});
