/**
 * Responsive breakpoint scale.
 *
 * design-system.md §3.7: the responsive *shed order* is ratified (D4) but the
 * exact pixel thresholds are PROVISIONAL — pending the T1 token-extraction
 * pass. Treat these numbers as indicative, not canonical.
 *
 * Shed order (D4), as width decreases:
 *   1. remove / reduce the right contextual rail   (< wide)
 *   2. collapse the sidebar to icons               (< laptop)
 *   3. sidebar becomes a drawer / nav pattern      (< tablet → mobile)
 * The main column's content and its one primary action survive at every size,
 * and the page never scrolls sideways (see base.css `body { overflow-x: clip }`).
 *
 * CSS cannot read custom properties inside `@media`, so these live here (for JS)
 * and are mirrored as literal values in feature CSS media queries.
 */
export const breakpoints = {
  /** below this: sidebar → drawer, single column */
  tablet: 640,
  /** below this: sidebar labels collapse to icons */
  laptop: 1024,
  /** at/above this: sidebar + main + right rail all visible */
  wide: 1280,
} as const;

export type BreakpointName = keyof typeof breakpoints;

/** `@media` string that matches at or above the named breakpoint. */
export const up = (name: BreakpointName): string =>
  `(min-width: ${breakpoints[name]}px)`;

/** `@media` string that matches below the named breakpoint. */
export const down = (name: BreakpointName): string =>
  `(max-width: ${breakpoints[name] - 0.02}px)`;
