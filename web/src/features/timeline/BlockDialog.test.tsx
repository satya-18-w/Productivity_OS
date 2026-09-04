import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { BlockDialog } from "./BlockDialog";
import { api, ApiError, type Category, type PositionedBlock } from "../../api";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return {
    ...actual,
    api: {
      ...actual.api,
      createBlock: vi.fn(),
      updateBlock: vi.fn(),
      deleteBlock: vi.fn(),
    },
  };
});

const CATS: Category[] = [
  { id: "c1", name: "Deep Work" },
  { id: "c2", name: "Admin" },
];

const EXISTING: PositionedBlock = {
  id: "b9",
  kind: "actual",
  starts_at: "",
  ends_at: "",
  category_id: "c2",
  category_name: "Admin",
  start_minute: 600,
  end_minute: 720,
  from_prev_day: false,
  to_next_day: false,
  local_date: "2026-09-04",
  local_start: "10:00",
  local_end: "12:00",
  ends_next_day: false,
};

beforeEach(() => vi.clearAllMocks());

describe("BlockDialog", () => {
  it("creates a block from the form (planned/actual selectable)", async () => {
    vi.mocked(api.createBlock).mockResolvedValue({} as never);
    const onSaved = vi.fn();
    render(
      <BlockDialog open target={{ mode: "new" }} date="2026-09-04" categories={CATS} onClose={() => {}} onSaved={onSaved} />,
    );
    expect(screen.getByRole("radiogroup", { name: "Block type" })).toBeDefined();
    await userEvent.click(screen.getByRole("radio", { name: "Actual" }));
    await userEvent.click(screen.getByRole("button", { name: "Add" }));
    expect(api.createBlock).toHaveBeenCalledWith(
      expect.objectContaining({ kind: "actual", date: "2026-09-04", start: "09:00", end: "10:00" }),
    );
    expect(onSaved).toHaveBeenCalled();
  });

  it("pre-fills when editing and locks the kind", () => {
    render(
      <BlockDialog open target={{ mode: "edit", block: EXISTING }} date="2026-09-04" categories={CATS} onClose={() => {}} onSaved={() => {}} />,
    );
    expect(screen.getByRole("heading", { name: "Edit block" })).toBeDefined();
    expect(screen.queryByRole("radiogroup")).toBeNull();
    expect(screen.getByText("Actual block")).toBeDefined();
    expect((screen.getByLabelText("Start") as HTMLInputElement).value).toBe("10:00");
  });

  it("calls updateBlock on save when editing", async () => {
    vi.mocked(api.updateBlock).mockResolvedValue(undefined);
    render(
      <BlockDialog open target={{ mode: "edit", block: EXISTING }} date="2026-09-04" categories={CATS} onClose={() => {}} onSaved={vi.fn()} />,
    );
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(api.updateBlock).toHaveBeenCalledWith("b9", expect.objectContaining({ kind: "actual" }));
  });

  it("deletes the block", async () => {
    vi.mocked(api.deleteBlock).mockResolvedValue(undefined);
    const onSaved = vi.fn();
    render(
      <BlockDialog open target={{ mode: "edit", block: EXISTING }} date="2026-09-04" categories={CATS} onClose={() => {}} onSaved={onSaved} />,
    );
    await userEvent.click(screen.getByRole("button", { name: "Delete" }));
    expect(api.deleteBlock).toHaveBeenCalledWith("b9");
    expect(onSaved).toHaveBeenCalled();
  });

  it("surfaces a field-level validation error", async () => {
    vi.mocked(api.createBlock).mockRejectedValue(
      new ApiError(400, "VALIDATION_ERROR", "bad", { end: "End must be after start." }),
    );
    render(
      <BlockDialog open target={{ mode: "new" }} date="2026-09-04" categories={CATS} onClose={() => {}} onSaved={vi.fn()} />,
    );
    await userEvent.click(screen.getByRole("button", { name: "Add" }));
    expect(await screen.findByText("End must be after start.")).toBeDefined();
  });
});
