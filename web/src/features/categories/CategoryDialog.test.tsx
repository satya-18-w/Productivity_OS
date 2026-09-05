import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CategoryDialog } from "./CategoryDialog";
import { api, ApiError, type Category } from "../../api";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return { ...actual, api: { ...actual.api, createCategory: vi.fn(), renameCategory: vi.fn() } };
});

const EXISTING: Category = { id: "c9", name: "Admin" };

beforeEach(() => vi.clearAllMocks());

describe("CategoryDialog", () => {
  it("creates a category with the name only", async () => {
    vi.mocked(api.createCategory).mockResolvedValue({} as never);
    const onSaved = vi.fn();
    render(<CategoryDialog open target={{ mode: "new" }} onClose={() => {}} onSaved={onSaved} />);
    // no colour picker, no icon field — name only (§2)
    expect(screen.queryByLabelText(/colou?r|icon/i)).toBeNull();
    await userEvent.type(screen.getByLabelText(/Name/), "Deep Work");
    await userEvent.click(screen.getByRole("button", { name: "Create" }));
    expect(api.createCategory).toHaveBeenCalledWith("Deep Work");
    expect(onSaved).toHaveBeenCalled();
  });

  it("pre-fills and renames on edit", async () => {
    vi.mocked(api.renameCategory).mockResolvedValue(undefined);
    render(<CategoryDialog open target={{ mode: "rename", category: EXISTING }} onClose={() => {}} onSaved={vi.fn()} />);
    expect((screen.getByLabelText(/Name/) as HTMLInputElement).value).toBe("Admin");
    await userEvent.clear(screen.getByLabelText(/Name/));
    await userEvent.type(screen.getByLabelText(/Name/), "Administration");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(api.renameCategory).toHaveBeenCalledWith("c9", "Administration");
  });

  it("trims the name before creating (mirrors the backend trim)", async () => {
    vi.mocked(api.createCategory).mockResolvedValue({} as never);
    const onSaved = vi.fn();
    render(<CategoryDialog open target={{ mode: "new" }} onClose={() => {}} onSaved={onSaved} />);
    await userEvent.type(screen.getByLabelText(/Name/), "  Deep Work  ");
    await userEvent.click(screen.getByRole("button", { name: "Create" }));
    expect(api.createCategory).toHaveBeenCalledWith("Deep Work");
    expect(onSaved).toHaveBeenCalled();
  });

  it("disables Create while the name is blank or whitespace-only", async () => {
    render(<CategoryDialog open target={{ mode: "new" }} onClose={() => {}} onSaved={vi.fn()} />);
    const create = screen.getByRole("button", { name: "Create" });
    expect((create as HTMLButtonElement).disabled).toBe(true);
    await userEvent.type(screen.getByLabelText(/Name/), "   ");
    expect((create as HTMLButtonElement).disabled).toBe(true);
    await userEvent.type(screen.getByLabelText(/Name/), "Work");
    expect((create as HTMLButtonElement).disabled).toBe(false);
  });
  it("surfaces a duplicate-name (409) error", async () => {
    vi.mocked(api.createCategory).mockRejectedValue(new ApiError(409, "CONFLICT", "dup"));
    render(<CategoryDialog open target={{ mode: "new" }} onClose={() => {}} onSaved={vi.fn()} />);
    await userEvent.type(screen.getByLabelText(/Name/), "Admin");
    await userEvent.click(screen.getByRole("button", { name: "Create" }));
    expect(await screen.findByText("A category with this name already exists.")).toBeDefined();
  });
});
