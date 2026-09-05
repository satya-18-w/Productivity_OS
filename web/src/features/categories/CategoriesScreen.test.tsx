import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CategoriesScreen } from "./CategoriesScreen";
import { api, type Category } from "../../api";
import { renderShell } from "../../test/renderShell";

vi.mock("../../api", async (io) => {
  const actual = await io<typeof import("../../api")>();
  return { ...actual, api: { ...actual.api, listCategories: vi.fn(), archiveCategory: vi.fn() } };
});

const CATS: Category[] = [
  { id: "c1", name: "Deep Work" },
  { id: "c2", name: "Admin" },
];

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.listCategories).mockResolvedValue(CATS);
  vi.mocked(api.archiveCategory).mockResolvedValue(undefined);
});

describe("CategoriesScreen", () => {
  it("lists active categories with no item counts or breakdowns", async () => {
    const { container } = renderShell(<CategoriesScreen />, { route: "/categories" });
    await waitFor(() => expect(api.listCategories).toHaveBeenCalled());
    expect(screen.getByText("Deep Work")).toBeDefined();
    expect(screen.getByText("Admin")).toBeDefined();
    expect(container.textContent).not.toMatch(/\d+\s*(items|tasks|habits|goals)/i);
  });

  it("archives a category after confirmation", async () => {
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(true);
    renderShell(<CategoriesScreen />, { route: "/categories" });
    await waitFor(() => expect(api.listCategories).toHaveBeenCalled());
    await userEvent.click(screen.getByRole("button", { name: /Actions for Deep Work/ }));
    await userEvent.click(screen.getByRole("menuitem", { name: "Archive" }));
    expect(confirmSpy).toHaveBeenCalled();
    expect(api.archiveCategory).toHaveBeenCalledWith("c1");
    confirmSpy.mockRestore();
  });

  it("does not archive when the confirmation is declined", async () => {
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(false);
    renderShell(<CategoriesScreen />, { route: "/categories" });
    await waitFor(() => expect(api.listCategories).toHaveBeenCalled());
    await userEvent.click(screen.getByRole("button", { name: /Actions for Admin/ }));
    await userEvent.click(screen.getByRole("menuitem", { name: "Archive" }));
    expect(api.archiveCategory).not.toHaveBeenCalled();
    confirmSpy.mockRestore();
  });

  it("opens the new-category dialog", async () => {
    renderShell(<CategoriesScreen />, { route: "/categories" });
    await waitFor(() => expect(api.listCategories).toHaveBeenCalled());
    await userEvent.click(screen.getByRole("button", { name: "New category" }));
    expect(screen.getByRole("dialog", { name: "New category" })).toBeDefined();
  });

  it("shows an empty state when there are no categories", async () => {
    vi.mocked(api.listCategories).mockResolvedValueOnce([]);
    renderShell(<CategoriesScreen />, { route: "/categories" });
    expect(await screen.findByText("No categories yet")).toBeDefined();
  });

  it("shows an error state with retry on load failure", async () => {
    vi.mocked(api.listCategories).mockRejectedValueOnce(new Error("x"));
    renderShell(<CategoriesScreen />, { route: "/categories" });
    expect(await screen.findByText("Could not load your categories.")).toBeDefined();
  });
});
