import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TimezoneSelect, timezoneOptions } from "./TimezoneSelect";
import { browserTimezone } from "../../api";

describe("timezoneOptions", () => {
  it("always offers UTC and the browser-detected zone", () => {
    const names = timezoneOptions();
    expect(names).toContain("UTC");
    expect(names).toContain(browserTimezone());
  });

  it("keeps an unknown current value selectable", () => {
    expect(timezoneOptions("Mars/Olympus")[0]).toBe("Mars/Olympus");
  });
});

describe("TimezoneSelect", () => {
  it("reports the chosen timezone", async () => {
    const onChange = vi.fn();
    render(<TimezoneSelect aria-label="Timezone" value="UTC" onChange={onChange} />);
    await userEvent.selectOptions(screen.getByLabelText("Timezone"), "UTC");
    expect(onChange).toHaveBeenCalledWith("UTC");
  });
});
