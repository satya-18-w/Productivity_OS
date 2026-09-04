# Screen — Timeline (Week view)

**Reference:** `docs/design/references/timeline-week.png`
**Purpose:** A 7-day grid of time blocks for one week.
**Proposed route:** `/timeline?view=week&date=<in-week>`

Extends `timeline.md` (shell, toolbar, view switcher, date model, right rail).

---

## V1 scope alignment

⚠ **Excluded from V1 — MUST NOT be implemented** (`design-system.md` §6.4). `requirements`
§5 scope boundary: *"One date at a time; no week or month timeline."* Ratified 2026-09-04.

Documented here only so that, if a future requirements revision adds it, it is built to the
shared language — not reinvented.

Post-V1 elements beyond even "a week timeline": KPI row (Tasks 18/24, Focused time 32h15m,
Habits 5/7, Goals 3/5), "Weekly Goals" with per-goal fractions, "Insights" ("You were most
productive on Wednesday…"), "Tip:" coaching copy (VP3 violation).

---

## Layout

- Shell + Timeline toolbar as `timeline.md`; date control shows the week range
  ("31 Aug – 6 Sep 2026"); view switcher = **Week**.
- Optional KPI row (post-V1) — 4 stat cards.
- **Week grid:** left hour axis (06:00–22:00), then **7 day columns** (Mon–Sun — ISO,
  Monday-first). Column header = weekday name + date; **today** column header emphasised
  (dark pill on the date).
- Each column: stacked block cards positioned/ordered by time, category-tinted, showing
  title + time range. Empty time = whitespace with faint hour guide lines.
- Right rail: quote card, mini month calendar, "Weekly Goals", "Habit Tracker" week strip,
  "Insights" (post-V1).

## Screen-specific components

- **Week grid** — CSS Grid: `grid-template-columns: <axis> repeat(7, 1fr)`; hour rows as
  background guides. Blocks are flow children of a column (or absolutely positioned within
  the column if time-proportional height is wanted — container is still a grid cell, VP5).
- **Day column header** — weekday + date, today variant.
- **Compact block card** — title (truncate) + time range only; category fill + left
  border. No checkbox/tags.

## Interactions (inferred)

- Prev/next step by one week; "Today" returns to current week.
- Click a block → edit/delete (same as day). Click a column header / empty slot → add a
  block on that date. Click a day header → drill into Day view for that date.

## Responsive

- Below `laptop`: horizontal scroll of the 7 columns inside `.tl-scroll`, or collapse to a
  vertical day-by-day list (agenda-like). Page never scrolls sideways.
- KPI row wraps 4→2→1; right rail drops first.

## Cannot be inferred / ambiguous

- Whether blocks are time-proportional in height or a simple ordered stack (reference
  looks like a loose stack aligned to start hour).
- How planned vs actual are distinguished in a 7-column grid (the day-view lane split
  doesn't scale — needs a rule: outline vs fill, or a toggle).
- Overlap handling within a column.

## Design-system references

`timeline.md` · §4.3 view switcher · §4.6 KPI card · §4.13 mini calendar ·
`visual-principles.md` VP3, VP5, VP9, VP10.
