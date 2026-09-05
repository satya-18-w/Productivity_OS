import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { Login } from "./Login";
import { AuthContext } from "../auth";
import { api, ApiError } from "../api";

vi.mock("../api", async (io) => {
  const actual = await io<typeof import("../api")>();
  return { ...actual, api: { ...actual.api, login: vi.fn() } };
});

function renderLogin() {
  const refresh = vi.fn(async () => {});
  const setAccount = vi.fn();
  render(
    <MemoryRouter initialEntries={["/login"]}>
      <AuthContext.Provider value={{ account: null, loading: false, refresh, setAccount }}>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/" element={<h1>Home page</h1>} />
        </Routes>
      </AuthContext.Provider>
    </MemoryRouter>,
  );
  return { refresh };
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.login).mockResolvedValue(undefined);
});

describe("Login", () => {
  it("submits the credentials and redirects home", async () => {
    const { refresh } = renderLogin();
    await userEvent.type(screen.getByLabelText("Email *(required)"), "sam@example.com");
    await userEvent.type(screen.getByLabelText("Password *(required)"), "Secret!1");
    await userEvent.click(screen.getByRole("button", { name: "Log in" }));
    await waitFor(() => expect(api.login).toHaveBeenCalledWith("sam@example.com", "Secret!1"));
    expect(refresh).toHaveBeenCalled();
    expect(await screen.findByRole("heading", { name: "Home page" })).toBeDefined();
  });

  it("shows the wrong-password message without redirecting", async () => {
    vi.mocked(api.login).mockRejectedValue(new ApiError(401, "INVALID_CREDENTIALS", "Nope"));
    renderLogin();
    await userEvent.type(screen.getByLabelText("Email *(required)"), "sam@example.com");
    await userEvent.type(screen.getByLabelText("Password *(required)"), "Wrong!1");
    await userEvent.click(screen.getByRole("button", { name: "Log in" }));
    expect(await screen.findByText("Incorrect email or password.")).toBeDefined();
    expect(screen.queryByRole("heading", { name: "Home page" })).toBeNull();
  });

  it("shows the rate-limit message on 429", async () => {
    vi.mocked(api.login).mockRejectedValue(new ApiError(429, "RATE_LIMITED", "Slow down"));
    renderLogin();
    await userEvent.type(screen.getByLabelText("Email *(required)"), "sam@example.com");
    await userEvent.type(screen.getByLabelText("Password *(required)"), "Secret!1");
    await userEvent.click(screen.getByRole("button", { name: "Log in" }));
    expect(await screen.findByText("Too many attempts. Please wait a few minutes.")).toBeDefined();
  });
});
