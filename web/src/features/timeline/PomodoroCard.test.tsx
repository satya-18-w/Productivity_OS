import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PomodoroCard } from "./PomodoroCard";

beforeEach(() => {
  vi.useFakeTimers({ shouldAdvanceTime: true });
});
afterEach(() => {
  vi.useRealTimers();
});

describe("PomodoroCard", () => {
  it("defaults to a 25:00 Focus session", () => {
    render(<PomodoroCard />);
    expect(screen.getByText("25:00")).toBeDefined();
    expect(screen.getByRole("radio", { name: "Focus", checked: true })).toBeDefined();
  });

  it("counts down once started, and never calls any API (standalone, v1.md §4)", async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<PomodoroCard />);
    await user.click(screen.getByRole("button", { name: "Start" }));
    await vi.advanceTimersByTimeAsync(3000);
    expect(screen.getByText("24:57")).toBeDefined();
  });

  it("pauses and resumes", async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<PomodoroCard />);
    await user.click(screen.getByRole("button", { name: "Start" }));
    await vi.advanceTimersByTimeAsync(2000);
    await user.click(screen.getByRole("button", { name: "Pause" }));
    expect(screen.getByText("24:58")).toBeDefined();
    await vi.advanceTimersByTimeAsync(5000);
    expect(screen.getByText("24:58")).toBeDefined(); // unchanged while paused
    await user.click(screen.getByRole("button", { name: "Resume" }));
    await vi.advanceTimersByTimeAsync(2000);
    expect(screen.getByText("24:56")).toBeDefined();
  });

  it("switching preset resets to that preset's full duration", async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<PomodoroCard />);
    await user.click(screen.getByRole("radio", { name: "Short break" }));
    expect(screen.getByText("05:00")).toBeDefined();
  });

  it("Reset restores the full duration and stops the countdown", async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<PomodoroCard />);
    await user.click(screen.getByRole("button", { name: "Start" }));
    await vi.advanceTimersByTimeAsync(3000);
    await user.click(screen.getByRole("button", { name: "Reset" }));
    expect(screen.getByText("25:00")).toBeDefined();
    expect(screen.getByRole("button", { name: "Start" })).toBeDefined();
  });
});
