# App Shell + Routing — SPEC & PLAN  (Phase 1)

> **Combines plan stages 1 (App Shell, D3) and 2 (Routing, D10)** — they are one unit: a
> shell with nav links needs routes, and routes need a shell to render into.
> Governing decisions: D3, D10 (`design-system.md §6.1`), D4 (shed order), D6 (no
> motivational surfaces), and the exclusion list (`design-system.md §6.4`).
>
> **Not a `v1.md` feature** — this is architecture. SPEC + PLAN here stand in for a
> `docs/specs/` entry; approve this document before implementing.

---

## SPEC — what Phase 1 delivers

A persistent three-region application shell that every authenticated screen renders into,
plus the full route table with a placeholder screen at each not-yet-built route.

### Regions

```
┌───────────┬─────────────────────────────────────┐
│           │  main  (#main, <main>)               │
│  Sidebar  │  ┌───────────────────┬────────────┐  │
│  <aside>  │  │ screen content    │ right rail │  │  ← rail is per-screen,
│           │  │ (PageHeader+body) │ (optional) │  │    optional, ≥ wide only
│           │  └───────────────────┴────────────┘  │
└───────────┴─────────────────────────────────────┘
```

- **Sidebar** (`<aside>`, full height, own scroll):
  - **Brand** — leaf glyph in a `--brand` rounded tile + "Productivity OS". Links to `/`.
  - **Primary nav** (`<nav aria-label="Primary">`) — `NavLink` per destination: line icon +
    label, `--radius-sm` full-width hit area. States: default (muted) / hover
    (`--surface-hover`) / **active** (`--brand-soft` pill, `--brand` icon+text,
    `aria-current="page"`). Optional trailing count `Badge` (data wired later; omit for now).
  - **Footer** (pinned bottom) — `ThemeToggle` + `UserMenu`.
  - **No "Spaces" section** (C1). **No motivational card** (D6 — the slot stays empty;
    a purely decorative brandmark is allowed later, no text, no progress bar).
- **Main** (`<main id="main">`) — renders `<Outlet/>`. A screen wraps its content in
  `<ScreenLayout rail={…}>`.
- **Right rail** (`<aside aria-label="…">` supplied by the screen) — contextual widgets.
  Optional. `≥ wide`: beside main. `< wide`: **stacks below** main content (D4 — rail
  sheds first). Never a horizontal-scroll cause.

### Primary nav destinations (V1 — no dashboard/notes/calendar/analytics)

| Label | Icon | Route |
|---|---|---|
| Timeline | calendar-clock | `/timeline` (`/` redirects here) |
| Tasks | check-square | `/tasks` |
| Board | columns | `/board` |
| Habits | repeat | `/habits` |
| Goals | target | `/goals` |
| Categories | tag | `/categories` |
| Reports | bar-chart | `/reports` |
| Reviews | notebook | `/reviews/daily` |

**Utility** (shown in the sidebar footer / `UserMenu`, not the primary list):
- Account → `/account`
- Export data → `/export`
- Log out (existing behaviour)

### Route table (D10)

| Route | Element (Phase 1) | Later |
|---|---|---|
| `/` | redirect → `/timeline` | — |
| `/timeline` | `<Placeholder name="Timeline" />` | Phase 3–4 |
| `/tasks` | `<Placeholder name="Tasks" />` | Phase 5 |
| `/board` | `<Placeholder name="Board" />` | Phase 6 |
| `/habits` | `<Placeholder name="Habits" />` | Phase 7 |
| `/goals` | `<Placeholder name="Goals" />` | Phase 8 |
| `/categories` | `<Placeholder name="Categories" />` | Phase 9 |
| `/reports` | `<Placeholder name="Reports" />` | Phase 10 |
| `/reviews/daily` · `/reviews/weekly` | `<Placeholder name="… review" />` | Phase 11–12 |
| `/account` | existing `Account` page, wrapped in shell | Phase 13 (restyle) |
| `/export` | `<Placeholder name="Export" />` | Phase 15 |
| `/login` · `/register` | existing pages, **no shell** | Phase 14 (restyle) |
| `*` (unknown) | redirect → `/` | — |

- Auth guard unchanged: unauthenticated → `/login`; authenticated on `/login|/register` → `/`.
- Old routes removed: `/board` stays; `/` no longer renders `Timeline` component directly
  (redirects); no `/dashboard`, `/notes`, `/calendar`, `/timeline/week|month`.

### Theme toggle

- 3-way: **Light / Dark / System**. Persists to `localStorage` (`pos-theme`).
- Applies `data-theme="light|dark"` on `<html>`; **System** removes the attribute
  (tokens.css already handles `prefers-color-scheme`).
- `ThemeToggle` = a small `SegmentedControl` or 3 `IconButton`s (sun / moon / monitor) in
  the sidebar footer. Init runs before first paint (inline in `index.html` head is
  ideal; acceptable as the first line of `main.tsx` for now — note any FOUC).

### Responsive (D4 shed order)

| Width | Sidebar | Rail | Extra |
|---|---|---|---|
| `≥ wide` (1280) | full (248px, labels) | beside main | — |
| `≥ laptop` (1024) | full | **stacked below main** | — |
| `≥ tablet` (640) | **icons only** (56px); label via `Tooltip` + accessible name | stacked below | — |
| `< tablet` | **drawer** (slide-over, focus-trapped, Esc/backdrop closes) | stacked below | slim top bar: hamburger `IconButton` + brand |

- Main content + the screen's one primary action are visible at every width.
- Page never scrolls sideways (`body { overflow-x: clip }` already in `base.css`; verify per screen).

### Accessibility

- Skip link (`<a href="#main">Skip to content</a>`) first in DOM, visible on focus.
- `<nav aria-label="Primary">`, `<main id="main">`, rail `<aside aria-label="<screen> details">`.
- Active nav item: `aria-current="page"`.
- Icon-only sidebar: every item keeps its full text as accessible name (visible label
  hidden with `.ui-visually-hidden`, or `aria-label`) + `Tooltip` on hover/focus.
- Drawer: `role="dialog" aria-modal="true"`, focus moves in on open, returns to the
  hamburger on close, `Esc` closes, background not focusable while open.
- `ThemeToggle`: a labelled `radiogroup` ("Theme"); each option a real control.
- `UserMenu`: standard menu button pattern (`aria-haspopup`, `aria-expanded`, arrow keys, `Esc`).
- Focus-visible rings on everything (inherited from `base.css`).

### New design tokens (additive — recorded in `design-system.md §3.6`)

| Token | Provisional value | Role |
|---|---|---|
| `--sidebar-w` | 248px *(exists)* | expanded sidebar |
| `--sidebar-w-collapsed` | 56px | icon-only sidebar |
| `--rail-w` | 320px *(exists)* | right rail |
| `--topbar-h` | 52px | mobile top bar |
| `--z-sidebar` | 30 | |
| `--z-drawer` | 50 | drawer + backdrop |

### Out of scope for Phase 1

Global search, notification bell, "Spaces", motivational card, count-badge data,
per-screen rail contents (each screen adds its own later), removing `AuthLayout.tsx`
(keep until every screen is migrated — but the shell replaces its *use*).

---

## PLAN — how to build it

### Files

```
web/src/theme.ts                     — get/set/apply theme, init()
web/src/shell/
  AppShell.tsx                        — the layout route element (replaces AuthLayout use)
  Sidebar.tsx                         — brand + nav + footer; expanded / collapsed / drawer modes
  SidebarNavItem.tsx                  — NavLink + icon + label + optional Badge + Tooltip (collapsed)
  UserMenu.tsx                        — Avatar + name → Menu (Account / Export / Log out)
  ThemeToggle.tsx                     — 3-way light/dark/system
  MobileTopBar.tsx                    — hamburger + brand (< tablet)
  ScreenLayout.tsx                    — <main> content + optional rail; the wrapper screens use
  Placeholder.tsx                     — temporary "screen coming in Phase N" panel
  navItems.ts                         — nav config (label, route, icon key)
  useShellState.ts                    — drawer open/close; derived collapsed/drawer mode from useMediaQuery
  shell.css                           — .app-shell / .sidebar / .rail / .drawer styles (ui- prefix family)
  index.ts
  *.test.tsx
web/src/components/ui/icons.tsx       — add: nav icons + hamburger + sun/moon/monitor
web/src/styles/tokens.css            — add shell tokens
web/src/styles/index.css             — @import "../shell/shell.css"
web/src/App.tsx                       — new <Routes>: AppShell layout + route table + placeholders
```
`AuthLayout.tsx` — left in place (dead once `App.tsx` stops importing it; delete at §13/cleanup).

### Order

1. `theme.ts` + `ThemeToggle` + tokens → verify light/dark/system switch on the current app.
2. `icons.tsx` additions.
3. `shell.css` + `Sidebar` + `SidebarNavItem` + `navItems` (static, expanded mode only).
4. `AppShell` + `ScreenLayout` + `Placeholder`; wire `App.tsx` routes; `/` redirect.
5. `UserMenu` (Account / Export / Log out — reuse existing `api.logout()` flow).
6. Responsive: `useShellState`, collapsed mode, `MobileTopBar`, drawer (native `<dialog>` slide-over, like `Dialog`).
7. Tests. Browser verify at 4 widths + dark. Screenshots. QA. Acceptance.

### Tests (`pnpm test`)

- `navItems` → `Sidebar` renders every destination as a link to the right route.
- Active route → `aria-current="page"` + active class on exactly one item.
- `ThemeToggle`: selecting Dark sets `data-theme="dark"` + persists; System clears it.
- `UserMenu`: opens on click, `Esc` closes, "Log out" calls the logout handler.
- Drawer: hamburger opens it; `Esc` / backdrop closes; focus trapped; focus returns to hamburger.
- `ScreenLayout`: renders `rail` when passed; main is full-width when not.
- `Placeholder`: renders the screen name + phase note.
- Router: `/` → `/timeline`; unknown → `/`; unauthenticated → `/login` (existing guard intact).

### Playwright verification

- Load `/` (→ `/timeline` placeholder inside the shell); **zero new console errors**.
- Screenshot at ~1366 / ~1120 / ~800 / ~390 px + one dark.
- Confirm: sidebar at wide/laptop; icons-only at tablet; drawer at mobile; rail stacks < wide.
- Nav to each route; active item updates; no sideways scroll at any width.
- Keyboard: Tab to skip-link → nav → footer; open/close drawer with keyboard.
- `/login` still renders with **no shell**.

### Visual acceptance (vs `overall.png` + the sidebar in `dashboard.png` / `timeline.png`)

- Sidebar proportions, brand lockup, nav item rhythm, active pill = `--brand-soft`/`--brand`.
- Warm paper background, one system, tokens only.
- No search bar, no bell, no "Spaces", no motivational card, no "Free Plan" label
  (no plan concept in V1) — name only in the user chip.
- Light + dark parity.

### Acceptance criteria (Phase 1 done when)

- [ ] Every authenticated route renders inside the shell; `/` lands on Timeline.
- [ ] All 8 primary nav items + Account/Export/Log out reachable and correct.
- [ ] Responsive shed order works at all 4 widths; no horizontal page scroll.
- [ ] Theme toggle (light/dark/system) works and persists.
- [ ] a11y checklist (§6 of the plan) passes for the shell; keyboard-only usable.
- [ ] `pnpm typecheck && pnpm test && pnpm build` green.
- [ ] Playwright screenshots captured; visual acceptance signed off.
- [ ] Existing `/login` `/register` unaffected (no shell, still render).
- [ ] `docs/design/design-system.md §3.6` updated with the final shell token values used.

### Dependencies / blockers

- None hard. D3 + D10 approved. `--sidebar-w` etc. stay provisional (T1).
- `useMediaQuery` / `breakpoints.ts` (exist). `Dialog` pattern (exists) — reuse for drawer.

---

## Status — ✅ COMPLETE (2026-09-04)

- [x] SPEC + PLAN approved (product owner)
- [x] Implemented — `web/src/shell/**`, `web/src/theme.ts`, routes in `web/src/App.tsx`
- [x] Tests green — 25 shell tests (59 total), `pnpm test`
- [x] Browser-verified + screenshots — Chromium at 1440/900/430 px + dark; no console errors; no h-scroll
- [x] Visual QA + Responsive QA signed off — sidebar matches `overall.png`; D4 shed order confirmed
- [x] Accepted
- [x] Committed

### QA fix applied

- `Avatar` gained a `decorative` prop — the user-menu trigger had a doubled
  accessible name ("email email"); now the visible name is the only label.

### Deferred (not blockers)

- Token values remain PROVISIONAL (T1). `AuthLayout.tsx` is now dead code — delete
  at a cleanup phase. Legacy pages render un-restyled inside the shell until their
  own phase. CI `pnpm test` wiring is a global follow-up.
