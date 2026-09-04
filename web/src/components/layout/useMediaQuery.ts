import { useSyncExternalStore } from "react";
import { breakpoints, down, up, type BreakpointName } from "../../styles/breakpoints";

/**
 * Subscribe to a media query. SSR-safe default is `false`.
 * Foundation infrastructure for the responsive shell (D3, pending) — see
 * styles/breakpoints.ts for the ratified shed order (D4).
 */
export function useMediaQuery(query: string): boolean {
  return useSyncExternalStore(
    (onChange) => {
      if (typeof window === "undefined" || !window.matchMedia) return () => {};
      const mql = window.matchMedia(query);
      mql.addEventListener("change", onChange);
      return () => mql.removeEventListener("change", onChange);
    },
    () =>
      typeof window !== "undefined" && window.matchMedia
        ? window.matchMedia(query).matches
        : false,
    () => false,
  );
}

/** True at or above the named breakpoint. */
export const useBreakpointUp = (name: BreakpointName) => useMediaQuery(up(name));
/** True below the named breakpoint. */
export const useBreakpointDown = (name: BreakpointName) => useMediaQuery(down(name));

export { breakpoints };
