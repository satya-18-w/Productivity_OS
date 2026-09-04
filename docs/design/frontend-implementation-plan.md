# Frontend Implementation Plan & Checklist — Productivity OS

> **Status:** planning complete 2026-09-04. Nothing in this list is implemented except
> the design-system **Foundation** (§1). Execute one stage at a time, top to bottom.
>
> **This document does not restate the design system or the requirements.** It links
> them. Authoritative sources:
>
> | For | Read |
> |---|---|
> | What a screen must let the user do | `docs/requirements/v1.md` (§1–§14, N1–N7) |
> | Ratified visual tokens & component contracts | `docs/design/design-system.md` |
> | Visual judgement calls | `docs/design/visual-principles.md` (VP1–VP10) |
> | Per-screen visual spec | `docs/design/screens/<screen>.md` (where one exists) |
> | Frontend stack / styling / test conventions | `docs/architecture/conventions.md` → "Frontend" |
> | System shape, API contract, auth, time model | `docs/architecture/overview.md`, ADR-0002/0004/0005/0006 |
> | Reference images (visual language only) | `docs/design/references/*.png` |

---

## 0. Ground rules

- **0.1 V1 scope is authoritative.** Build only what `docs/requirements/v1.md` grants.
  The reference-only exclusion list in `design-system.md §6.4` is a hard "do not build":
  Dashboard, Notes, Calendar/events, Timeline Week/Month, analytics beyond the five §13
  reports, Focus/Pomodoro, recurring tasks, task priorities/tags/assignees, categories
  on tasks/habits/goals, "Spaces", goal milestones/linked-tasks/%-progress, habit
  longest-streak/consistency%, social/collab, AI, calendar sync, notifications, global
  search, gamification/scoring.
- **0.2 Reference images = visual language only.** Never lift a feature from a PNG that
  `v1.md` does not grant. For an excluded feature shown in a mock, take spacing / colour /
  type only.
- **0.3 One visual system.** Consume tokens from `web/src/styles/tokens.css`; reuse the
  primitives in `web/src/components/**`. No new token / class / colour without approval
  (`CLAUDE.md` → "Design System Changes").
- **0.4 Per-screen workflow (every screen, in order):**
  `SPEC → PLAN → IMPLEMENT → TEST → BROWSER VERIFY → SCREENSHOT → VISUAL QA → RESPONSIVE QA → ACCEPTANCE → COMMIT`
  - **SPEC** — a feature spec under `docs/specs/v1/<milestone>/` (Draft→Approved), written
    against `v1.md §N`, informed by the design `screens/*.md`. No code before it is
    **Approved**.
  - **PLAN** — implementation plan appended to / beside the SPEC; approved before code.
  - **COMMIT** — only after ACCEPTANCE passes; branch off `main`, never commit to `main`
    directly; message trailers per session attribution rules.
- **0.5 Provisional tokens.** All hex in `tokens.css` and the breakpoint px are
  `PROVISIONAL` pending **T1**. Screens may be built against them (names/structure are
  stable); final visual acceptance (§13) waits for T1.

---

## 1. Foundation checklist  —  ✅ DONE (design-system foundation stage)

- [x] Token layer — `web/src/styles/tokens.css` (colour, type, spacing, radius,
      elevation, layout; light + provisional dark)
- [x] Base / reset — `web/src/styles/base.css` (reset, element type, focus-visible,
      reduced-motion, `overflow-x: clip`)
- [x] Primitive styles — `web/src/styles/primitives.css` (`ui-*`)
- [x] Breakpoint infra — `web/src/styles/breakpoints.ts` + `useMediaQuery`
- [x] Import orchestrator — `web/src/styles/index.css`; `main.tsx` updated
- [x] Legacy `styles.css` trimmed to feature classes; inherits new palette via `--accent*`
- [x] UI / layout / productivity primitives + barrels — `web/src/components/**`
- [x] Test runner — Vitest + @testing-library (devDeps), `pnpm test`, 36 tests green
- [x] `conventions.md` → "Frontend" section (styling mechanism, structure, testing)
- [ ] **Follow-up:** add `pnpm test` to the CI workflow (ADR-0007) — do at/after §13
- [ ] **Follow-up (T1):** replace provisional token values; re-run visual QA on all screens

---

## 2. App shell checklist  —  STAGE 1  ·  ✅ COMPLETE (2026-09-04)

Full three-region shell (sidebar + main + per-screen right rail). Build spec + status:
`docs/design/screens/app-shell.md`.

- [x] SPEC + PLAN approved
- [x] `AppShell` — CSS-Grid shell; `ScreenLayout` gives each screen main + optional rail
- [x] `Sidebar` — brand lockup; nav list (`SidebarNavItem`); user chip (`UserMenu`); no "plan" label (no plan concept in V1)
- [x] Nav item states — default / hover / **active** (`--brand-soft` pill, `--brand`)
- [x] Right rail — supplied per-screen via `<ScreenLayout rail>`; stacks below main `< wide` (D4)
- [x] `ThemeToggle` (light/dark/system) + user avatar; **no search, no bell**
- [x] No "Spaces" list (C1)
- [x] Nav = 8 V1 destinations (Timeline/Tasks/Board/Habits/Goals/Categories/Reports/Reviews); no dashboard
- [x] `App.tsx` uses `<AppShell>`; `AuthLayout.tsx` no longer referenced (dead — delete at cleanup)
- [x] Tokens: `--sidebar-w`, `--sidebar-w-collapsed`, `--rail-w`, `--topbar-h`, z-index (all provisional/T1)
- [x] a11y — skip-link, `<nav aria-label="Primary">`, `<main id="main">`, `aria-current="page"`, focus-trapped drawer, labelled controls
- [x] Responsive (D4) — rail stacks `< wide` → labels collapse `< laptop` → drawer `< tablet`; no h-scroll
- [x] Tests — 25 shell tests (`Sidebar`, `ThemeToggle`, `UserMenu`, `AppShell`, `App` routing)
- [x] Playwright — verified at 1440 / 900 / 430 px + dark; screenshots captured
- [x] Acceptance — matches `overall.png` sidebar; a11y + responsive pass
- **Status:** ✅ COMPLETE

---

## 3. Routing checklist  —  STAGE 2  ·  ✅ COMPLETE (2026-09-04, shipped with Phase 1)

Merged into Phase 1 — a shell with dead nav links isn't testable.

- [x] Routes in `web/src/App.tsx`: `/`→`/timeline` · `/timeline` · `/tasks` · `/board` · `/habits` · `/goals` · `/categories` · `/reports` · `/reviews/daily` · `/reviews/weekly` · `/account` · `/export`; `/login` `/register` (no shell)
- [x] Un-built routes render `<Placeholder>`; existing pages wrapped in `<ScreenLayout>`
- [x] Auth guard intact: unauthenticated → `/login`; authed on auth route → `/`
- [x] Unknown route → `/`
- [x] Nav items → ratified routes; `NavLink` active + `aria-current` wiring
- [x] Tests: `App.test.tsx` — `/`→Timeline, placeholders, unknown redirect, guard
- [x] Playwright: every route loads inside the shell
- Deferred: route-level code splitting (not needed at V1 size); Timeline `?view=` param (Phase 2/3)
- **Status:** ✅ COMPLETE

---

## 4. Shared component checklist

Foundation primitives exist (`web/src/components/**`). Extend **only** as a screen
genuinely needs it, and add to `design-system.md §4` + `primitives.css` first.

- [x] Present: Button, IconButton, Card, Badge, Chip, Avatar, Checkbox, Switch,
      ToggleCircle, Field, Input, Textarea, Select, SegmentedControl, Tabs, ProgressBar,
      Divider, Dialog, Tooltip, Stack, Inline, Container, Section, PageHeader,
      CategoryIndicator, StatusBadge, EmptyState, LoadingState, ErrorState, StatCard,
      ListRow + ListGroupHeader
- [ ] **SplitButton** (primary + menu caret) — needed by Timeline "Add ▾" (§4.5)
- [ ] **Menu / Dropdown** (kebab actions, "Add" menu) — accessible menu pattern
- [ ] **DatePicker / DateStepper** — `‹ date ›` + "Today" + native date input (Timeline, Reports range)
- [ ] **DateRangePicker** — Reports (the one required control, §13)
- [ ] **Mini month calendar** (right-rail, Monday-first D8) — Timeline rail
- [ ] **Donut / RingChart**, **BarChart**, **BarList** — Reports only (choice pending **R1**)
- [ ] **DataTable / totals table** — reuse legacy `table.totals`; wrap as a primitive for Reports
- [ ] **Toast / inline status** — save/delete feedback (decide pattern; keep minimal)
- [ ] Each new primitive: style in `primitives.css`, tokens only, a11y, `*.test.tsx`, barrel export
- **Status:** foundation done; the above are added per-screen on demonstrated need

---

## 5. Design-token checklist

Reference `design-system.md §3`. Do not add values ad hoc.

- [x] Colour — brand (D1), warm neutrals (D5), semantic, category palette (D2, presentation-only), text
- [x] Typography — Inter stack (D9); semantic steps (`--fs-display/heading/title/body/small/caption`); line-heights; weights
- [x] Spacing — 4px scale `--sp-1..7` (unchanged)
- [x] Radius — `--radius-xs..full` (unchanged) + `--radius-tile`
- [x] Elevation — `--shadow-sm/md/lg` (unchanged)
- [x] Layout — `--content-max`, `--gutter`, `--sidebar-w`/`--rail-w` (provisional, D3)
- [x] Breakpoints — `breakpoints.ts` (`tablet/laptop/wide`, provisional px)
- [ ] **T1:** ratify exact hex (light + dark) + final breakpoint px; update `tokens.css`;
      remove `PROVISIONAL` markers; re-run §8 on every built screen
- [ ] Per screen: audit for any hardcoded value → replace with a token or propose one
- **Status:** structure done; values provisional pending T1

---

## 6. Accessibility checklist  (apply to EVERY screen — VP8)

- [ ] Semantic HTML: one `<h1>` per screen (the `PageHeader` title); logical heading order; landmarks (`<nav> <main>` + labelled regions)
- [ ] Keyboard: every interactive element reachable & operable; logical tab order; no keyboard trap (except intended Dialog); `SegmentedControl`/`Tabs` arrow-key nav; kebab `Menu` arrow/Esc
- [ ] Focus: visible `:focus-visible` ring on all interactives; focus moved into Dialog on open and restored on close; drawer traps focus
- [ ] Names: icon-only controls have `aria-label`/`title`; form controls wired via `<Field>` (`label` + `aria-describedby` for hint/error); `aria-invalid` on error
- [ ] State not by colour alone: category = dot **+ name**; goal status = dot **+ text**; overdue = colour **+ label/icon**; completed = strikethrough **+** check
- [ ] Live regions: async errors as `role="alert"`; loading as `role="status" aria-live="polite"`
- [ ] Contrast: text & meaningful UI ≥ WCAG AA in light **and** dark (re-check at T1)
- [ ] Hit targets: ≥ 40px primary; ≥ 32px in dense grids (habit toggles, week strips)
- [ ] `prefers-reduced-motion` respected (handled globally in `base.css`; don't override)
- [ ] Dark-mode parity: nothing readable only in one theme

---

## 7. Responsive checklist  (apply to EVERY screen — D4 / VP9)

- [ ] Layout uses Grid / Flexbox / normal flow — no absolute positioning for structure
      (timeline blocks against the hour axis are the only sanctioned exception, in a grid cell)
- [ ] Shed order: right-rail content → sidebar labels → sidebar drawer
- [ ] Main content + the one primary action survive to `< tablet`
- [ ] Wide content scrolls inside its **own** `overflow-x:auto` container (habit grid,
      totals table, board columns) — the page never scrolls sideways
- [ ] Segmented controls / filter-chip rows scroll horizontally on narrow widths
- [ ] KPI/stat rows wrap 4 → 2 → 1
- [ ] Verified at ~1366, ~1024, ~768, ~390 CSS px (indicative; final px at T1)

---

## 8. Browser / Playwright visual-QA checklist  (apply to EVERY screen)

Tooling: `playwright-cli` (Chromium available). Vite dev server; backend running for
data screens (or routed fixtures). No committed E2E suite (ADR-0007) — this is per-screen
verification.

- [ ] App starts; screen route loads; **zero console errors** (favicon 404 / expected 401s noted, not new errors)
- [ ] Screenshot at desktop width → save to scratchpad, attach to the screen's PR/notes
- [ ] Screenshot at mobile width (`--mobile` or `resize`)
- [ ] Dark-mode screenshot (`data-theme="dark"`)
- [ ] **Visual comparison** against the reference PNG (where the feature is in scope):
      layout structure, spacing rhythm, type hierarchy, colour roles, component variants
      — treat the PNG as a spec, note every deliberate deviation and why (v1.md conflicts)
- [ ] Computed-style spot checks: brand colour, surface, border, radius, spacing resolve to tokens
- [ ] No horizontal page scroll at any width
- [ ] Keyboard walk-through of the primary flow
- [ ] Interaction smoke: primary action, one edit, one delete, empty/loading/error states

---

## 9. V1 screen implementation order

| # | Stage | Req / spec | Ref image | Blocking deps (beyond §14) |
|---|---|---|---|---|
| 1 | **App Shell** | architecture (§2) | `overall.png` | **D3**; T1 for final dims |
| 2 | **Routing** | architecture (§3) | — | **D10**; depends on §9 list |
| 3 | **Timeline — Day** | `v1.md §3,§4,§5,§6` · `screens/timeline.md` | `references/timeline.png` | Shell, Routing; **G1** (block geometry) in the SPEC; timeline API |
| 4 | **Timeline — Agenda** | `v1.md §5,§6` · `screens/timeline-agenda.md` | `references/timeline-agenda.png` | Stage 3 (shares Timeline shell/toolbar) |
| 5 | **Tasks — List** | `v1.md §7` · `screens/tasks.md` | `references/tasks.png` | Shell, Routing; tasks API |
| 6 | **Board / Kanban** | `v1.md §8` · *(no design spec, no ref image)* | — | Stage 5 (same task model); tasks API; drag-and-drop approach decision |
| 7 | **Habits** | `v1.md §9` (+Q9,Q11) · `screens/habits.md` | `references/habits.png` | Shell, Routing; habits API |
| 8 | **Goals** | `v1.md §10` · `screens/goals.md` | `references/goals.png` | Shell, Routing; goals API |
| 9 | **Categories** | `v1.md §2` · `screens/categories.md` | `references/categories.png` | Shell, Routing; categories API; **C1** bounds scope (build name+archive only) |
| 10 | **Reports** | `v1.md §13` (+Q7,Q8,Q10), N4 · `screens/analytics.md` | `references/analytics.png` *(visual language only)* | Shell, Routing; **R1** (viz choices) in the SPEC; **reports API** |
| 11 | **Daily Review** | `v1.md §11` (+Q1) · *(no design spec, no ref)* | — | Shell, Routing; reviews API; daily reference-totals API (§6, habits) |
| 12 | **Weekly Review** | `v1.md §12` (+Q2) · *(no design spec, no ref)* | — | Stage 11; weekly reference-totals (actual/category, habit counts, tasks→DONE) |
| 13 | **Account** | `v1.md §1` · *(no design spec, no ref)* | — | Shell, Routing; account API (exists: email, password change, timezone) |
| 14 | **Auth (Login/Register)** | `v1.md §1` (+Q4,Q6) · *(no spec; `overall.png` panel 12)* | `overall.png` | Exists in code — **restyle to primitives**; no shell (uses `.center` layout) |
| 15 | **Data Export** | `v1.md §14` (+**Q3**) · *(no spec, no ref)* | — | Shell, Routing; **export format (Q3)** + export endpoint |

> Auth (14) can be done any time — it has no shell/routing dependency and largely exists.
> It is last per the requested priority order; pull it forward if convenient.

---

## 10 + 11. Per-screen checklists

Each block = the **implementation checklist** (§10) followed by the **visual acceptance
checklist** (§11). All routes are **PENDING D10**. Detail lives in the linked
`screens/*.md`; this is the execution list. Every screen also runs §6, §7, §8, §12.

### Shared visual-acceptance template (referenced by each screen)

- [ ] Shell chrome identical to other screens (§2); only the main column + rail differ
- [ ] `PageHeader`: eyebrow (if any) + H1 + factual subtitle; **no motivational copy** (VP3/D6)
- [ ] One `--brand`-filled primary action; everything else ghost/secondary/icon/link (VP1)
- [ ] Category colour only as identification, always with a text label (VP2/VP8)
- [ ] Spacing rhythm, radii, shadows, type steps = tokens; matches the reference PNG's proportions
- [ ] Empty / loading / error states present and styled
- [ ] Light + dark parity; no h-scroll; responsive per §7
- [ ] Deviations from the PNG documented with the `v1.md` clause that requires them

---

### Stage 3 — Timeline (Day)

- **Req/spec:** `v1.md §3,§4,§5,§6` · `docs/design/screens/timeline.md` (owns the shared Timeline shell)
- **Reference:** `references/timeline.png`
- **Route:** `/timeline?view=day&date=<today>` *(D10 ratified)*
- **Layout:** shell → header → **Timeline toolbar** (view switcher `Day / Agenda` only · date stepper `‹ date › Today` · `SplitButton` "Add ▾") → timeline body (hour axis + **planned lane / actual lane**, blocks positioned by time) → right rail (mini month calendar, "Today's tasks" read-only, optional)
- **Components:** `SegmentedControl`, DateStepper, `SplitButton`, `Menu`, `CategoryIndicator`, mini calendar, Dialog (create/edit block: **start / end / category only**), `EmptyState`
- **States:** loading; empty (no blocks); block hover; edit dialog open; save error; midnight-spanning block; block outside visible hours (define: clip / scroll — resolve in SPEC / G1)
- **Interactions:** prev/next/today/pick date (URL `?date`); add planned block; add actual block; click block → edit; delete; **no** per-block checkbox / tags / assignees (excluded, 0.1)
- **Responsive:** rail drops first; lanes stack or scroll horizontally (`.tl-scroll`); switcher scrolls
- **A11y:** blocks are buttons with an accessible name (title + time + plan/actual + category); axis labels; planned vs actual distinguished by text/shape not colour alone
- **Tests:** toolbar renders; date param changes; add-block dialog validates end>start; midnight block placement; planned/actual rendered distinctly
- **Playwright:** load with fixture blocks; screenshot desktop/mobile/dark; verify axis + two lanes; compare to `timeline.png` (note: ref shows one merged list — we keep §5's planned/actual distinction; document)
- **Visual acceptance:** shared template + hour axis legible; blocks time-proportional-or-ordered per SPEC; planned vs actual obvious; "now" indicator
- **Acceptance criteria:** user can view a chosen date's planned+actual blocks positioned against hours, visually distinguishable, midnight-spanning correct (§5); per-category planned/actual totals reachable (§6); create/edit/delete blocks (§3,§4)
- **Status:** ✅ **COMPLETE (2026-09-04)** — G1 resolved (time-proportional, two lanes, category colour + dashed/solid). `web/src/features/timeline/**` + `web/src/components/date/MiniCalendar`. 23 tests. Deferred: `SplitButton`, Agenda view (Stage 4). Details + status: `screens/timeline.md` → "Phase 2".

### Stage 4 — Timeline (Agenda)

- **Req/spec:** `v1.md §5,§6` · `screens/timeline-agenda.md`
- **Reference:** `references/timeline-agenda.png`
- **Route:** `/timeline?view=agenda&date=<today>` *(D10 ratified)*
- **Layout:** shares Timeline shell/toolbar (§3); body = **filter-chip row** (All + per-category counts) → chronological agenda list (time-range rail + block rows) → foot "add" affordance; rail = "Time allocation" (§6 planned-vs-actual for the date), optional donut
- **Components:** `Chip` (filter, toggle), `ListRow` variant, `CategoryIndicator`, rail totals table / bar, `Dialog` (same block form as §3)
- **States:** loading; empty; filtered-empty; single-day list
- **Interactions:** filter by category; sort (Time only — meaningful V1 sort); row → edit; add block; date stepper
- **Responsive:** single column already; rail drops first; chips scroll — **this is the mobile fallback for Day**
- **A11y:** filter chips = toggle buttons with `aria-pressed`; list is an ordered list; time column associated with each row
- **Tests:** filter narrows list; sort stable; add affordance opens the block form; totals reflect fixture data
- **Playwright:** screenshot desktop/mobile/dark; compare to `timeline-agenda.png` (drop checkboxes/tags/avatars/"priorities" — excluded)
- **Visual acceptance:** shared template + rail connector + node dots + category-tinted rows
- **Acceptance criteria:** same single date as Day, list form; §6 per-category totals visible; create/edit/delete via the block form
- **Status:** ✅ **COMPLETE (2026-09-04)** — `AgendaList` + view switcher on the shared `TimelineScreen` (`TimelineDay` refactored → `TimelineScreen`). Merged time-ordered list, category filter chips, Planned/Actual badges, `?view=agenda` param. 88 tests. Details: `screens/timeline-agenda.md` → "Phase 3".

### Stage 5 — Tasks (List)

- **Req/spec:** `v1.md §7` (+ §8 shares the model) · `screens/tasks.md`
- **Reference:** `references/tasks.png`
- **Route:** `/tasks` *(D10 ratified)*
- **Layout:** shell → header (eyebrow "TASKS" + H1 + subtitle) → view switcher `All / Today / Upcoming / Overdue / Completed` (**no "Starred"**) → `SplitButton` "Add Task" → toolbar (sort: Due date) → **grouped list** (`ListGroupHeader` with accent bar + count; `ListRow` per task) → rail: "Task stats" donut, optional
- **Components:** `SegmentedControl`, `SplitButton`, `ListGroupHeader`, `ListRow`, `Checkbox`, `Menu` (state change / edit / delete), `Dialog` (create/edit: **title / description / due date only**), donut, `EmptyState`
- **States:** loading; empty (per tab); completed (strikethrough + muted); overdue (date in `--danger` + label); edit dialog; save error
- **Interactions:** checkbox → `DONE` ⇄ prior state; menu → set any of `BACKLOG/TODO/IN_PROGRESS/DONE`; edit; delete; tab/sort via URL params. **No** priority / star / assignee / category (excluded, 0.1)
- **Responsive:** rail drops; KPI wraps; row meta wraps under title; switcher scrolls
- **A11y:** checkbox labelled by task title; group headings are real headings or labelled; overdue conveyed by text not just red
- **Tests:** checkbox toggles state; group assignment by due date + state; tab filter; create dialog rejects empty title; completed styling
- **Playwright:** fixture tasks across states; screenshot desktop/mobile/dark; compare to `tasks.png` **minus** priority/star/assignee/category columns and the "Categories"/"Priority Breakdown" rail widgets
- **Visual acceptance:** shared template + grouped sections with coloured accent bars + donut by state
- **Acceptance criteria:** create (title + optional description + due date); edit those; change state any→any; delete; see all tasks (§7)
- **Status:** ☐ NOT STARTED

### Stage 6 — Board / Kanban

- **Req/spec:** `v1.md §8` · **no design spec, no reference image** — visual language from `tasks.png` + existing `.board-*` CSS
- **Route:** `/board` (or `/tasks?view=board`) *(PENDING D10; decide toggle-vs-route)*
- **Layout:** shell → header → four fixed columns `BACKLOG → TODO → IN_PROGRESS → DONE` (that order, not configurable) → column head (label + count `Badge`) → task cards
- **Components:** reuse/migrate `.board-*` → `ui-` board column + task card; `Menu`; `Dialog` (same task form as §5); `EmptyState` per column
- **States:** loading; empty column; drag-over column; card dragging; save error on move
- **Interactions:** move a card between columns → sets state (§8); reflects create/edit/delete from §5. **No** column CRUD, WIP limits, swimlanes, filters, saved views, manual reorder within a column (excluded). Ordering within a column: newest-first (Q5)
- **Decision needed:** drag-and-drop approach — native HTML DnD (no dep) vs a11y-friendly keyboard-move fallback. **Prefer native DnD + an explicit "move to column" control** for keyboard/touch (also satisfies a11y).
- **Responsive:** columns horizontal-scroll in their own container below `laptop`; never page h-scroll
- **A11y:** DnD must have a non-drag equivalent (the move control / `Menu`); column regions labelled; card has accessible name + current column
- **Tests:** four columns in fixed order; count per column; move control changes state; new/edited/deleted task reflected
- **Playwright:** fixture tasks; screenshot desktop/mobile/dark; verify column order + counts; drag one card; keyboard-move one card
- **Visual acceptance:** shared template + columns match `tasks.png` card styling; counts; drag affordance
- **Acceptance criteria:** view all tasks in four fixed columns; move any→any sets state; board reflects §7 changes (§8)
- **Status:** ☐ NOT STARTED

### Stage 7 — Habits

- **Req/spec:** `v1.md §9` (+ Q9 future dates, Q11 archive keeps history) · `screens/habits.md`
- **Reference:** `references/habits.png`
- **Route:** `/habits` *(D10 ratified)*
- **Layout:** shell → header → view switcher `Today / This Week / This Month / All Habits` → `SplitButton` "Add Habit" → **habit grid** (rows = active habits: name + 7 dated weekday `ToggleCircle`s + current-streak number + `Menu`); rail: current streak, this-week dots, optional
- **Components:** `SegmentedControl`, `SplitButton`, `ToggleCircle`, `Menu` (archive / unarchive / edit name / delete), `Dialog` (create: **name only**), grid table, `EmptyState`; separate Archived list/tab
- **States:** loading; empty; today-only checklist (mobile); archived list; toggle save error; future-date toggle allowed (Q9)
- **Interactions:** click weekday circle → mark/unmark that date; streak = consecutive completed dates ending today/yesterday (display only; server-computed); archive/unarchive (history preserved, Q11); edit name; delete. **No** longest streak, consistency %, habit categories, sub-labels (excluded, 0.1)
- **Responsive:** grid scrolls horizontally below `tablet`, first column sticky; mobile prefers "Today" checklist
- **A11y:** each `ToggleCircle` has a full accessible name ("<habit> — <weekday> <date>, completed/not"); grid is a table with header cells; today column marked
- **Tests:** toggle marks a date; streak text renders from fixture; archive moves row to archived; unarchive restores; week is ISO/Monday (D8)
- **Playwright:** fixture habits + completions; screenshot desktop/mobile/dark; compare to `habits.png` **minus** longest-streak/consistency/"Habit Categories"/motivational banner
- **Visual acceptance:** shared template + grid alignment + filled-green vs hollow circles + today emphasis + plain streak number (no flame animation, VP3)
- **Acceptance criteria:** create habit (name); mark/unmark a date; archive/unarchive; see current streak per active habit; see a chosen date's completions (§9)
- **Status:** ☐ NOT STARTED

### Stage 8 — Goals

- **Req/spec:** `v1.md §10` · `screens/goals.md`
- **Reference:** `references/goals.png`
- **Route:** `/goals` *(D10 ratified)*
- **Layout:** shell → header → filter by **state** (`All / Not started / In progress / Achieved / Abandoned`) — **not by category** → primary "New Goal" → goal list (rows: title + description + target-date chip + `StatusBadge` + `Menu`); rail: donut by state, optional
- **Components:** `SegmentedControl` (state filter), `Button`, `ListRow`/goal row (reuse `.goal-*`), `StatusBadge` (the four V1 labels verbatim), `Select`/`Menu` for state, `Dialog` (create/edit: **title / description / target date only**), donut, `EmptyState`
- **States:** loading; empty (per filter); edit dialog; save error
- **Interactions:** set progress state any→any (Not started / In progress / Achieved / Abandoned); edit title/description/target date; delete; filter by state. **No** %, progress bars, "X/Y tasks", categories, milestones, "On Track/At Risk" (excluded, 0.1)
- **Responsive:** rows reflow (meta wraps); rail drops; KPI wraps
- **A11y:** `StatusBadge` conveys state by dot **+ text**; state control is a labelled select/menu
- **Tests:** four states render with correct labels; state change persists via fixture handler; filter narrows; create rejects empty title; **no `%` or task-count text anywhere**
- **Playwright:** fixture goals in all four states; screenshot desktop/mobile/dark; compare to `goals.png` **minus** progress bars / % / task counts / category chips / milestones rail
- **Visual acceptance:** shared template + state chip colours + target-date chip; list rhythm
- **Acceptance criteria:** create (title + optional description + target date); edit those; set one of four states manually; list with state; delete (§10)
- **Status:** ☐ NOT STARTED

### Stage 9 — Categories

- **Req/spec:** `v1.md §2` · `screens/categories.md`
- **Reference:** `references/categories.png`
- **Route:** `/categories` *(D10 ratified)*
- **Layout:** shell → header → optional `Active / Archived` switch → primary "New Category" → **simple list** (row: name + `Menu` [Rename / Archive]); minimal rail
- **Components:** `Button`, `ListRow`, `Menu`, `Dialog`/inline form (create/rename: **name only**), `EmptyState`
- **States:** loading; empty; rename inline/dialog; archived list; name-conflict error (409); validation error
- **Interactions:** create; rename; archive (removes from assignment pickers, keeps existing block→category links). **No** item counts / cross-entity breakdowns / icons / donut / import-export / recently-used (excluded, 0.1)
- **C1 (unresolved) bounds scope:** build name + archive. **Unarchive** only if C1 confirms it exists. **Stored per-category colour** only if C1 confirms; until then any colour is client-side presentation (D2), not persisted.
- **Responsive:** single-column list throughout; rail drops
- **A11y:** rename control labelled; archive confirmation accessible; list is a real list
- **Tests:** create; rename; duplicate-name error; archive removes from active list; (unarchive iff C1)
- **Playwright:** fixture categories; screenshot desktop/mobile/dark; take visual language (card/tile styling) from `categories.png` but **not** the counts/breakdown/donut
- **Visual acceptance:** shared template; restrained — closer to `v1.md`'s "flat list" than the mock's rich cards
- **Acceptance criteria:** create (name); rename; archive; see active list (§2)
- **Status:** ☐ NOT STARTED · scope bounded by C1

### Stage 10 — Reports

- **Req/spec:** `v1.md §13` (exhaustive 5-report set) + Q7 (sum of durations), Q8 (explicit Uncategorized), Q10 (DONE transitions), N4 (DST-correct) · `screens/analytics.md`
- **Reference:** `references/analytics.png` — **visual language only**; the screen is a *Reports* page, not an analytics dashboard
- **Route:** `/reports` *(D10 ratified)*
- **Layout:** shell → header ("Reports") → **DateRangePicker** (the one required control) → card grid, one card per report → minimal rail
- **The five reports (nothing else):** 1 Time by category · 2 Planned vs actual by category (planned/actual/diff) · 3 Habit completion (days + rate) · 4 Task throughput (count → DONE in range) · 5 Daily actual totals (per day)
- **Components:** `DateRangePicker`, `Card`, and per **R1**: Donut/BarList (r1), totals `DataTable` / grouped bar (r2, reuse `table.totals` w/ `.pos`/`.neg`), table (r3), `StatCard` (r4), `BarChart` (r5). Each card shows the figures as a caption (P3 — numbers are the point). `dataviz` skill before any chart.
- **States:** loading; empty range (no data); range crossing DST still totals correctly
- **Interactions:** change date range → all reports recompute deterministically. **No** drill-down, saved views, range-vs-range, trend lines, deltas, insights, streak, per-report export, focus-time, "top categories by tasks completed" (excluded, 0.1)
- **Responsive:** cards 2-up → 1-up; charts min-height; wide tables scroll in-container
- **A11y:** each chart has a text alternative / adjacent figures table; range picker fully keyboard-operable; colour-blind-safe series (category palette + labels)
- **Tests:** the five reports render for a fixture range; Uncategorized bucket present (Q8); diff sign classes on r2; r4 is a single count; range change refetches
- **Playwright:** fixture range data; screenshot desktop/mobile/dark; compare **only** visual language to `analytics.png` (KPI/card styling) — not its trend/delta/insight/heatmap widgets
- **Visual acceptance:** shared template + one card per report + honest figure captions + no motivational/insight copy
- **Acceptance criteria:** user views each of the five §13 reports over a chosen range, deterministic, DST-correct (§13, N4)
- **Blockers:** **R1** (viz choice) in the SPEC; **reports backend API** must exist
- **Status:** ☐ NOT STARTED · blocked on R1 + reports API

### Stage 11 — Daily Review

- **Req/spec:** `v1.md §11` + Q1 (four fixed prompts) · **no design spec, no reference** — use shared primitives + form patterns
- **Route:** `/reviews/daily?date=<date>` *(D10 ratified)*
- **Layout:** shell → header → date stepper → **reference panel** (that date's actual time per category + which habits completed — display only, §6/§9 data) → **four fixed free-text prompts** (`Field` + `Textarea`) → Save
- **Components:** DateStepper, `Card`, `Field`, `Textarea`, `Button`, read-only totals list, `EmptyState` (no review yet)
- **States:** new (blank); previously completed (editable); viewing a past review; save success/error; loading reference data
- **Interactions:** answer/edit prompts; save; navigate dates; edit an existing review. Prompts are **fixed, not user-editable** (§11). Reference data is **display-only**. May record for future dates (Q9)
- **Responsive:** single column; reference panel stacks above prompts on mobile
- **A11y:** each prompt is a labelled textarea; reference panel is a labelled region; save state announced
- **Tests:** exactly four prompts, correct wording (Q1); save persists; reload shows saved answers; reference totals render from fixture
- **Playwright:** screenshot desktop/mobile/dark; verify four prompts + reference panel
- **Visual acceptance:** shared template; calm form; reference data visually secondary
- **Acceptance criteria:** complete a daily review (four free-text answers) for a date; edit it; view a past one; see that date's actual-per-category + habit completions while doing it (§11)
- **Blockers:** reviews backend API; the §6 per-date totals + habit-completion reads
- **Status:** ☐ NOT STARTED

### Stage 12 — Weekly Review

- **Req/spec:** `v1.md §12` + Q2 (four fixed prompts) · **no design spec, no reference**
- **Route:** `/reviews/weekly?week=<ISO-week>` *(D10 ratified)*
- **Layout:** as §11 but week-scoped; reference panel = week's actual time per category + habit completion **counts** + number of tasks that entered `DONE` that week
- **Components:** week stepper (ISO week, Monday-first, D8), otherwise as §11
- **States:** as §11
- **Interactions:** as §11; week boundaries = ISO week in the account timezone (§12, N4)
- **Responsive / A11y / Tests / Playwright / Visual acceptance:** as §11, week variant; test week resolution is ISO/Monday and DST-safe
- **Acceptance criteria:** complete/edit/view a weekly review for an ISO week; see that week's totals (actual/category, habit counts, tasks→DONE) while doing it (§12)
- **Blockers:** reviews backend API; weekly reference reads (incl. tasks→DONE-in-range, Q10)
- **Status:** ☐ NOT STARTED

### Stage 13 — Account

- **Req/spec:** `v1.md §1` + Q4 (browser-detected IANA tz, fallback UTC), Q6 (password policy) · **no design spec, no reference**. `web/src/pages/Account.tsx` exists — audit & restyle
- **Route:** `/account` *(D10 ratified)*
- **Layout:** shell → header → sections: email (display), change password (`Field`s: current / new / confirm), timezone (`Select` of IANA names), log out
- **Components:** `Card`, `Field`, `Input`, `Select`, `Button`, inline success/error
- **States:** loading; password change success/error (policy violation → field errors, Q6); timezone save; logout
- **Interactions:** change password (enforces Q6 rules client-side + server); set timezone; log out. **No** profile fields beyond email/password/timezone; **no** self-service delete, email change, MFA (excluded)
- **Responsive:** single column; sections stack
- **A11y:** password fields have visible requirements text via `aria-describedby`; errors as `role="alert"`; current/new/confirm properly labelled + `autocomplete`
- **Tests:** password form validates Q6 rules; timezone select persists; logout clears session + redirects
- **Playwright:** screenshot desktop/mobile/dark; verify all three sections
- **Visual acceptance:** shared template; plain settings form
- **Acceptance criteria:** log in/out; change password while logged in; set/change timezone; see only own data (§1)
- **Status:** ☐ NOT STARTED (migrate existing page)

### Stage 14 — Auth (Login / Register)

- **Req/spec:** `v1.md §1` + Q4, Q6 · no spec; visual language from `overall.png` panel 12. `web/src/pages/Login.tsx` + `Register.tsx` **exist and render** — restyle to primitives, no shell
- **Route:** `/login`, `/register` *(exist; keep)*
- **Layout:** centered `Container --narrow` on the `.center` background (no AppShell) → `Card` → brand → H1 + subtitle → `Field`s → primary `Button` → cross-link
- **Components:** `Container`, `Card`, `Field`, `Input`, `Button`, `Alert`/inline error
- **States:** idle; submitting; auth error (invalid credentials / 429 rate-limit copy); register validation errors (Q6 password, email format); timezone auto-detected + hidden or shown (Q4)
- **Interactions:** login → session cookie → home; register → account (browser IANA tz, fallback UTC) → logged in; link between the two
- **Responsive:** already single-column; card full-width with margin on mobile
- **A11y:** `autocomplete` (`email`, `current-password`, `new-password`); errors `role="alert"`; password requirements described; submit disabled while busy with a busy label
- **Tests:** login submits + redirects (mock api); wrong-password message; register enforces Q6; 429 → wait message
- **Playwright:** screenshot login + register, desktop/mobile/dark (already verified rendering in the foundation stage); confirm primitives, green primary
- **Visual acceptance:** shared template minus shell; matches the calm auth panel in `overall.png`
- **Acceptance criteria:** create account (email + Q6 password + tz); log in; log out (§1)
- **Status:** ☐ NOT STARTED (pages exist; restyle + wire Q4/Q6)

### Stage 15 — Data Export

- **Req/spec:** `v1.md §14` + **Q3 (format — single JSON doc vs CSV archive: OPEN)** · no spec, no reference
- **Route:** `/export` (or a section of `/account`) *(D10 ratified)*
- **Layout:** shell → header → `Card`: what's included (categories, planned/actual blocks, tasks, habits + completions, goals, daily + weekly reviews) → single "Export my data" `Button` → download
- **Components:** `Card`, `Button`, progress/success/error inline
- **States:** idle; generating; download ready; error
- **Interactions:** one click → full snapshot download, no support needed (§14, P5). **No** scheduled / partial / filtered export, no import, no third-party destination (excluded)
- **Responsive:** single column
- **A11y:** button has a clear name; generating state announced; the download is a real link/response, not a JS-only trigger where avoidable
- **Tests:** button triggers the export request; success/error states; (format assertions once Q3 is resolved)
- **Playwright:** screenshot desktop/mobile/dark; trigger export against a fixture endpoint
- **Visual acceptance:** shared template; minimal
- **Acceptance criteria:** user exports **all** their data in one open documented format, self-service (§14)
- **Blockers:** **Q3** (export format) must be resolved; export endpoint must exist
- **Status:** ☐ NOT STARTED · blocked on Q3 + export API

---

## 12. Testing / build checklist  (per screen + global)

Per screen, before ACCEPTANCE:
- [ ] `pnpm typecheck` — clean
- [ ] `pnpm test` — new `*.test.tsx` for the screen's interactive logic + a11y-relevant behaviour; all green
- [ ] `pnpm build` — succeeds
- [ ] Component tests cover: primary action, one edit, one delete, empty/loading/error, keyboard path, any excluded-feature guard (assert the excluded UI is absent)
- [ ] No test needs the real backend — use fixture handlers / props
- [ ] Playwright pass per §8

Global (before §13):
- [ ] Add `pnpm test` to the CI workflow (ADR-0007)
- [ ] Full `pnpm typecheck && pnpm test && pnpm build` green
- [ ] Bundle size sanity (note growth per screen; no surprise deps)

---

## 13. Final V1 frontend acceptance checklist

- [ ] Every screen in §9 (stages 3–15) is `Status: DONE` and committed
- [ ] Every `v1.md` observable capability (§1–§14) is reachable in the UI; cross-check each clause
- [ ] **No excluded feature is present** anywhere (walk `design-system.md §6.4` list; grep the codebase)
- [ ] App shell + routing consistent across all screens (§2, §3)
- [ ] **T1 complete** — provisional token values replaced; every screen re-screenshotted & re-QA'd in light + dark
- [ ] Accessibility (§6) verified on every screen; keyboard-only walkthrough of every primary flow; automated a11y check if available
- [ ] Responsive (§7) verified on every screen at 4 widths; no page ever scrolls sideways
- [ ] Visual QA (§8, §11) signed off against every available reference PNG; deviations logged with `v1.md` justification
- [ ] `pnpm typecheck && pnpm test && pnpm build` green; CI runs all three
- [ ] N5 (responsive), N7 (feels-immediate at expected load) — spot-checked
- [ ] Design docs updated: `conventions.md` reflects final shell/routing; `design-system.md` D3/D10/T1/C1/R1 marked resolved; screen specs' "completion status" updated
- [ ] Legacy `web/src/styles.css` — dead feature classes removed once every screen is migrated to `ui-*`
- [ ] `web/src/AuthLayout.tsx` removed if fully replaced by `AppShell`

---

## 14. Stage dependencies & blockers — what must be resolved before each stage

| Before stage | Must be resolved / in place |
|---|---|
| **1 App Shell** | ✅ **D3 approved** (2026-09-04) & recorded. Nav list = §9. "Spaces" stays out (C1). T1 for final `--sidebar-w`/`--rail-w` (proceeding provisional). Needs `screens/app-shell.md` SPEC+PLAN approved. |
| **2 Routing** | ✅ **D10 approved** (2026-09-04) & recorded — proposed routes, Tasks/Board separate, `/` = Timeline. Needs SPEC+PLAN approved. |
| **3 Timeline Day** | Stages 1–2 done. **Timeline SPEC** approved, resolving **G1** (block geometry, off-window blocks) and how planned/actual render. Timeline API (planned + actual blocks, per-date; §6 totals). Category list API. |
| **4 Timeline Agenda** | Stage 3 done (shared shell). Same APIs. |
| **5 Tasks List** | Stages 1–2. Tasks SPEC. Tasks API (CRUD + state transitions). |
| **6 Board** | Stage 5 done. Drag-and-drop approach decided (native DnD + explicit move control). Same tasks API + transition-time recording (Q10). |
| **7 Habits** | Stages 1–2. Habits SPEC (streak rules, archive/history Q11, future dates Q9). Habits API (habit CRUD, completion toggle per date, streak, archived list). |
| **8 Goals** | Stages 1–2. Goals SPEC. Goals API (exists — M5). |
| **9 Categories** | Stages 1–2. Categories SPEC. Categories API. **C1** decision (unarchive? stored colour?) — or build the minimal name+archive slice and defer the rest. |
| **10 Reports** | Stages 1–2. Reports SPEC resolving **R1** (viz per report) + Q7/Q8/Q10 interpretations. **Reports backend API** (the five aggregations, date-range, DST-safe). `dataviz` skill. |
| **11 Daily Review** | Stages 1–2. Reviews SPEC (prompt set Q1 — resolved; layout). Reviews API. Per-date reference reads (§6 totals, habit completions). |
| **12 Weekly Review** | Stage 11 done. Q2 prompt set (resolved). Weekly reference reads incl. tasks→DONE-in-week (Q10). ISO-week resolution in account tz (N4). |
| **13 Account** | Stages 1–2. Account SPEC. Account API (exists). Q4 (tz default), Q6 (password policy) — resolved; wire them. |
| **14 Auth** | None hard (pages exist). Q4 + Q6 wiring. Can be done anytime. |
| **15 Data Export** | Stages 1–2. **Q3 (export format) — OPEN, must be decided.** Export endpoint must exist. |
| **13 Final acceptance** | Every stage done + **T1 complete** + CI runs tests. |

---

## Report

**Files created:** `docs/design/frontend-implementation-plan.md` (this file). No source
modified. Nothing committed.

**Implementation order:** 1 App Shell → 2 Routing → 3 Timeline Day → 4 Timeline Agenda →
5 Tasks List → 6 Board → 7 Habits → 8 Goals → 9 Categories → 10 Reports → 11 Daily Review
→ 12 Weekly Review → 13 Account → 14 Auth → 15 Data Export. (Auth may be pulled forward —
no shell/routing dependency.)

**Blockers (hard, must clear before the stage runs):**
- Stages 1–2 — ✅ **cleared.** D3 (full 3-region shell) and D10 (routes, Tasks/Board
  separate, `/` = Timeline) approved 2026-09-04. Still need each stage's SPEC+PLAN approved.
- Stage 10 — **R1** (report visualisations) + a **reports backend API**.
- Stages 11–12 — **reviews backend API** + the reference-total reads.
- Stage 15 — **Q3** (export format) open + an **export endpoint**.
- Stage 3 — **G1** (timeline block geometry) to be settled in the Timeline SPEC.
- Stage 9 — **C1** bounds Categories scope (buildable minimally without it).
- Every stage — its **feature SPEC + PLAN approved** first (E4); a **backend API** for its
  data (Timeline/Tasks/Habits/Goals/Categories/Account exist; Reports/Reviews/Export need
  verification or build).
- Final acceptance — **T1** (exact token extraction) not done.

**Decisions still requiring approval:** T1, C1, R1, G1 (`design-system.md §6.2`);
Q3 (export format, `v1.md` Open Questions); Board drag-and-drop approach.

**Current stage:** **Stage 1 — App Shell.** SPEC + PLAN: `docs/design/screens/app-shell.md`.
