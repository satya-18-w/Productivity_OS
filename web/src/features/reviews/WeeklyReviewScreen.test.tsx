import { describe, it, expect } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { WeeklyReviewScreen } from "./WeeklyReviewScreen";
import { renderShell } from "../../test/renderShell";

describe("WeeklyReviewScreen", () => {
  it("shows the four fixed Q2 prompts and the mocked weekly reference", async () => {
    renderShell(<WeeklyReviewScreen />, { route: "/reviews/weekly" });
    expect(await screen.findByText("What were the highlights of this week?")).toBeDefined();
    expect(screen.getByText("What did you struggle with?")).toBeDefined();
    expect(screen.getByText("Did your time go where you intended?")).toBeDefined();
    expect(screen.getByText("What is the one priority for next week?")).toBeDefined();
    // mocked reference panel with sample-data notice
    expect(screen.getByText(/Sample data/)).toBeDefined();
    expect(screen.getByText(/entered Done this week/)).toBeDefined();
  });

  it("starts blank with a 'Save review' button when nothing is saved yet", async () => {
    renderShell(<WeeklyReviewScreen />, { route: "/reviews/weekly" });
    await screen.findByText("What were the highlights of this week?");
    expect(
      (screen.getByLabelText("What were the highlights of this week?") as HTMLTextAreaElement).value,
    ).toBe("");
    expect(screen.getByRole("button", { name: "Save review" })).toBeDefined();
  });

  it("saves an answer and keeps it for the same week", async () => {
    renderShell(<WeeklyReviewScreen />, { route: "/reviews/weekly" });
    const box = (await screen.findByLabelText(
      "What is the one priority for next week?",
    )) as HTMLTextAreaElement;
    await userEvent.type(box, "Ship the weekly review");
    await userEvent.click(screen.getByRole("button", { name: "Save review" }));
    await waitFor(() => expect(screen.getByText("Saved")).toBeDefined());
    expect(box.value).toBe("Ship the weekly review");
  });

  it("steps by ISO week and enables 'This week' off the current week", async () => {
    renderShell(<WeeklyReviewScreen />, { route: "/reviews/weekly" });
    await screen.findByText("What were the highlights of this week?");
    expect(screen.getByRole("button", { name: "This week" }).hasAttribute("disabled")).toBe(true);
    await userEvent.click(screen.getByRole("button", { name: "Previous week" }));
    await waitFor(() =>
      expect(screen.getByRole("button", { name: "This week" }).hasAttribute("disabled")).toBe(false),
    );
  });
});
