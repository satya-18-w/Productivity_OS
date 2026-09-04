import { vi } from "vitest";

/**
 * Replace window.matchMedia so `(max-width: Npx)` queries resolve against a
 * fixed viewport width. Call in a test; the setup file restores a no-op stub
 * between tests via cleanup of the module, so re-set per test as needed.
 */
export function mockViewport(width: number): void {
  window.matchMedia = vi.fn((query: string) => {
    const m = /\(max-width:\s*([\d.]+)px\)/.exec(query);
    const min = /\(min-width:\s*([\d.]+)px\)/.exec(query);
    let matches = false;
    if (m) matches = width <= parseFloat(m[1]);
    else if (min) matches = width >= parseFloat(min[1]);
    return {
      matches,
      media: query,
      onchange: null,
      addEventListener: () => {},
      removeEventListener: () => {},
      addListener: () => {},
      removeListener: () => {},
      dispatchEvent: () => false,
    } as unknown as MediaQueryList;
  });
}
