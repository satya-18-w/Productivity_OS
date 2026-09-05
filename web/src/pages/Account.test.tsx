import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { Account } from "./Account";
import { AuthContext } from "../auth";
import { api, ApiError, type Account as AccountShape } from "../api";

vi.mock("../api", async (io) => {
  const actual = await io<typeof import("../api")>();
  return {
    ...actual,
    api: { ...actual.api, setTimezone: vi.fn(), changePassword: vi.fn(), logout: vi.fn() },
  };
});

const ACCOUNT: AccountShape = { email: "sam@example.com", timezone: "UTC" };

function renderAccount(account: AccountShape | null = ACCOUNT) {
  const setAccount = vi.fn();
  const refresh = vi.fn(async () => {});
  render(
    <MemoryRouter initialEntries={["/account"]}>
      <AuthContext.Provider value={{ account, loading: false, refresh, setAccount }}>
        <Routes>
          <Route path="/account" element={<Account />} />
          <Route path="/login" element={<h1>Login page</h1>} />
        </Routes>
      </AuthContext.Provider>
    </MemoryRouter>,
  );
  return { setAccount, refresh };
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.setTimezone).mockResolvedValue(undefined);
  vi.mocked(api.changePassword).mockResolvedValue(undefined);
  vi.mocked(api.logout).mockResolvedValue(undefined);
});

describe("Account", () => {
  it("shows the header and email with no profile fields", () => {
    renderAccount();
    expect(screen.getByRole("heading", { level: 1, name: "Account" })).toBeDefined();
    expect(screen.getByText("sam@example.com")).toBeDefined();
    // §1 boundary: no profile fields beyond email/password/timezone.
    expect(screen.queryByLabelText(/first name/i)).toBeNull();
    expect(screen.queryByLabelText(/last name/i)).toBeNull();
    expect(screen.queryByText(/delete account/i)).toBeNull();
  });

  it("saves the timezone from the IANA select", async () => {
    const { refresh } = renderAccount();
    await userEvent.selectOptions(screen.getByLabelText("IANA timezone"), "UTC");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(api.setTimezone).toHaveBeenCalledWith("UTC"));
    expect(refresh).toHaveBeenCalled();
    expect(await screen.findByText("Saved.")).toBeDefined();
  });

  it("blocks a password change when the confirmation does not match", async () => {
    renderAccount();
    await userEvent.type(screen.getByLabelText("Current password *(required)"), "Oldpass!1");
    await userEvent.type(screen.getByLabelText("New password *(required)"), "Newpass!1");
    await userEvent.type(screen.getByLabelText("Confirm new password *(required)"), "Different!1");
    await userEvent.click(screen.getByRole("button", { name: "Change password" }));
    expect(await screen.findByText("Passwords do not match.")).toBeDefined();
    expect(api.changePassword).not.toHaveBeenCalled();
  });

  it("surfaces the Q6 server policy error on the new-password field", async () => {
    vi.mocked(api.changePassword).mockRejectedValue(
      new ApiError(400, "VALIDATION_ERROR", "Invalid", { new_password: "Too weak." }),
    );
    renderAccount();
    await userEvent.type(screen.getByLabelText("Current password *(required)"), "Oldpass!1");
    await userEvent.type(screen.getByLabelText("New password *(required)"), "weak");
    await userEvent.type(screen.getByLabelText("Confirm new password *(required)"), "weak");
    await userEvent.click(screen.getByRole("button", { name: "Change password" }));
    expect(await screen.findByText("Too weak.")).toBeDefined();
  });

  it("logs out, clears the session, and lands on /login", async () => {
    const { setAccount } = renderAccount();
    await userEvent.click(screen.getByRole("button", { name: "Log out" }));
    await waitFor(() => expect(api.logout).toHaveBeenCalled());
    expect(setAccount).toHaveBeenCalledWith(null);
    expect(await screen.findByRole("heading", { name: "Login page" })).toBeDefined();
  });
});
