# Screen — Timeline (Agenda view)

**Reference:** `docs/design/references/timeline-agenda.png`
**Purpose:** A chronological list of one day's time blocks.
**Proposed route:** `/timeline?view=agenda&date=<today>`

Extends `timeline.md`.

---

## V1 scope alignment

Agenda is a **list rendering of a single day** → compatible with `requirements` §5 ("one
date at a time"). This is the most V1-friendly of the four timeline views and a good
mobile fallback for Day.

| Reference element | V1 status |
|---|---|
| Vertical rail with time-range nodes + block rows | in scope |
| Category filter chips (All (8) / Personal (2) / …) | in scope (filter a single day's blocks by category) |
| Sort by: Time / list-vs-grid toggle | plausible UI, not required |
| Per-row checkbox | **not V1** (blocks have no done state) |
| Tag chips, assignee avatars | **not V1** |
| Rail: "Agenda Overview" donut, "Time Allocation", "Top Priorities" | Time-per-category for the day = **§6 (planned vs actual comparison)** — legitimate V1 data; "Priorities" is not V1 |

**Recommendation:** build Agenda as the day-list view. Keep: time-range rail, category
filter chips, block rows (time, title, category). The rail's per-category time totals can
surface the **§6 planned-vs-actual comparison** for that date. Drop checkboxes / tags /
avatars / priorities.

---

## Layout

- Shell + Timeline toolbar; view switcher = **Agenda**; date stepper as Day.
- Optional decorative quote strip (VP3 — omit for honest V1).
- **Filter-chip row** (§4.4): "All (n)" + one chip per category present that day, with
  counts; active chip filters the list.
- Right: "Sort by" dropdown + list/grid toggle (optional).
- **Agenda list:** left time-range column (`06:00 – 06:45`) with a connecting rail + node
  dot per item (dot colour = category); right = **block row** card (category glyph, title,
  category name chip, category-tinted left border / faint fill).
- Right rail: "Agenda Overview" donut (blocks per category), **"Time Allocation"** = §6
  per-category totals (planned vs actual), quote card.

## Screen-specific components

- **Agenda rail** — flex row: fixed-width time column + vertical connector + content;
  node dot per item.
- **Block row** — reuses the list-row pattern (§4.8) with a category-tinted surface.
- **Add row** — a persistent "＋ Add an agenda item" affordance at the list foot →
  create planned/actual block (§4.16, fields: start/end/category).

## Interactions

- Filter chips narrow the visible blocks. Sort (Time is the only meaningful V1 sort).
- Row click → edit block; delete from the row/kebab.
- Date stepper / mini-calendar change the day.

## Responsive

- Single column already; rail drops first. Time column narrows; filter chips scroll.
- This view is the recommended **mobile fallback** for Day and Week.

## Cannot be inferred / ambiguous

- Whether "grid" toggle shows a different layout (masonry of block cards?).
- Whether planned and actual are both listed and how ordered when times coincide.
- "Top Priorities" has no V1 backing.

## Design-system references

`timeline.md` · §4.3 view switcher · §4.4 filter chips · §4.8 list row · §4.12 donut ·
§4.16 create/edit form · `requirements` §6 · `visual-principles.md` VP3, VP7, VP9, VP10.

---

## Phase 3 — Timeline (Agenda) — Status: ✅ COMPLETE (2026-09-04)

Built as a view of the shared **`TimelineScreen`** (`?view=agenda`), alongside Day.

- [x] **View switcher** (`SegmentedControl` Day / Agenda) added to the shared Timeline
      toolbar; `?view=` URL param (omitted when "day").
- [x] `TimelineDay` refactored → **`TimelineScreen`** — shared data load, date/view params,
      block dialog, mini-calendar rail, `ComparisonCard`; body switches Day grid ↔ Agenda.
- [x] **`AgendaList`** — planned + actual **merged, sorted by start time**; each row: time
      range (tabular, ▲/▼ markers) + category dot + name + **Planned / Actual `Badge`**
      (§5 distinction by text + tone, not colour alone). Past rows dimmed.
- [x] **Category filter** — `Chip` row: "All (n)" + one per category present that day with
      counts + colour dot; single-select, click active to clear.
- [x] `<ol>` list; each row a `<button>` with a full accessible name → opens the block
      dialog (same `BlockDialog` as Day). Empty + filtered-empty states.
- [x] Responsive — rows wrap (time on its own line) `< 480px`; chips scroll; **Agenda is
      the practical mobile view for a day**. No page h-scroll.
- [x] Tests — `AgendaList` (5) + `TimelineScreen` view-switch/param tests. 88 total green.
- [x] Browser-verified — desktop / mobile / dark; filter + Day↔Agenda switch; no console
      errors. Matches `references/timeline-agenda.png` visual language **minus** the
      excluded checkboxes / tags / avatars / "Top Priorities" / donut / list-grid toggle.
- [ ] Committed — pending product owner.

### Deferred
- Rail "Time allocation" mini-summary (kept the full `ComparisonCard` below, shared with
  Day — simpler, consistent).
- `list | grid` toggle (reference-only affordance, no V1 need).
- Multi-select category filter (single-select is enough for a single day).
