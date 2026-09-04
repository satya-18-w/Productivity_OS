import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { StatusBadge } from "./StatusBadge";
import { CategoryIndicator } from "./CategoryIndicator";
import { StatCard } from "./StatCard";
import { EmptyState } from "./states";
import { categoryColor } from "./categoryColor";

describe("StatusBadge", () => {
  it("uses the four V1 goal labels verbatim", () => {
    const { rerender } = render(<StatusBadge status="not_started" />);
    expect(screen.getByText("Not started")).toBeDefined();
    rerender(<StatusBadge status="in_progress" />);
    expect(screen.getByText("In progress")).toBeDefined();
    rerender(<StatusBadge status="achieved" />);
    expect(screen.getByText("Achieved")).toBeDefined();
    rerender(<StatusBadge status="abandoned" />);
    expect(screen.getByText("Abandoned")).toBeDefined();
  });
});

describe("CategoryIndicator", () => {
  it("conveys the category by name, not colour alone (VP8)", () => {
    render(<CategoryIndicator variant="chip" name="Deep Work" colorKey="cat-1" />);
    expect(screen.getByText("Deep Work")).toBeDefined();
  });

  it("labels the dot variant for assistive tech", () => {
    render(<CategoryIndicator variant="dot" name="Health" colorKey="h" />);
    expect(screen.getByRole("img", { name: "Health" })).toBeDefined();
  });
});

describe("categoryColor", () => {
  it("is deterministic and falls back to the unset hue", () => {
    expect(categoryColor("abc")).toBe(categoryColor("abc"));
    expect(categoryColor(null)).toBe("var(--cat-other)");
  });
});

describe("StatCard", () => {
  it("shows value and label with no delta affordance", () => {
    const { container } = render(<StatCard label="Tasks done" value="5 / 8" />);
    expect(screen.getByText("5 / 8")).toBeDefined();
    expect(screen.getByText("Tasks done")).toBeDefined();
    expect(container.textContent).not.toMatch(/vs last|%/);
  });
});

describe("EmptyState", () => {
  it("renders a status region with a title and message", () => {
    render(<EmptyState title="No tasks yet" message="Add a task to get started." />);
    const region = screen.getByRole("status");
    expect(region.textContent).toContain("No tasks yet");
    expect(region.textContent).toContain("Add a task to get started.");
  });
});
