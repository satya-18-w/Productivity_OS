import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { TaskThroughputReport } from "./TaskThroughputReport";

describe("TaskThroughputReport", () => {
  it("shows the completed-task count as a headline stat", () => {
    render(<TaskThroughputReport count={7} />);
    expect(screen.getByText("7")).toBeDefined();
    expect(screen.getByText("Tasks completed")).toBeDefined();
  });
});
