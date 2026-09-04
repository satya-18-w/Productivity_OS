import { describe, it, expect, vi } from "vitest";
import { useState } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Dialog } from "./Dialog";
import { Button } from "./Button";

function Harness() {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button onClick={() => setOpen(true)}>Open</Button>
      <Dialog open={open} onClose={() => setOpen(false)} title="Confirm">
        <p>Body text</p>
      </Dialog>
    </>
  );
}

describe("Dialog", () => {
  it("opens on demand and labels itself by its title", async () => {
    render(<Harness />);
    await userEvent.click(screen.getByRole("button", { name: "Open" }));
    const dialog = screen.getByRole("dialog");
    expect((dialog as HTMLDialogElement).open).toBe(true);
    // aria-labelledby points at the rendered title
    const labelId = dialog.getAttribute("aria-labelledby");
    expect(document.getElementById(labelId!)?.textContent).toBe("Confirm");
  });

  it("closes via the close button", async () => {
    render(<Harness />);
    await userEvent.click(screen.getByRole("button", { name: "Open" }));
    await userEvent.click(screen.getByRole("button", { name: "Close" }));
    expect((screen.getByRole("dialog", { hidden: true }) as HTMLDialogElement).open).toBe(false);
  });

  it("calls onClose when cancelled (Esc)", async () => {
    const onClose = vi.fn();
    render(
      <Dialog open onClose={onClose} title="T">
        <p>x</p>
      </Dialog>,
    );
    screen.getByRole("dialog").dispatchEvent(new Event("cancel", { cancelable: true }));
    expect(onClose).toHaveBeenCalled();
  });
});
