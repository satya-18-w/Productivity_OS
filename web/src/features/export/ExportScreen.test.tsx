import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ExportScreen } from "./ExportScreen";
import { api } from "../../api";
import { renderShell } from "../../test/renderShell";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return {
    ...actual,
    api: {
      ...actual.api,
      getAccount: vi.fn(),
      listCategories: vi.fn(),
      board: vi.fn(),
      habits: vi.fn(),
      goals: vi.fn(),
    },
  };
});

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.getAccount).mockResolvedValue({ email: "sam@example.com", timezone: "UTC" });
  vi.mocked(api.listCategories).mockResolvedValue([]);
  vi.mocked(api.board).mockResolvedValue({ columns: [] });
  vi.mocked(api.habits).mockResolvedValue({ date: "2026-09-04", habits: [], archived: [] });
  vi.mocked(api.goals).mockResolvedValue([]);
});

describe("ExportScreen", () => {
  it("lists what's included and offers one export action", () => {
    renderShell(<ExportScreen />, { route: "/export" });
    expect(screen.getByRole("heading", { level: 1, name: "Export my data" })).toBeDefined();
    expect(screen.getByText(/Provisional format/)).toBeDefined();
    expect(screen.getByRole("button", { name: "Export my data" })).toBeDefined();
  });

  it("gathers every entity and offers a download on success", async () => {
    renderShell(<ExportScreen />, { route: "/export" });
    await userEvent.click(screen.getByRole("button", { name: "Export my data" }));
    await waitFor(() => expect(api.board).toHaveBeenCalled());
    expect(api.getAccount).toHaveBeenCalled();
    expect(api.listCategories).toHaveBeenCalled();
    expect(api.habits).toHaveBeenCalled();
    expect(api.goals).toHaveBeenCalled();
    expect(await screen.findByRole("link", { name: /Download productivity-os-export-/ })).toBeDefined();
  });

  it("announces failures without a download link", async () => {
    vi.mocked(api.board).mockRejectedValueOnce(new Error("boom"));
    renderShell(<ExportScreen />, { route: "/export" });
    await userEvent.click(screen.getByRole("button", { name: "Export my data" }));
    expect(await screen.findByRole("alert")).toBeDefined();
    expect(screen.queryByRole("link", { name: /Download/ })).toBeNull();
  });
});
