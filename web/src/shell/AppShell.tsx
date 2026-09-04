import { Outlet } from "react-router-dom";
import { Sidebar, SidebarDrawer } from "./Sidebar";
import { MobileTopBar } from "./MobileTopBar";
import { useShellState } from "./useShellState";

/**
 * The three-region application shell (D3): sidebar · main · (per-screen right
 * rail, provided by each screen via <ScreenLayout>). Used as the layout route
 * element for every authenticated screen. Replaces <AuthLayout>.
 */
export function AppShell() {
  const { mode, drawerOpen, openDrawer, closeDrawer } = useShellState();

  return (
    <div className="app-shell" data-shell-mode={mode}>
      <a className="skip-link" href="#main">
        Skip to content
      </a>

      {mode === "drawer" ? (
        <>
          <MobileTopBar onOpenMenu={openDrawer} />
          <SidebarDrawer open={drawerOpen} onClose={closeDrawer} />
        </>
      ) : (
        <Sidebar mode={mode} />
      )}

      <main id="main">
        <Outlet />
      </main>
    </div>
  );
}
