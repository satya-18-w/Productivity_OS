/**
 * Category colour assignment — PRESENTATION ONLY.
 *
 * D2 (design-system.md §3.1): category colour is a visual/semantic
 * identification aid and must NEVER drive business logic, validation,
 * filtering semantics, ordering, totals, or any stored meaning. This module
 * only maps an opaque key to one of the palette hues for display; the domain
 * never sees a colour.
 */

/** The category palette, as CSS custom-property references (see tokens.css). */
export const CATEGORY_PALETTE = [
  "var(--cat-personal)",
  "var(--cat-work)",
  "var(--cat-study)",
  "var(--cat-health)",
  "var(--cat-projects)",
  "var(--cat-ideas)",
  "var(--cat-finance)",
] as const;

/** The neutral hue for an unset / "Uncategorized" category. */
export const CATEGORY_UNSET = "var(--cat-other)";

function hash(key: string): number {
  let h = 0;
  for (let i = 0; i < key.length; i++) h = (Math.imul(31, h) + key.charCodeAt(i)) | 0;
  return Math.abs(h);
}

/**
 * Deterministically pick a palette hue for a category key (its id, ideally).
 * Stable for a given key; not meaningful beyond "tells two categories apart".
 * Pass `null`/`undefined` for the uncategorized bucket.
 */
export function categoryColor(key: string | null | undefined): string {
  if (!key) return CATEGORY_UNSET;
  return CATEGORY_PALETTE[hash(key) % CATEGORY_PALETTE.length];
}
