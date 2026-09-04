import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Tabs } from "./Tabs";

const ITEMS = [
  { value: "a", label: "First", content: <p>Panel A</p> },
  { value: "b", label: "Second", content: <p>Panel B</p> },
  { value: "c", label: "Third", content: <p>Panel C</p> },
];

describe("Tabs", () => {
  it("shows the first panel by default and wires ARIA", () => {
    render(<Tabs items={ITEMS} label="Sections" />);
    expect(screen.getByRole("tablist", { name: "Sections" })).toBeDefined();
    const first = screen.getByRole("tab", { name: "First" });
    expect(first.getAttribute("aria-selected")).toBe("true");
    expect(screen.getByText("Panel A")).toBeDefined();
    const panel = screen.getByRole("tabpanel");
    expect(panel.getAttribute("aria-labelledby")).toBe(first.getAttribute("id"));
  });

  it("switches panels on tab click", async () => {
    render(<Tabs items={ITEMS} label="Sections" />);
    await userEvent.click(screen.getByRole("tab", { name: "Second" }));
    expect(screen.getByText("Panel B")).toBeDefined();
    expect(screen.queryByText("Panel A")).toBeNull();
  });

  it("navigates with arrow keys", async () => {
    render(<Tabs items={ITEMS} label="Sections" />);
    screen.getByRole("tab", { name: "First" }).focus();
    await userEvent.keyboard("{ArrowRight}{ArrowRight}");
    expect(screen.getByRole("tab", { name: "Third" }).getAttribute("aria-selected")).toBe("true");
    await userEvent.keyboard("{Home}");
    expect(screen.getByRole("tab", { name: "First" }).getAttribute("aria-selected")).toBe("true");
  });

  it("supports controlled usage", async () => {
    const seen: string[] = [];
    render(<Tabs items={ITEMS} label="Sections" value="a" onValueChange={(v) => seen.push(v)} />);
    await userEvent.click(screen.getByRole("tab", { name: "Third" }));
    expect(seen).toEqual(["c"]);
    // still shows A because the parent controls `value`
    expect(screen.getByText("Panel A")).toBeDefined();
  });
});
