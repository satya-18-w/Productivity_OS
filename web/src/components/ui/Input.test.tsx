import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Input } from "./Input";
import { Textarea } from "./Textarea";
import { Field } from "./Field";

describe("Input + Field", () => {
  it("links label to control and accepts typing", async () => {
    render(
      <Field label="Title" htmlFor="title">
        <Input id="title" />
      </Field>,
    );
    const input = screen.getByLabelText("Title");
    await userEvent.type(input, "Plan the week");
    expect((input as HTMLInputElement).value).toBe("Plan the week");
  });

  it("marks the control invalid and shows the error", () => {
    render(
      <Field label="Email" htmlFor="email" error="Required">
        <Input id="email" invalid />
      </Field>,
    );
    expect(screen.getByLabelText("Email").getAttribute("aria-invalid")).toBe("true");
    expect(screen.getByRole("alert").textContent).toBe("Required");
  });

  it("renders the required marker accessibly", () => {
    render(
      <Field label="Name" htmlFor="name" required>
        <Input id="name" />
      </Field>,
    );
    expect(screen.getByText("(required)")).toBeDefined();
  });

  it("Textarea forwards value changes", async () => {
    render(
      <Field label="Notes" htmlFor="notes">
        <Textarea id="notes" />
      </Field>,
    );
    const ta = screen.getByLabelText("Notes");
    await userEvent.type(ta, "hello");
    expect((ta as HTMLTextAreaElement).value).toBe("hello");
  });
});
