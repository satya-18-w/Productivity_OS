/**
 * Theme preference — light / dark / system.
 *
 * `system` removes the `data-theme` attribute so `tokens.css`'s
 * `prefers-color-scheme` media block takes over. `light` / `dark` stamp the
 * attribute and win in both directions.
 *
 * Persisted in localStorage. `initTheme()` is called once, as early as
 * possible, before React renders (see main.tsx).
 */
export type ThemePreference = "light" | "dark" | "system";

const STORAGE_KEY = "pos-theme";

export function getThemePreference(): ThemePreference {
  try {
    const v = localStorage.getItem(STORAGE_KEY);
    if (v === "light" || v === "dark" || v === "system") return v;
  } catch {
    /* private mode / disabled storage — fall through */
  }
  return "system";
}

export function applyTheme(pref: ThemePreference): void {
  const root = document.documentElement;
  if (pref === "system") root.removeAttribute("data-theme");
  else root.setAttribute("data-theme", pref);
}

export function setThemePreference(pref: ThemePreference): void {
  try {
    if (pref === "system") localStorage.removeItem(STORAGE_KEY);
    else localStorage.setItem(STORAGE_KEY, pref);
  } catch {
    /* ignore */
  }
  applyTheme(pref);
}

/** Apply the stored preference. Call once at startup. */
export function initTheme(): void {
  applyTheme(getThemePreference());
}
