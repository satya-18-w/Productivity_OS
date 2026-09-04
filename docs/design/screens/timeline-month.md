# Screen — Timeline (Month view)

**Reference:** `docs/design/references/timeline-month.png`
**Purpose:** A calendar-month grid summarising each day's blocks.
**Proposed route:** `/timeline?view=month&date=<in-month>`

Extends `timeline.md`. **Near-duplicate of `calendar.md`** — see "Overlap" below.

---

## V1 scope alignment

⚠ **Excluded from V1 — MUST NOT be implemented** (`design-system.md` §6.4). `requirements`
§5 scope boundary: *"no week or month timeline."* Ratified 2026-09-04.

Post-V1 rail content: "Monthly Overview" (Tasks 72 / Habits 18 / Goals 6, 73% progress),
"Top Categories" donut, "Upcoming Events" (V1 has no generic events), "Make this month
count" banner.

---

## Layout

- Shell + Timeline toolbar; date control shows the month ("September 2026"); view switcher
  = **Month**.
- **Month grid:** 7 columns (Mon–Sun, ISO), 5–6 week rows. Each **day cell**:
  - date number top-left; **today** circled in `--brand`; days outside the current month
    muted.
  - up to ~3 **event pills** (category glyph + truncated title, category-tinted);
    overflow → "+N more" link.
- Right rail: quote card, "Monthly Overview" stats, "Top Categories" donut, "Upcoming
  Events" list (post-V1).

## Screen-specific components

- **Month grid** — CSS Grid `repeat(7, 1fr)` × auto rows; equal-height cells with internal
  scroll or truncation.
- **Day cell** — header (date, today ring, out-of-month state) + pill stack + overflow
  link.
- **Event pill** — `design-system.md` category-tinted mini row, one line, truncates.

## Interactions (inferred)

- Prev/next step by month; "Today" → current month.
- Click a day → Day view for that date. Click a pill → edit that block. Click "+N more" →
  day view or a day popover. Click empty cell → add a block on that date.

## Responsive

- Below `tablet`: the 7-column grid is unusable squeezed → **fall back to Agenda/list**
  for the month (VP9), or horizontal scroll with min cell width.
- Right rail drops first.

## Overlap with `calendar.md` — ⚠ flag

`timeline-month.png` and `calander.png` are the same artefact with cosmetic differences
(Calendar uses **Sunday-first** — a defect per D8 — plus an "Add Event" button and "Event
Categories"). V1 has **one** time model (planned/actual blocks + categories) and **no**
separate "events" or "calendar" entity. Both this view and a separate Calendar screen are
**excluded from V1** (`design-system.md` §6.4).

## Cannot be inferred / ambiguous

- Whether planned and actual blocks both appear here, and how they'd be distinguished at
  pill size.
- Max pills per cell before overflow (3 shown).
- ~~Week-start~~ — **RESOLVED (D8):** Monday-first / ISO everywhere.

## Design-system references

`timeline.md` · `calendar.md` · §4.3 view switcher · §4.12 donut · §4.13 mini calendar ·
`visual-principles.md` VP7, VP9, VP10.
