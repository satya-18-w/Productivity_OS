import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { Register } from "./Register";
import { AuthContext } from "../auth";
import { api, ApiError, browserTimezone } from "../api";

vi.mock("../api", async (io) => {
  const actual = await io<typeof import("../api")>();
  return { ...actual, api: { ...actual.api, register: vi.fn() } };
});

function renderRegister() {
  const refresh = vi.fn(async () => {});
  const setAccount = vi.fn();
  render(
    <MemoryRouter initialEntries={["/register"]}>
      <AuthContext.Provider value={{ account: null, loading: false, refresh, setAccount }}>
        <Routes>
          <Route path="/register" element={<Register />} />
          <Route path="/" element={<h1>Home page</h1>} />
        </Routes>
      </AuthContext.Provider>
    </MemoryRouter>,
  );
  return { refresh };
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.register).mockResolvedValue({ email: "sam@example.com", timezone: "UTC" });
});

describe("Register", () => {
  it("shows the Q6 hint and defaults the timezone to the browser zone (Q4)", () => {
    renderRegister();
    expect(screen.getByText(/At least 6 characters/).textContent).toMatch(/special character/);
    expect((screen.getByLabelText("Timezone *(required)") as HTMLSelectElement).value).toBe(browserTimezone());
  });

  it("creates the account and redirects home", async () => {
    const { refresh } = renderRegister();
    await userEvent.type(screen.getByLabelText("Email *(required)"), "sam@example.com");
    await userEvent.type(screen.getByLabelText("Password *(required)"), "Secret!1");
    await userEvent.click(screen.getByRole("button", { name: "Create account" }));
    await waitFor(() =>
      expect(api.register).toHaveBeenCalledWith("sam@example.com", "Secret!1", browserTimezone()),
    );
    expect(refresh).toHaveBeenCalled();
    expect(await screen.findByRole("heading", { name: "Home page" })).toBeDefined();
  });

  it("maps server field errors onto the fields", async () => {
    vi.mocked(api.register).mockRejectedValue(
      new ApiError(400, "VALIDATION_ERROR", "Invalid", { password: "Too weak." }),
    );
    renderRegister();
    await userEvent.type(screen.getByLabelText("Email *(required)"), "sam@example.com");
    await userEvent.type(screen.getByLabelText("Password *(required)"), "weak");
    await userEvent.click(screen.getByRole("button", { name: "Create account" }));
    expect(await screen.findByText("Too weak.")).toBeDefined();
  });

  it("reports an already-registered email", async () => {
    vi.mocked(api.register).mockRejectedValue(
      new ApiError(409, "EMAIL_ALREADY_REGISTERED", "Taken"),
    );
    renderRegister();
    await userEvent.type(screen.getByLabelText("Email *(required)"), "sam@example.com");
    await userEvent.type(screen.getByLabelText("Password *(required)"), "Secret!1");
    await userEvent.click(screen.getByRole("button", { name: "Create account" }));
    expect(await screen.findByText("An account with this email already exists.")).toBeDefined();
  });
});
