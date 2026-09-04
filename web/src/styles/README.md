# Frontend styling — how it fits together

There is **one** visual system. Its source of truth is
[`docs/design/design-system.md`](../../../docs/design/design-system.md) (tokens and
component contracts) and [`visual-principles.md`](../../../docs/design/visual-principles.md)
(judgement). This folder is that system's CSS expression.

## Files

| File | Owns |
|---|---|
| `tokens.css` | Every design token as a CSS custom property (colour, type, spacing, radius, elevation, layout). The **canonical** token layer. |
| `base.css` | Element reset, base typography, `:focus-visible`, reduced-motion, the sideways-scroll guard. |
| `../styles.css` | **Legacy** feature/component CSS (`.card`, `.btn`, `.nav`, `.tl-*`, `.board-*`, `.habit-*`, `.goal-*`, `.progress-*`) still used by existing pages. Being migrated to `ui-` primitives screen by screen. Do not add rules here. |
| `primitives.css` | Styles for the design-system components in `src/components/**` (`.ui-*`). |
| `breakpoints.ts` | Breakpoint scale + `@media` string helpers for JS (CSS can't read custom props in `@media`). |
| `index.css` | Import orchestrator — the only stylesheet `main.tsx` imports. Order: tokens → base → legacy → primitives. |

## Rules

- Components and feature CSS consume **tokens**, never raw values. No one-off colours,
  spacing, radii, shadows or type steps — if something is genuinely missing, add a token
  here and get it approved (project `CLAUDE.md` → "Design System Changes").
- Class prefix `ui-` marks the design-system layer. It is not a second system — the
  tokens are the system; these are just its selectors.
- Layout uses Grid / Flexbox / normal flow. No absolute positioning for page structure.
- Full light/dark parity: define a colour once on bare `:root`, override under the dark
  blocks. Never give a colour its only definition inside a media query.

## Not yet ratified — see `docs/design/design-system.md` §6.2

- **T1** — every hex value in `tokens.css`, and the px thresholds in `breakpoints.ts`,
  are the *direction* values from design-system.md §3, used as the closest temporary
  implementation. They are marked `PROVISIONAL` and are **not** the final canonical
  values. A token-extraction pass replaces them.
- **D3** — `--sidebar-w` / `--rail-w` are indicative only; the three-region app shell is
  not approved, so nothing is built against it yet.
- **Dark palette** — provisional; extracted alongside the light palette in T1.
