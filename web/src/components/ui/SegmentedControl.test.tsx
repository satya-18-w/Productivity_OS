import { describe, it, expect, vi } from "vitest";
import { useState } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SegmentedControl } from "./SegmentedControl";

const OPTIONS = [
  { value: "day", label: "Day" },
  { value: "agenda", label: "Agenda" },
] as const;

function Harness({ onChange }: { onChange?: (v: string) => void }) {
  const [value, setValue] = useState<(typeof OPTIONS)[number]["value"]>("day");
  return (
    <SegmentedControl
      label="Timeline view"
      options={OPTIONS}
      value={value}
      onChange={(v) => {
        setValue(v);
        onChange?.(v);
      }}
    />
  );
}

describe("SegmentedControl", () => {
  it("renders as a labelled radiogroup with one checked option", () => {
    render(<Harness />);
    const group = screen.getByRole("radiogroup", { name: "Timeline view" });
    expect(group).toBeDefined();
    expect((screen.getByRole("radio", { name: "Day" }) as HTMLElement).getAttribute("aria-checked")).toBe("true");
  });

  it("changes selection on click", async () => {
    const onChange = vi.fn();
    render(<Harness onChange={onChange} />);
    await userEvent.click(screen.getByRole("radio", { name: "Agenda" }));
    expect(onChange).toHaveBeenCalledWith("agenda");
    expect(screen.getByRole("radio", { name: "Agenda" }).getAttribute("aria-checked")).toBe("true");
  });

  it("moves selection with arrow keys (roving tabindex)", async () => {
    render(<Harness />);
    const day = screen.getByRole("radio", { name: "Day" });
    day.focus();
    await userEvent.keyboard("{ArrowRight}");
    const agenda = screen.getByRole("radio", { name: "Agenda" });
    expect(agenda.getAttribute("aria-checked")).toBe("true");
    expect(agenda.getAttribute("tabindex")).toBe("0");
    expect(day.getAttribute("tabindex")).toBe("-1");
  });
});
