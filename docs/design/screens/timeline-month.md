# Screen — Timeline (Month view)

**Reference:** `docs/design/references/timeline-month.png`
**Purpose:** A calendar-month grid summarising each day's blocks.
**Proposed route:** `/timeline?view=month&date=<in-month>`

Extends `timeline.md`. **Near-duplicate of `calendar.md`** — see "Overlap" below.

---

## V1 scope alignment

**Approved 2026-09-05 (product owner, G2) — was excluded 2026-09-04, see `v1.md §5`
amendment.** The month-of-blocks rendering itself is in scope. This view shows **time
blocks only** (planned/actual, §3/§4) — it is not, and must not become, the separate
"Calendar" screen with a generic "event" entity, which **remains excluded** (§6.4,
unchanged — see "Overlap with `calendar.md`" below, still a live concern). Rail content
stays excluded too: "Monthly Overview" stats, "Top Categories" donut, "Upcoming Events",
"Make this month count" banner (§6.4, unchanged).

---

## Layout

- Shell + Timeline toolbar; date control steps by **month** and shows it ("September
  2026"); view switcher gains **Month**.
- **Month grid:** 7 columns (Mon–Sun, ISO, D8), 5–6 week rows — same `monthGrid()` helper
  `MiniCalendar` already uses. Each **day cell**:
  - date number top-left; **today** ringed in `--brand` (matches `MiniCalendar`); days
    outside the current month muted.
  - up to **3 block chips** (time + category, category-tinted, dashed=planned/solid=actual
    per G1); overflow → "+N more".
- Right rail: mini-calendar only (unchanged from Day). **No** stats/donut/events (§6.4).

## Screen-specific components

- **Month grid** — CSS Grid `repeat(7, 1fr)` × auto rows; equal-height cells, `overflow:
  hidden` with the "+N more" overflow link, not internal per-cell scroll (keeps every row
  the same height, VP1 calm surface).
- **Day cell** — header (date, today ring, out-of-month state) + chip stack + overflow
  link. Reuses the day-cell visual language from `MiniCalendar`.
- **Block chip** — one line: time + category name, category-tinted background, dashed
  (planned) or solid (actual) border — **no titles** (a block has no title field).

## Interactions

- `‹`/`›` step by one calendar month; "Today" → the month containing today.
- Click a day cell (or its "+N more") → jumps to the Day view for that date
  (`?view=day&date=`) — Month is a summary/navigation surface, not an edit surface;
  editing a block still happens on Day/Agenda, consistent with how dense the cell is.
- "Add block" (the existing split-button) still targets the currently selected single
  date, same as Week.

## Responsive

- Below `tablet`: cells shrink but the grid stays 7 columns (chips truncate to time-only
  once too narrow for the category name) rather than falling back to another view —
  simpler than a conditional view-swap and keeps one component to maintain.
- Right rail drops first (unchanged shell behaviour).

## Overlap with `calendar.md` — ⚠ still a live flag

`timeline-month.png` and `calander.png` are the same artefact with cosmetic differences
(Calendar uses **Sunday-first** — a defect per D8 — plus an "Add Event" button and "Event
Categories"). V1 still has **one** time model (planned/actual blocks + categories) and
**no** separate "events" or "calendar" entity — this amendment approves *this* view
(Timeline Month, blocks only) and does **not** approve a separate Calendar screen or a
generic event type, which stay excluded (`design-system.md §6.4`).

## Resolved (was "cannot be inferred")

- **Planned/actual at chip size**: dashed vs. solid border, same as every other Timeline
  view (G1) — no new distinguishing rule needed, no separate legend.
- **Max chips per cell**: 3, then "+N more" (matches the reference).
- ~~Week-start~~ — **RESOLVED (D8):** Monday-first / ISO everywhere.

## Design-system references

`timeline.md` · `calendar.md` · §4.3 view switcher · §4.12 donut · §4.13 mini calendar ·
`visual-principles.md` VP7, VP9, VP10.

---

## Status: ✅ COMPLETE (2026-09-05)

`web/src/features/timeline/MonthView.tsx`. Real data — the whole visible grid (up to
42 days, including adjacent-month overflow) fetched via one `api.timelineRange(from,
to)` call (was `Promise.all` of up to 42 `api.timeline(date)` calls; the backend
range read landed 2026-09-05, see `docs/left.md`). Refetches only on month identity
change (`YYYY-MM`), not on every day-jump within the same month.

- [x] 7×6 grid, Monday-first (D8), today ringed (`MiniCalendar`'s treatment), outside-
      month days muted.
- [x] Up to 3 block chips per cell (time + category, dashed=planned/solid=actual — G1
      reused, no new legend), "+N more" overflow.
- [x] Click a day / chip / "+N more" → jump to Day view (Month is a summary/navigation
      surface; editing still happens on Day/Agenda).
- [x] **Found and fixed:** on narrow viewports, 7 columns squeezed to fit produced
      unreadable micro-cells (chip text truncated to 2–3 characters). Switched to the
      same horizontal-scroll pattern as Week (`.tl-month__scroll`) instead of shrinking
      cells further, plus an auto-scroll to today's cell on load.
- [x] No "Monthly Overview"/"Top Categories" donut/"Upcoming Events" (§6.4, unchanged);
      not the separate excluded Calendar/"event" screen (blocks only).
- [x] Responsive — verified 1440px and 390px; no page h-scroll either width, inner
      horizontal scroll reaches every column; no console errors.
- [x] Tests — 6 (fetches all 42 visible days incl. overflow, chip placement, "+N more"
      overflow, chip click → onPick, day-number click → onJumpToDay, refetch-scoped-to-
      month-identity).
- [ ] Committed — pending product owner.
