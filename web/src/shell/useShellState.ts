import { useEffect, useState } from "react";
import { useMediaQuery } from "../components/layout/useMediaQuery";
import { breakpoints } from "../styles/breakpoints";

export type ShellMode = "expanded" | "collapsed" | "drawer";

export interface ShellState {
  /** Layout mode, derived from viewport width (D4 shed order). */
  mode: ShellMode;
  /** Drawer open (only meaningful when mode === "drawer"). */
  drawerOpen: boolean;
  openDrawer: () => void;
  closeDrawer: () => void;
}

/**
 * Shell layout state. D4 shed order:
 *   >= laptop            → "expanded" (full sidebar; rail beside main >= wide)
 *   >= tablet, < laptop  → "collapsed" (icon-only sidebar)
 *   < tablet             → "drawer" (off-canvas sidebar + mobile top bar)
 */
export function useShellState(): ShellState {
  const belowLaptop = useMediaQuery(`(max-width: ${breakpoints.laptop - 0.02}px)`);
  const belowTablet = useMediaQuery(`(max-width: ${breakpoints.tablet - 0.02}px)`);
  const [drawerOpen, setDrawerOpen] = useState(false);

  const mode: ShellMode = belowTablet ? "drawer" : belowLaptop ? "collapsed" : "expanded";

  // Leaving drawer mode (e.g. rotate / resize up) must not strand an open drawer.
  useEffect(() => {
    if (mode !== "drawer" && drawerOpen) setDrawerOpen(false);
  }, [mode, drawerOpen]);

  return {
    mode,
    drawerOpen,
    openDrawer: () => setDrawerOpen(true),
    closeDrawer: () => setDrawerOpen(false),
  };
}
