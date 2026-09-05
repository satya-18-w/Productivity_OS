import { describe, it, expect } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ReportsScreen } from "./ReportsScreen";
import { renderShell } from "../../test/renderShell";

describe("ReportsScreen", () => {
  it("renders all five fixed reports and the sample-data notice", () => {
    renderShell(<ReportsScreen />, { route: "/reports" });
    expect(screen.getByRole("heading", { name: "Time by category" })).toBeDefined();
    expect(screen.getByRole("heading", { name: "Planned vs actual by category" })).toBeDefined();
    expect(screen.getByRole("heading", { name: "Habit completion" })).toBeDefined();
    expect(screen.getByRole("heading", { name: "Task throughput" })).toBeDefined();
    expect(screen.getByRole("heading", { name: "Daily actual totals" })).toBeDefined();
    expect(screen.getByRole("note").textContent).toMatch(/Sample data/);
  });

  it("defaults the range to the trailing 30 days when the URL has none", () => {
    renderShell(<ReportsScreen />, { route: "/reports" });
    const from = screen.getByLabelText("From") as HTMLInputElement;
    const to = screen.getByLabelText("To") as HTMLInputElement;
    const days = (new Date(to.value).getTime() - new Date(from.value).getTime()) / 86_400_000;
    expect(Math.round(days)).toBe(29);
  });

  it("reads an explicit range from the URL", () => {
    renderShell(<ReportsScreen />, { route: "/reports?from=2026-08-01&to=2026-08-07" });
    expect((screen.getByLabelText("From") as HTMLInputElement).value).toBe("2026-08-01");
    expect((screen.getByLabelText("To") as HTMLInputElement).value).toBe("2026-08-07");
  });

  it("updates the reports when a preset range is chosen", async () => {
    renderShell(<ReportsScreen />, { route: "/reports?from=2026-08-01&to=2026-08-07" });
    await userEvent.click(screen.getByRole("button", { name: "Last 7 days" }));
    const from = screen.getByLabelText("From") as HTMLInputElement;
    expect(from.value).not.toBe("2026-08-01");
  });

  it("normalises an inverted range from the URL so every report agrees", () => {
    renderShell(<ReportsScreen />, { route: "/reports?from=2026-08-10&to=2026-08-01" });
    expect((screen.getByLabelText("From") as HTMLInputElement).value).toBe("2026-08-01");
    expect((screen.getByLabelText("To") as HTMLInputElement).value).toBe("2026-08-10");
    // 10-day inclusive range → 10 daily bars, habit denominators of 10.
    expect(screen.getAllByText(/\/ 10 days \(/).length).toBeGreaterThan(0);
  });
});
