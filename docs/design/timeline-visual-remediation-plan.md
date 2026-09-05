# Timeline Visual Remediation — Gap Analysis & Plan

**Date:** 2026-09-05. **Status: PLAN ONLY — no application code changed.** Per this
task's instructions and `CLAUDE.md`, implementation does not start until this plan is
reviewed and approved.

**Targets (visual specification, not inspiration):** `docs/design/references/timeline.png`
(Day), `timeline-week.png` (Week), `timeline-month.png` (Month), `timeline-agenda.png`
(Agenda). **Functional source of truth:** `docs/requirements/v1.md` §3–§6 (planned/actual
blocks, one-date-at-a-time semantics, amended by the G2 decision for Week/Month + a
standalone focus timer).

**Inputs read for this pass:** `CLAUDE.md`; `docs/design/design-system.md` (full);
`docs/design/visual-principles.md` (full); `docs/design/screens/timeline.md`,
`timeline-week.md`, `timeline-month.md`, `timeline-agenda.md`,
`timeline-implementation-plan.md` (a prior, already-shipped audit of this same screen);
the four reference PNGs above; the shared shell (`web/src/shell/{Sidebar,ScreenLayout}.tsx`,
`web/src/styles/tokens.css`); and the current Timeline implementation
(`web/src/features/timeline/**`, `web/src/styles/timeline.css`). No other screens or
backend modules were read beyond `internal/timeline`, `internal/categories`,
`internal/tasks` HTTP contracts (needed to seed realistic fixture data). This follows
`CLAUDE.md`'s frontend context-discipline rule.

**Method:** both dev servers were already running (`:5173` frontend, `:8080` backend). A
throwaway QA account (`timeline-remediation-qa@example.com`, Asia/Calcutta) was
registered through the real UI, then populated **through the real API** (not a frontend
fixture — the backend fully supports everything Timeline needs, confirmed by the prior
audit) with V1-valid categories (Personal/Work/Study/Health/Projects — no colour/icon
set, matching what a real user gets today) and blocks using only the task's suggested
V1-valid activity names (Morning Routine, DSA Practice, Work on Productivity OS, Team
Sync, Lunch Break, Read Research Papers, Build Timeline UI, Workout, Plan Tomorrow, Read
a Book, Sleep) as the **category assigned**, never as a stored title (a time block has no
title field in V1 — confirmed in code and in the prior audit). Density: ~11
planned + 5 actual blocks on the current day, 5–7 blocks/day across the current ISO week,
3 blocks/day across the rest of the month, plus 5 real tasks due today — comparable
density to the references. Screenshots were then taken with Playwright at 1440×900 (all
four views) and 390×844 (Day) against the live app and compared side-by-side against the
four reference PNGs. **This account and its data are QA-only** — flagged here for deletion
after review, same convention as the prior Timeline audit (`timeline-day-audit@example.com`
was deleted; this account should be too, or repurposed if the product owner wants to keep
poking at it).

---

## Executive summary

Most of what the references show that *reads* as "materially different" from the current
build is **already-ratified, intentional V1 exclusion** — confirmed correct in this pass,
not a gap: no personalised greeting, no per-block titles/checkboxes/tags/assignee avatars,
no dashboard-style KPI rows/donuts/"Insights"/"Weekly Goals"/"Habit Tracker"/"Monthly
Overview"/"Upcoming Events" around Week and Month, no "Spaces" sidebar, no search/
notifications. These are documented in `design-system.md §6.4` and were re-confirmed
correct in `timeline-implementation-plan.md` (2026-09-04) and are **re-confirmed correct
again in this pass** — see §7 below. Do not reopen them.

The real, actionable gaps found in this pass are narrower but genuine:

1. **CRITICAL, cross-screen:** category **colour and icon** are real, backend-ready fields
   (`ADR-0009`, `internal/categories`) that the frontend still completely ignores — every
   block/chip on every Timeline view (Day, Agenda, Week, Month) is coloured by a
   deterministic **hash** of the category id (`categoryColor()`), not the colour the user
   actually picked, and **no icon glyph is rendered anywhere**. This is the single biggest
   driver of "looks different from the reference" — the references' most visually
   distinctive language (a consistent icon tile + a chosen colour per category, repeated
   identically across every screen) is entirely absent from the running app for a reason
   that has nothing to do with V1 scope: the data exists, the frontend just never
   consumed it. **This was already found and documented, unfixed, in the prior Timeline
   audit** (`timeline-implementation-plan.md`, "Adjacent finding") — it is still true today.
2. **MINOR, easy:** the view-switcher segment order is `Day · Agenda · Week · Month` in
   the running app vs. the references' `Day · Week · Month · Agenda` everywhere. Purely
   cosmetic, zero functional risk, trivial fix.
3. **MINOR, shared-shell, not Timeline-owned:** the sidebar brand lockup is missing the
   `design-system.md §4.1` tagline ("Plan · Focus · Grow") under the wordmark. Flagged
   for completeness (the task asked to analyze the global shell) but this belongs to
   `screens/app-shell.md`, not a Timeline-scoped fix — see §8.
4. **Everything else checked** (block geometry, planned/actual distinction, category
   tinting, axis, "now" line, comparison table, mini-calendar, responsive shedding,
   Week's 7-column stack, Month's grid + overflow, Agenda's rail + filters) **already
   matches the target visual language within the V1 boundary**, confirmed by direct
   side-by-side comparison in this pass, not just by re-reading the old audit.

No new design tokens are required for anything in scope below. Consuming the existing
`colour`/`icon` category fields (finding #1) does **not** introduce a new token — it
changes which value populates an existing CSS custom property
(`--tl-cat` / `categoryColor()`'s output) and adds an icon *rendering slot* using icons
already in `components/ui/icons`. If a category's `icon` key needs a **new** icon not
already in that set, that one addition would need the Categories-screen owner's sign-off
per `CLAUDE.md` "Design System Changes" — flagged as an open question in §9, not decided
here.

---

## V1 classification legend

Every reference element below is tagged:

- **A** — V1, implement (a genuine gap against ratified scope).
- **B** — V1 visual pattern only (already correctly built, or to be built, without the
  underlying reference feature — e.g. dashed/solid borders standing in for a "status" the
  reference shows via colour).
- **C** — V2/future, do not implement (present in the reference, absent from
  `v1.md`/`design-system.md §6.4`).
- **D** — unsupported/unknown, do not invent.

---

## 1. Cross-cutting: category colour + icon (CRITICAL — affects Day, Agenda, Week, Month)

- **Element:** category colour and icon glyph on every block/chip.
- **Current:** `categoryColor(category_id)` (`web/src/components/productivity/categoryColor.ts`)
  hashes the category's UUID into one of 7 fixed palette hues — **not** the colour the user
  chose when creating the category. No icon is rendered anywhere (`TimelineGrid.tsx`,
  `AgendaList.tsx`, `WeekView.tsx`, `MonthView.tsx` never read a `colour`/`icon` field —
  `web/src/api.ts`'s `Category` type doesn't even declare them).
- **Target:** every block carries a small icon tile (leaf, code brackets, monitor, people,
  fork/knife, book, dumbbell, moon, …) in a category-consistent colour; the same category
  always renders the same colour and icon everywhere it appears.
- **Gap:** the backend already stores and returns a real `colour` and `icon` string per
  category (`internal/categories/{categories,http,service}.go`, `ADR-0009`) — this is
  **not** a reference-only invention, it is shipped, tested backend capability the
  frontend has never wired up. Confirmed still true today by reading current
  `web/src/api.ts` and `categoryColor.ts`.
- **Severity:** CRITICAL (visual) / cross-screen. Not a Timeline-only bug — it belongs to
  the Categories screen + a shared colour/icon consumption point, but it is the largest
  single reason Timeline "looks different" from the reference.
- **V1 class:** **A** (colour) — a category legitimately has a colour field today.
  Icon is **A** for *consuming* the field once a fixed icon set is chosen, but **choosing
  that icon set is a Categories-screen decision**, not Timeline's — see dependency below.
- **Required change (out of this task's direct scope, but a hard dependency):**
  1. Extend `web/src/api.ts`'s `Category` interface with `colour`/`icon` (already returned
     by `GET /api/categories`, just not typed).
  2. Decide the fixed colour-key → CSS-token mapping and the icon-key → icon-component
     mapping (a **Categories-screen + shared-component** decision, per `CLAUDE.md`
     "Design System Changes" if any new icon is needed beyond the existing set in
     `components/ui/icons`).
  3. Once that mapping exists, swap `categoryColor()`'s call sites (or the function
     itself) to prefer the category's real `colour` key, falling back to the current hash
     only for the "Uncategorized" bucket. Add an icon-tile slot to `TimelineGrid`'s block,
     `AgendaList`'s row, `WeekView`'s chip, `MonthView`'s chip.
  4. Update the Categories screen's create/edit form to offer colour + icon pickers (not
     built yet — confirmed by reading `web/src/features/categories/`).
- **Reusable component involved:** `categoryColor()` (or its replacement), a new shared
  `CategoryIcon`/`CategoryGlyph` component consumed by all four Timeline views + Categories.
- **Design token involved:** none new — the existing `--cat-*` tokens in `tokens.css`
  become the *target* of a per-category choice instead of a hash output. No token value
  changes.
- **Functional/backend dependency:** **none** — backend is already done (ADR-0009). This
  is a pure frontend consumption gap.
- **Acceptance criterion:** a category created with a chosen colour/icon renders that
  exact colour/icon consistently on Day, Agenda, Week, and Month for every block in that
  category; the Uncategorized bucket keeps its neutral grey; no business logic reads
  colour or icon (D2 unchanged).
- **Recommendation:** treat as its **own follow-up task** scoped "Categories colour/icon
  picker + Timeline consumption," per the prior audit's same recommendation. Do **not**
  fold it into this Timeline-only remediation pass without an explicit go-ahead, since it
  necessarily touches the Categories screen (`CLAUDE.md`: don't modify unrelated
  screens). Flagging it here because it is the single highest-leverage visual fix
  available and the task instructions ask to record backend-ready capability the frontend
  ignores.

---

## 2. Day view — element-by-element

### 2.1 Global application shell (shared, not Timeline-owned)

- **Current:** sidebar ~248px (`--sidebar-w`), icon+label nav (Timeline/Tasks/Board/
  Habits/Goals/Categories/Reports/Reviews), brand tile + "Productivity OS" wordmark, no
  tagline, footer = theme toggle + user email chip. Main content fluid. Right rail 320px
  (`--rail-w`), sheds below `wide`.
- **Target:** same three-region shape; sidebar additionally shows a "Plan · Focus · Grow"
  tagline under the wordmark, a "SPACES" category-switcher list (excluded, C1/§6.4), a
  decorative nature-art quote card above the user chip (D6, optional), and a user chip
  showing a display name + "Free Plan" (no plan/billing concept in V1 — correctly absent).
- **Gap:** missing tagline only (MINOR). SPACES list, decorative card, plan label are
  either explicitly excluded already or optional (D6) — not gaps.
- **Severity:** MINOR.
- **V1 class:** B (tagline is legitimate shell copy, not a scope question).
- **Required change:** add the tagline span to `Sidebar.tsx`'s brand lockup.
- **Component:** `web/src/shell/Sidebar.tsx`.
- **Token:** none new.
- **Dependency:** none.
- **Acceptance criterion:** sidebar brand block shows wordmark + tagline, matching
  `design-system.md §4.1`.
- **Ownership note:** this is a **shared-shell** fix (`screens/app-shell.md`'s territory,
  per VP6 "sidebar is global, one implementation"), not a Timeline-specific change. Listed
  here only because the task instructions asked to analyze the global shell; implement it
  in the "shared shell/layout corrections" step (§10 implementation order), not inside any
  Timeline file.

### 2.2 Main content width / page margins

- **Current:** matches shell grid; `ScreenLayout` main column + rail, `--gutter` (24px)
  outer margins. ✅ matches target proportions.
- **Severity:** none (no gap).

### 2.3 Right contextual rail (Day)

- **Current:** `MiniCalendar` → `PomodoroCard` ("Focus timer") → `TodayTasks` ("Today's
  tasks", hides itself when empty). Order top-to-bottom: calendar, timer, tasks.
- **Target:** mini-calendar → "Today's Tasks" → "Focus Mode" (Pomodoro) → quote card.
  Order in the reference is calendar → tasks → timer; current is calendar → timer → tasks.
- **Gap:** rail widget **order** differs from the reference (tasks and timer swapped);
  content itself is correct (both are V1-legitimate — timer per G2, tasks per §5).
- **Severity:** MINOR (cosmetic ordering only, no missing/extra widget).
- **V1 class:** B.
- **Required change:** reorder the `rail` JSX in `TimelineScreen.tsx` so `<TodayTasks>`
  renders before `<PomodoroCard>`.
- **Component:** `web/src/features/timeline/TimelineScreen.tsx` (the `rail={...}` prop).
- **Token:** none.
- **Dependency:** none.
- **Acceptance criterion:** rail order is MiniCalendar → Today's Tasks → Focus timer.
- **Also confirmed correct, no action:** no "Agenda Overview" donut, no motivational
  quote card (D6, optional — omitted like every other screen, consistent with VP3).

### 2.4 Header — eyebrow / title / subtitle / primary action

- **Current:** eyebrow "TIMELINE" (screen-name convention, shared by every V1 screen);
  title = full date ("Saturday, September 5, 2026"); subtitle = factual description;
  primary action = `SplitButton` "Add block ▾".
- **Target:** date line + personalised greeting ("Good morning, Satyajit 👋") + a rotating
  motivational quote in the subtitle position; primary action "+ Add ▾".
- **Gap:** none — this is a **already-ratified, deliberate divergence** (VP3: honest over
  motivational; a full date beats a fake "good morning" that doesn't know the time of
  day). Documented as correct in both `screens/timeline.md` and the prior audit.
- **Severity:** n/a (not a gap).
- **V1 class:** C (personalised greeting / rotating quote are reference-only decoration,
  §6.4 "no fake or adaptive encouragement").

### 2.5 View switcher

- **Current:** segment order `Day · Agenda · Week · Month` (`TimelineScreen.tsx`
  `VIEW_OPTIONS`).
- **Target:** segment order `Day · Week · Month · Agenda` (all four Timeline reference
  images agree on this order).
- **Gap:** cosmetic ordering mismatch only — all four segments exist, all four route
  correctly, active-segment styling already matches `design-system.md §4.3` (brand fill).
- **Severity:** MINOR.
- **V1 class:** B.
- **Required change:** reorder the `VIEW_OPTIONS` array.
- **Component:** `web/src/features/timeline/TimelineScreen.tsx` line ~42.
- **Token:** none.
- **Dependency:** none.
- **Acceptance criterion:** segmented control reads Day/Week/Month/Agenda left to right,
  matching every reference image; no behavioural change (URL params unaffected — the
  `view` param values themselves don't change, only display order).

### 2.6 Date navigator / Today button

- **Current:** `‹` `IconButton` + native `<input type=date>` pill + `›` + "Today" button
  (`DateStepper`), disabled correctly on today.
- **Target:** `‹` + date pill with a calendar icon (opens a picker) + `›` + "Today".
- **Gap:** current uses the browser's native date input (shows a platform-specific
  calendar icon and picker chrome) instead of a custom-styled pill; functionally
  equivalent, visually slightly different chrome (native OS date-picker vs. a bespoke
  dropdown).
- **Severity:** MINOR (already flagged and accepted as a reasonable native-control choice
  in earlier phases; revisiting would mean building a custom date-picker component —
  disproportionate for a cosmetic delta given the mini-calendar in the rail already
  covers "pick any date" with the target's exact visual language).
- **V1 class:** B.
- **Required change:** none recommended — documenting as an accepted, not a real, gap.
- **Acceptance criterion:** n/a.

### 2.7 Timeline axis / grid / time labels

- **Current:** fixed-width hour gutter, tick every 3h (00/03/06/09/12/15/18/21/24),
  tabular-nums, full 00:00–24:00 scrollable range, auto-scrolls to "now" (today) or the
  earliest block (non-today) on load.
- **Target:** hour axis 06:00–22:00 shown (reference crops to the "active" hours);
  per `screens/timeline.md`, full 00:00–24:00 with a sensible default scroll position was
  **already ratified** as the correct V1 resolution (the reference's cropped range isn't a
  requirement, just what the mock happened to show).
- **Gap:** none — matches the ratified resolution. ✅.
- **Severity:** n/a.

### 2.8 Current-time indicator

- **Current:** a horizontal red line across both lanes at the current minute, with a
  time-pill ("10:18") on the Planned lane only, `z-index` above blocks.
- **Target:** a dark pill with the time, sitting on the axis, with a connecting line.
- **Gap:** colour differs (current uses `--danger`-style red; reference uses a dark
  neutral pill) — worth checking against `--text`/`--brand` tokens for consistency, but
  this is a legibility-positive choice (red reads unambiguously as "current position",
  consistent with how a live indicator is typically coloured) and no token is missing to
  do it either way.
- **Severity:** MINOR (a colour-language nit, not a structural gap).
- **V1 class:** B.
- **Required change:** optional — consider changing the "now" line/pill colour from red to
  a neutral dark (`--text` or `--brand`) fill to match the reference's calmer language,
  since VP1 discourages using a semantic-danger colour for a non-error, non-destructive
  indicator (red usually means "problem" elsewhere in this system, e.g. negative
  difference in the comparison table).
- **Component:** `web/src/styles/timeline.css` (`.tl2__now`-style rules).
- **Token:** likely `--text`/`--brand` instead of `--danger` — needs confirming which
  class currently drives the colour before changing it.
- **Dependency:** none.
- **Acceptance criterion:** the "now" indicator is visually distinct from the
  planned/actual difference colour language (which legitimately uses red/green for
  negative/positive).
- **Also confirmed, already known and correctly deferred (not re-litigated here):** the
  "now" line can cross an active block's text — documented and deliberately deferred in
  `timeline-implementation-plan.md` as a MINOR cosmetic issue needing a larger structural
  fix (split line/pill stacking). Still true; still deferred; not re-scoped by this task.

### 2.9 TimeBlock geometry / colours / typography

- **Current:** time-proportional height, single-line content ("06:00–06:45 · Personal"),
  category-tinted fill + dashed(planned)/solid(actual) border, ellipsizes on overflow.
- **Target:** same geometry model; additionally a leading icon tile per block (see §1) and
  (reference-only) a checkbox, tag chips, assignee avatars.
- **Gap:** geometry/typography/colour-application mechanism already matches (G1, ratified).
  The only real gap is the icon (§1, cross-cutting) — checkbox/tags/avatars are correctly
  absent (C, not V1).
- **Severity:** see §1 for the icon; otherwise none.

### 2.10 TimeBlock category / Task↔TimeBlock relationship

- **Current:** a task-linked block shows "↳ `<task title>`" + a thicker left border
  (built in a 2026-09-05 follow-up, confirmed present in code).
- **Target:** not shown in the reference at all (the reference has no task-linking
  concept).
- **Gap:** none — this is an in-scope V1 feature correctly built beyond what the
  reference shows; no visual conflict.
- **Severity:** n/a.

### 2.11 Planned column / Actual column

- **Current:** two labelled lane headers "PLANNED"/"ACTUAL" above the grid, equal-width
  columns.
- **Target:** the reference shows one merged list, not two lanes — **already resolved in
  favour of the stricter §5 requirement** (two lanes, or another clearly distinct
  treatment) in the original Phase 2 spec. Confirmed still correct.
- **Severity:** n/a (ratified, deliberate deviation).

### 2.12 Empty / loading / error states

- **Current:** `.tl2__empty`-style hint per empty lane; "Loading…" text; `ErrorState` +
  Retry on failure.
- **Target:** not shown (references never depict these states).
- **Gap:** none to compare against — current states were spot-checked in this pass by
  viewing a category with a zero-block window (e.g. scrolling to a normally-empty hour)
  and match the existing, already-accepted pattern.
- **Severity:** n/a.

### 2.13 Cards / borders / shadows / radius / spacing / typography (visual system)

- **Current:** tokens-only throughout `timeline.css` (`--sp-*`, `--radius-xs/md`,
  `--border`/`--border-strong`, no shadow on the dense grid, `--shadow-sm` on `Card`/
  `Dialog`). Confirmed by reading the stylesheet — no arbitrary literals found.
- **Target:** consistent with `design-system.md §3`.
- **Gap:** none found.
- **Severity:** n/a.

### 2.14 Accessibility / hover / focus / active states

- **Current:** each block is a `<button>` with a full accessible name; lanes have
  `aria-label`s; `:focus-visible` ring present (shared primitive); comparison table has
  real `<th>` + sign glyphs, not colour-only.
- **Gap:** not re-audited byte-for-byte in this pass (out of this task's visual-diff
  scope) but nothing observed regressed; the prior audit's a11y checklist pass is assumed
  still valid since no a11y-relevant markup changed since.
- **Severity:** n/a — flag for a full re-check only if implementation touches these files.

### 2.15 Responsive (Day, 390×844)

- **Current:** header/toolbar/view-switcher/date-stepper stack correctly; two-lane grid
  scrolls horizontally **inside its own container**, not the page (confirmed: page body
  itself doesn't scroll sideways at 390px — only the lane-grid's inner scroll container
  does, matching VP9); rail (mini-calendar, focus timer, today's tasks) stacks below main.
- **Target:** VP9 shed order (rail first, then sidebar labels, then sidebar drawer);
  wide grids scroll inside their own container, page never scrolls sideways.
- **Gap:** none found — matches VP9. (A full-page screenshot at 390px clips block text at
  the container's right edge before scrolling, which looks alarming in a static image but
  is expected inner-scroll-container behaviour, not a bug — same "screenshot artifact"
  class already documented in the prior audit for a different element.)
- **Severity:** n/a.

---

## 3. Agenda view — element-by-element

Agenda was the most recently, most thoroughly re-audited view (`timeline-agenda.md`
"Reference-accuracy audit", 2026-09-04) and this pass's own screenshot comparison found
**no new gaps** beyond the cross-cutting icon/colour issue in §1:

- Category filter chips with counts ✅ matches target.
- Time-range rail with connecting line + node dots ✅ matches target (this was the subject
  of the prior audit's own fix).
- Row = time range + category name + Planned/Actual badge (text + tone, not colour alone)
  ✅ matches target's row shape, minus the correctly-excluded checkbox/tags/avatars/"Top
  Priorities" (C, §6.4).
- "Add an agenda item" footer row ✅ present, matches target's add affordance language.
- Right rail: mini-calendar only in the current build; **the target additionally shows an
  "Agenda Overview" donut and a "Time Allocation" bar**, both **already correctly excluded**
  per `screens/timeline-agenda.md`'s own deferred-items note ("kept the full
  `ComparisonCard` below, shared with Day — simpler, consistent") — the `ComparisonCard`
  under the list is V1's honest equivalent of "Time Allocation." Not a gap.

No action items for Agenda beyond §1.

---

## 4. Week view — element-by-element (additions beyond §2/§3's shared items)

### 4.1 Seven-day grid / day headers

- **Current:** `grid-template-columns: repeat(7, 1fr)`, Monday-first, weekday abbreviation
  + date number per header, today's header visually distinct (`is-today` class), clicking
  a header jumps to Day for that date. Horizontal scroll container (`.tl-week__scroll`)
  auto-anchors to today's column on load.
- **Target:** 7 columns, Monday-first, today emphasised, matches `MiniCalendar`'s "today"
  treatment.
- **Gap:** none found — matches.
- **Severity:** n/a.

### 4.2 Activity density / compact chips / overflow

- **Current:** every block in a day renders as its own chip in a chronological stack — no
  cap, no "+N more" in Week (by design — `screens/timeline-week.md` specifies a full
  stack, not a capped list; overflow/"+N more" is a **Month**-only pattern).
- **Target:** same — the Week reference also shows a full stack per day, capped only by
  available vertical space in the mock; Month is the one with an explicit "+N more".
- **Gap:** none.
- **Severity:** n/a.

### 4.3 Contextual information hierarchy / right-rail composition (Week)

- **Current:** right rail = mini-calendar only (`TimelineScreen`'s rail is shared across
  all views; `PomodoroCard` is Day-only via `view === "day" &&`).
- **Target:** reference's rail carries a KPI row + "Weekly Goals" + "Habit Tracker" +
  "Insights" — **all correctly excluded** (§6.4, unchanged by the G2 amendment, confirmed
  again in this pass).
- **Gap:** none (correct exclusion, not a gap).
- **Severity:** n/a.
- **V1 class:** C for all four excluded widgets.

### 4.4 Metric cards

- **Current:** none rendered on Week/Month.
- **Target:** 4-up KPI card row ("18/24 Tasks completed", "32h15m Focused time", "5/7
  Habits done", "3/5 Goals progress").
- **Gap:** none — correctly excluded. "Focused time" specifically has **no V1 measure**
  (no focus-time tracking exists outside the standalone, non-persisted Pomodoro timer);
  period totals like "Tasks completed" would imply a range-aggregation V1 doesn't compute
  for Tasks. Confirmed still correctly absent.
- **Severity:** n/a.
- **V1 class:** C.

### 4.5 Selected day / today treatment

- **Current:** today's column header gets an `is-today` class (distinct fill, per the
  screenshot: a filled circle matching the mini-calendar's today-ring language).
- **Target:** today's header similarly emphasised.
- **Gap:** none.
- **Severity:** n/a.

---

## 5. Month view — element-by-element (additions beyond §2/§3's shared items)

### 5.1 Monthly grid / day cells

- **Current:** 7×6 CSS grid, Monday-first, today ringed (`tl-month__today-ring`), outside-
  month days muted (`tl-month__cell--outside`), up to 3 chips + "+N more" overflow,
  horizontal scroll container (`.tl-month__scroll`) on narrow widths, auto-anchors to
  today's cell.
- **Target:** same shape exactly (`screens/timeline-month.md` §"Resolved").
- **Gap:** none found.
- **Severity:** n/a.

### 5.2 Chip content / overflow

- **Current:** chip text = `"<time> <category>"` (e.g. "06:00 Personal"), dashed/solid
  border per G1; "+N more" jumps to Day.
- **Target:** same (no per-block titles — a block has no title field).
- **Gap:** none beyond the cross-cutting icon/colour issue (§1) — the reference's chips
  also carry a tiny icon.
- **Severity:** covered by §1.

### 5.3 Right-rail composition (Month)

- **Current:** mini-calendar only.
- **Target:** "Monthly Overview" stat row + "Top Categories" donut + "Upcoming Events" +
  a "Make this month count" banner.
- **Gap:** none — all four correctly excluded. "Upcoming Events" specifically implies a
  generic calendar-event entity that **does not exist in V1** (`design-system.md §6.4`
  "Calendar as a separate feature... and any generic 'event' entity"); a donut of "Top
  Categories" by task count conflates Tasks with Timeline blocks and isn't a defined V1
  report. Confirmed still correctly absent.
- **Severity:** n/a.
- **V1 class:** C for all.

### 5.4 Overlap with the excluded "Calendar" screen

- **Current:** Month shows **time blocks only** (planned/actual, category-tinted);
  clicking a cell/chip navigates to Day, never opens a generic "event" editor.
- **Target concern:** `timeline-month.png` and the separately-excluded `calander.png` are
  near-duplicates; the live risk is scope creep toward a generic Calendar/event screen.
- **Gap:** none — confirmed Month still only renders blocks, no event concept was
  introduced anywhere in the current code.
- **Severity:** n/a (a standing caution, not a defect).

---

## 6. Responsive summary across views (spot-checked at 1440×900 and 390×844)

| Viewport | Day | Week | Month |
|---|---|---|---|
| 1440×900 | ✅ matches, rail present | ✅ 7 columns, horizontal inner scroll only | ✅ 7×6 grid, no page h-scroll |
| 390×844 | ✅ rail stacks below, inner lane-scroll only | not re-shot this pass (already verified in the prior Week audit at 390px — no regressions expected since `WeekView.tsx` is unchanged since that audit) | not re-shot this pass (same reasoning — `MonthView.tsx` unchanged since its own prior 390px verification) |

No new responsive regressions found at the viewports actually captured in this pass. Full
5-viewport (1440/1280/1024/768/390) re-verification is scheduled as part of the mandated
Browser Visual QA step **after** implementation, per this task's own instructions — not
required to complete the plan itself.

---

## 7. V1 scope re-confirmation (explicitly re-checked against the references this pass, all correct as-is — do not touch)

| Reference element | Views it appears in | V1 class | Status |
|---|---|---|---|
| Personalised greeting ("Good morning, Satyajit 👋") | Day, Week, Agenda | C | Correctly absent — factual title used instead (VP3) |
| Rotating motivational quote in header/subtitle | All four | C | Correctly absent |
| Per-block checkbox ("mark done") | Day, Week (implied), Agenda | D | Correctly absent — a time block has no completion state |
| Tag chips on blocks ("Deep Work", "LeetCode") | Day, Agenda | C | Correctly absent — no tags field on a block |
| Assignee avatars | Day ("Team Sync") | C | Correctly absent — no collaboration in V1 |
| "Focus Mode" widget beyond the standalone timer | Day rail | B | Built as a standalone, non-persisted timer per G2 — matches the amendment exactly |
| KPI/metric row | Week, Month | C | Correctly absent |
| "Weekly Goals" / "Habit Tracker" cross-widgets | Week rail | C | Correctly absent |
| "Insights" / "Tip:" coaching copy | Week rail | C | Correctly absent |
| "Monthly Overview" / "Top Categories" donut / "Upcoming Events" / banner | Month rail | C | Correctly absent |
| Generic Calendar / "event" entity | Month (via `calander.png` overlap) | C | Correctly absent — Month renders blocks only |
| "Sort by" / grid-view toggle | Agenda | C (plausible UI, not required) | Correctly absent |
| "Top Priorities" list | Agenda rail | D | Correctly absent — no V1 backing |
| "Agenda Overview" donut / "Time Allocation" bar | Agenda rail | B | `ComparisonCard` (shared with Day) is the honest V1 equivalent — already built |
| Search bar, notification bell | Global top bar (all references) | C | Correctly absent |
| "SPACES" sidebar switcher | Global sidebar | C | Correctly absent (C1 not ratified) |

---

## 8. Implementation order (per this task's mandate)

1. **Shared shell/layout corrections** — sidebar tagline (§2.1) only; nothing else in the
   shared shell showed a gap in this pass.
2. **Shared design-token corrections** — none required. No new token, no changed token
   value is needed for anything found in this pass (the category colour/icon fix in §1
   reassigns which *value* fills an existing token slot; it does not add or change a
   token).
3. **Shared calendar/date-navigation components** — none required; `MiniCalendar` and
   `DateStepper` both already match their specs.
4. **Shared task/time-block presentation components** — this is where §1 (category
   colour/icon) and the rail-order fix (§2.3) land, plus the view-switcher order fix
   (§2.5), since `VIEW_OPTIONS` and the rail JSX both live in the shared
   `TimelineScreen.tsx`.
5. **Timeline Day** — the "now"-indicator colour reconsideration (§2.8), optional.
6. **Timeline Agenda** — no changes identified beyond the shared fixes above.
7. **Timeline Week** — no changes identified beyond the shared fixes above.
8. **Timeline Month** — no changes identified beyond the shared fixes above.

Given how few and how small the confirmed gaps are, steps 5–8 mostly resolve to "verify
the shared fix rendered correctly on this view" rather than per-view code changes.

---

## 9. Open questions requiring product-owner approval before implementation

1. **Scope of the category colour/icon fix (§1):** implement now as part of this
   remediation pass (touches the Categories screen too, beyond `CLAUDE.md`'s "don't touch
   unrelated screens" guidance for a Timeline-scoped task), or split into its own
   follow-up task and ship the smaller Timeline-only fixes (§2.1, §2.3, §2.5) here? This
   plan **recommends splitting it out**, consistent with the prior audit's own
   recommendation, but flags it since it's the highest-visual-impact single change
   available.
2. **Icon set:** if #1 proceeds, the icon-key → icon-component mapping is a new design
   decision (which icons represent which category) — needs explicit sign-off before any
   icon renders, per `CLAUDE.md` "Design System Changes."
3. **"Now" indicator colour (§2.8):** change from red to a neutral dark tone, or leave as
   is? Low stakes either way; included as a judgment call, not a blocker.

Everything else in this plan (§2.1 tagline, §2.3 rail order, §2.5 view-switcher order) is
low-risk, no-new-token, no-scope-change cosmetic alignment and can proceed without further
discussion once this plan is accepted.

---

## 10. Fixture/reference data used (for reproducibility, and cleanup tracking)

- Account: `timeline-remediation-qa@example.com` (Asia/Calcutta) — **QA-only, recommend
  deleting after this plan is reviewed**, same convention as the prior Timeline audit.
- Categories (no colour/icon set — matches what the frontend actually offers a real user
  today): Personal, Work, Study, Health, Projects.
- Blocks: today (2026-09-05) got the full V1-valid activity list as **categories**
  (Morning Routine→Personal, DSA Practice→Study, Work on Productivity OS→Projects, Team
  Sync→Work, Lunch Break→Personal, Read Research Papers→Study, Build Timeline UI→
  Projects, Workout→Health, Plan Tomorrow→Personal, Read a Book→Personal, Sleep→Health,
  midnight-spanning) plus 5 actual blocks with slightly shifted times (for a non-trivial
  planned-vs-actual comparison). The rest of the current ISO week got 5–7 planned blocks/
  day from a rotating subset of the same list. The rest of the current month got 3
  planned blocks/day. Tasks: 5 real tasks due today, reusing the same V1-valid titles, to
  populate the "Today's tasks" rail widget.
- **No unsupported metadata used** anywhere: no titles on blocks (V1 has none), no
  priorities/tags/assignees, no Focus/Pomodoro persistence, no generic "events."

---

## 11. Summary for the product owner

- **Plan file:** this document.
- **Files that will change, if this plan is approved as-is (§8, minus the split-out §1):**
  `web/src/features/timeline/TimelineScreen.tsx` (view-switcher order, rail order),
  `web/src/shell/Sidebar.tsx` (tagline), optionally `web/src/styles/timeline.css` (now-line
  colour).
  Nothing else — no `WeekView.tsx`, `MonthView.tsx`, `AgendaList.tsx`, `TimelineGrid.tsx`,
  `ComparisonCard.tsx`, `BlockDialog.tsx`, or backend changes are needed for anything found
  in this pass.
- **Not implemented, and why:** the category colour/icon fix (§1) is deliberately left
  out of "ready to implement now" pending the scope decision in §9.1 — it is real,
  valuable, and backend-ready, but touches the Categories screen, which is outside a
  Timeline-scoped task per `CLAUDE.md`.
- **No functional/backend work required** for any Timeline-scoped item in this plan.
- **No new design tokens.**
- **Waiting for:** explicit approval of this plan (and an answer to §9's three questions)
  before any application code is touched, per this task's own instructions.
