import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CategoryRow } from "./CategoryRow";
import type { Category } from "../../api";

const CAT: Category = { id: "c1", name: "Deep Work" };

describe("CategoryRow", () => {
  it("shows the name and a decorative colour tile", () => {
    render(<CategoryRow category={CAT} onRename={vi.fn()} onArchive={vi.fn()} />);
    expect(screen.getByText("Deep Work")).toBeDefined();
    expect(screen.getByRole("img", { name: "Deep Work" })).toBeDefined();
  });

  it("kebab offers Rename and Archive only", async () => {
    render(<CategoryRow category={CAT} onRename={vi.fn()} onArchive={vi.fn()} />);
    await userEvent.click(screen.getByRole("button", { name: /Actions for Deep Work/ }));
    expect(screen.getAllByRole("menuitem").map((m) => m.textContent)).toEqual(["Rename", "Archive"]);
  });

  it("rename and archive call their handlers", async () => {
    const onRename = vi.fn();
    const onArchive = vi.fn();
    render(<CategoryRow category={CAT} onRename={onRename} onArchive={onArchive} />);
    await userEvent.click(screen.getByRole("button", { name: /Actions for Deep Work/ }));
    await userEvent.click(screen.getByRole("menuitem", { name: "Rename" }));
    expect(onRename).toHaveBeenCalledWith(CAT);

    await userEvent.click(screen.getByRole("button", { name: /Actions for Deep Work/ }));
    await userEvent.click(screen.getByRole("menuitem", { name: "Archive" }));
    expect(onArchive).toHaveBeenCalledWith(CAT);
  });
});
