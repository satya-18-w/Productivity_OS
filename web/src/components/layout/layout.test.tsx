import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { PageHeader } from "./PageHeader";
import { Stack } from "./Stack";
import { breakpoints } from "../../styles/breakpoints";
import { up, down } from "../../styles/breakpoints";

describe("PageHeader", () => {
  it("renders the title as the page h1 with optional eyebrow and subtitle", () => {
    render(<PageHeader eyebrow="TASKS" title="Tasks" subtitle="Turn your to-dos into progress." />);
    expect(screen.getByRole("heading", { level: 1, name: "Tasks" })).toBeDefined();
    expect(screen.getByText("TASKS")).toBeDefined();
    expect(screen.getByText("Turn your to-dos into progress.")).toBeDefined();
  });
});

describe("Stack", () => {
  it("maps the gap prop to a spacing token", () => {
    const { container } = render(
      <Stack gap={5}>
        <div />
      </Stack>,
    );
    expect((container.firstChild as HTMLElement).style.gap).toBe("var(--sp-5)");
  });
});

describe("breakpoints", () => {
  it("keeps the ratified shed order (tablet < laptop < wide)", () => {
    expect(breakpoints.tablet).toBeLessThan(breakpoints.laptop);
    expect(breakpoints.laptop).toBeLessThan(breakpoints.wide);
  });

  it("builds sane media query strings", () => {
    expect(up("laptop")).toBe("(min-width: 1024px)");
    expect(down("tablet")).toBe("(max-width: 639.98px)");
  });
});
