import type { ReactElement, ReactNode } from "react";
import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { AuthContext } from "../auth";
import type { Account } from "../api";

const TEST_ACCOUNT: Account = { email: "sam@example.com", timezone: "UTC" };

export function AuthStub({ children, account = TEST_ACCOUNT }: { children: ReactNode; account?: Account | null }) {
  return (
    <AuthContext.Provider
      value={{ account, loading: false, refresh: async () => {}, setAccount: () => {} }}
    >
      {children}
    </AuthContext.Provider>
  );
}

/** Render a shell component with a router + a synthetic authenticated session. */
export function renderShell(ui: ReactElement, { route = "/timeline" }: { route?: string } = {}) {
  return render(
    <MemoryRouter initialEntries={[route]}>
      <AuthStub>{ui}</AuthStub>
    </MemoryRouter>,
  );
}
