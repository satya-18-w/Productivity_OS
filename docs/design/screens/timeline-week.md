# Screen — Timeline (Week view)

**Reference:** `docs/design/references/timeline-week.png`
**Purpose:** A 7-day grid of time blocks for one week.
**Proposed route:** `/timeline?view=week&date=<in-week>`

Extends `timeline.md` (shell, toolbar, view switcher, date model, right rail).

---

## V1 scope alignment

**Approved 2026-09-05 (product owner, G2) — was excluded 2026-09-04, see `v1.md §5`
amendment.** The week-of-blocks rendering itself is in scope. The reference's
dashboard-style surroundings are **still excluded** (`design-system.md §6.4`, unchanged
by this amendment): KPI row (Tasks 18/24, Focused time 32h15m, Habits 5/7, Goals 3/5),
"Weekly Goals" with per-goal fractions, "Habit Tracker" week strip, "Insights" ("You were
most productive on Wednesday…"), "Tip:" coaching copy (VP3).

---

## Layout

- Shell + Timeline toolbar as `timeline.md`; date control steps by **week** and shows the
  range ("31 Aug – 6 Sep 2026"); view switcher gains **Week**.
- **No KPI row** (§6.4, unchanged).
- **Week grid:** **7 day columns** (Mon–Sun — ISO, Monday-first, D8), no hour axis (G2 —
  a week is too dense for hour-proportional height; see resolution below). Column header
  = weekday name + date; **today**'s header emphasised (matches `MiniCalendar`'s "today"
  treatment).
- Each column: a **chronological stack** of that day's blocks — planned and actual
  together, sorted by start time (same merge `AgendaList` already does for one day).
  Category-tinted, dashed border = planned / solid border = actual (G1's language,
  reused exactly — no new visual system for "planned vs actual").
- Right rail: mini-calendar only, same as Day. **No** "Weekly Goals" / "Habit Tracker" /
  "Insights" (§6.4, unchanged).

## Screen-specific components

- **Week grid** — CSS Grid: `grid-template-columns: repeat(7, 1fr)`, normal flow (no
  absolute positioning — VP5 — since blocks are no longer time-proportional here).
- **Day column header** — weekday + date, today variant (reuses the day-cell language
  from `MiniCalendar`).
- **Compact block chip** — time range + category name only (no titles — a block has no
  title field, `v1.md §3/§4`); category fill + dashed/solid border per G1. Click → the
  same `BlockDialog` used by Day/Agenda.

## Interactions

- `‹`/`›` step by one ISO week; "Today" returns to the week containing today.
- Click a block chip → edit/delete (`BlockDialog`, unchanged).
- Click a day column's header → jumps to the Day view for that date (`?view=day&date=`).
- "Add block" (the existing split-button) still adds to **the currently selected single
  date** — Week doesn't need its own per-column add affordance; clicking a column header
  to jump into Day is the add path for a different day than the one currently selected.

## Responsive

- Below `laptop`: the 7 columns scroll horizontally inside `.tl-week__scroll` (same
  pattern as `.tl2__scroll`). Page never scrolls sideways.
- Right rail drops first (unchanged shell behaviour).

## Resolved (was "cannot be inferred")

- **Time-proportional vs. stack**: a stack, not proportional (G2) — 7 columns of
  hour-scaled blocks would be far too dense/narrow to read; a stack keeps every block's
  category + time legible at any zoom.
- **Planned vs actual in 7 columns**: reuse G1's dashed/solid border exactly — no new
  rule needed.
- **Overlap within a column**: blocks are allowed to overlap (§3/§4 — "overlaps are
  allowed and not flagged"); the stack renders them in start-time order same as Agenda,
  it does not attempt to detect or lay out overlaps side-by-side.

## Design-system references

`timeline.md` · §4.3 view switcher · §4.6 KPI card · §4.13 mini calendar ·
`visual-principles.md` VP3, VP5, VP9, VP10.

---

## Status: ✅ COMPLETE (2026-09-05)

`web/src/features/timeline/WeekView.tsx`. Real data — one `api.timelineRange(from,
to)` call for the whole week (was 7 parallel `api.timeline(date)` calls; the backend
range read landed 2026-09-05, see `docs/left.md`).

- [x] 7 day-columns, Monday-first (D8), today's header emphasised.
- [x] Chronological stack per column, dashed=planned/solid=actual (G1 reused exactly).
- [x] Click a chip → `BlockDialog` (shared with Day/Agenda). Click a column header → jump
      to Day view for that date.
- [x] **Found and fixed, not a design choice:** at common desktop widths the 7-column
      grid needs more room than the main content area (rail included) gives it, so it
      scrolls horizontally — without an anchor this silently opened on Monday with
      today's column off-screen. Added an auto-scroll to today's column on load, mirroring
      the Day grid's vertical anchor.
- [x] No KPI row, donut, "Weekly Goals"/"Habit Tracker"/"Insights" (§6.4, unchanged).
- [x] Responsive — verified 1440px (scrolls, today centred) and 390px (no page h-scroll,
      inner horizontal scroll only); no console errors either width.
- [x] Tests — 5 (fetch scope/Monday-first, planned/actual chip classes, empty-day
      placeholder, chip click → onPick, header click → onJumpToDay).
- [ ] Committed — pending product owner.
