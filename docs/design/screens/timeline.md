# Screen — Timeline (Day view) + shared Timeline shell

**Reference:** `docs/design/references/timeline.png` (also panel 9 of `overall.png`)
**Purpose:** View a chosen date's time blocks positioned against the hours of the day.
**Proposed route:** `/timeline` (default `?view=day&date=<today>`). Currently `/` renders
the existing day timeline.

This spec owns the **Timeline shell** shared by all four views
(`timeline-week.md`, `timeline-month.md`, `timeline-agenda.md` extend it).

---

## V1 scope alignment

Timeline (Day) maps to **`requirements` §5** and depends on §3 (planned blocks) and §4
(actual blocks).

| Reference element | V1 status |
|---|---|
| Hour axis 06:00–22:00 with "now" indicator | in scope (granularity is a frontend choice, §5) |
| Blocks with time range, category, icon | in scope |
| **Planned vs actual distinction** | **required by §5** — the reference does **not** clearly show two lanes / two visual states. Must be added. The existing `web/src/styles.css` `.tl-planned` / `.tl-actual` lanes already solve this. |
| Per-block checkbox (mark done) | **not V1** — a time block has no completion state; only tasks and habits do. Likely the mock is conflating blocks with tasks. |
| Tag chips on blocks ("Deep Work", "LeetCode") | **not V1** — blocks have only start/end/category |
| Assignee avatars on "Team Sync" | **not V1** (P4 — no collaboration) |
| Right-rail "Today's Tasks", "Focus Mode" | Tasks list plausible; Focus Mode not V1 |
| View switcher Day/**Week/Month/Agenda** | **§5 scope boundary: "one date at a time; no week or month timeline."** Week & Month views are **excluded from V1** (`design-system.md` §6.4) — do not build them. Agenda is an acceptable list rendering of the *same single day*. |

**Recommendation:** build **Day**, plus **Agenda** as a list rendering of the same day.
Keep the planned/actual model from the existing implementation. Drop checkboxes, tags,
avatars. **The V1 view switcher offers Day / Agenda only** — no Week, no Month.

---

## Layout (Day view)

- Shell: `design-system.md` §4.1.
- Header: date line + greeting/quote (reference reuses the dashboard greeting — a plain
  "Timeline — <date>" is the honest V1 version, VP3).
- **Timeline toolbar** row:
  - left: **view switcher** segmented control (§4.3) — **Day / Agenda only** for V1.
  - centre: date stepper — `‹` + date pill (with calendar icon, opens picker) + `›` +
    "Today" button.
  - right: **split primary** "＋ Add ▾" (§4.5) — menu: Add planned block / Add actual
    block.
- **Timeline body:** left hour axis (fixed ~48px), then the block area. Per the existing
  design system this is **two lanes** (Planned | Actual), blocks absolutely positioned by
  start/end against the hour grid (`.tl-*` classes). The reference's single-column look is
  acceptable *only if* planned/actual stay visually distinct (fill vs outline, or a lane
  label) — reconcile with existing CSS, do not regress §5.
- Empty lane → `.tl-empty` hint.
- Right rail: mini month calendar (§4.13, drives date), "Today's Tasks" list (V1 tasks due
  today — read-only or with state control), optional quote card.

## Screen-specific components

- **Hour axis** — `.tl-axis` / `.tl-tick` (exists).
- **Now indicator** — dark pill with current time + node dot on the axis (exists as
  concept; reference shows "10:24").
- **Time block** — `.tl-block` + `.tl-planned` / `.tl-actual` (exists). Content: time
  range, category name, category-tinted fill + left border. **No** checkbox/tags/avatars.
- **Date stepper** — `.date-nav` (exists).
- **Add-block menu** — split button → the inferred create form (§4.16) with fields
  **start, end, category** only (§3, §4).

## Interactions

- Prev/next/today/date-pick change the date (URL `?date=`).
- Click a block → edit (start / end / category); delete from the same surface (§3, §4).
- "＋ Add" → create planned or actual block for the current date; end may fall on the next
  calendar day (§3, §4).
- A block spanning midnight renders correctly at the day boundary (§5).

## Responsive

- Right rail drops first.
- Lanes: on mobile, stack Planned above Actual, or keep side-by-side in a horizontally
  scrolling container (`.tl-scroll` exists). Page never scrolls sideways.
- View switcher scrolls horizontally.

## Cannot be inferred / ambiguous

- Whether the reference intends one merged lane or two — resolved here in favour of §5
  (two, or clearly distinct).
- Vertical scale / visible hour range (06–22 shown; blocks outside, e.g. "Sleep 22:00–
  06:00", need a rule — clip, wrap, or scroll).
- Whether "Today's Tasks" in the rail is interactive.

## Design-system references

§3.6 shell grid · §4.1 shell · §4.2 header · §4.3 view switcher · §4.5 buttons ·
§4.13 mini calendar · §4.16 create/edit form · existing `.tl-*` classes in
`web/src/styles.css` · `visual-principles.md` VP3, VP5, VP10.

---

# Phase 2 — Timeline (Day) — SPEC & PLAN

> **Maps to** `v1.md §3, §4, §5, §6`. **G1 resolved** (design-system.md §6.1).
> The existing `web/src/pages/Timeline.tsx` already implements this feature end-to-end
> against the backend — Phase 2 **migrates it to the design system + app shell** and adds
> the a11y / responsive / test coverage. It is not a rewrite of the data flow.

## SPEC — what Phase 2 delivers

The `/timeline` screen (also `/`), showing a chosen date's planned and actual blocks
positioned against a 24-hour axis, with the per-category planned-vs-actual comparison.

### Kept from the existing implementation

- Backend calls: `api.timeline(date)`, `api.comparison(date)`, `api.listCategories()`,
  `api.createBlock/updateBlock/deleteBlock`. **No API changes.**
- Two lanes (Planned | Actual); **time-proportional** blocks on a full 00:00–24:00 axis (G1).
- Midnight-spanning blocks: ▲ (from prev day) / ▼ (to next day) markers.
- Block form fields: **kind** (planned/actual, locked when editing), **date**, **start**,
  **end**, **ends next day**, **category** — nothing else (§3/§4).
- §6 comparison: per-category planned / actual / difference + totals; `Uncategorized`
  bucket (Q8); `pos`/`neg` colouring on the difference.

### Changed for Phase 2

| Area | From | To |
|---|---|---|
| Container | bare `.stack` on `AuthLayout` | `<ScreenLayout rail={<MiniCalendar>}>` in the app shell |
| Header | `.date-nav` row | `<PageHeader eyebrow="TIMELINE" title="<full date>" subtitle="Plan and log your day.">` (factual, VP3) + a **date toolbar** below |
| Date toolbar | prev/next/today/date input | `‹`/`›` `IconButton`s + native date `Input` + "Today" `Button`; drives `?date=` URL param |
| Add | one `<button>Add block</button>` | primary `Button` "Add block" → block **`Dialog`** (kind toggle inside) — `SplitButton` deferred to a later polish |
| Block form | inline `.card` form | `<Dialog>` + `<Field>`/`<Input>`/`<Select>`/`<Checkbox>` primitives; Delete in dialog actions |
| Block visuals | `.tl-planned` (green) / `.tl-actual` (green) — **collide** | `ui-tl-*`: fill/border = **category colour** (`categoryColor(category_id)`, VP2). **Planned** = dashed border + 14 % fill. **Actual** = solid border + 22 % fill. Lanes stay labelled — planned/actual reads from lane + line-style, not hue (G1). |
| Comparison | `<section class="card">` + `table.totals` | `<Card title="Planned vs actual">` wrapping the (tokenised) totals table |
| "now" line | none | a horizontal marker at the current time **only when the date is today** |
| Right rail | none | `<MiniCalendar>` (Monday-first, D8) — selecting a day sets `?date=` |

### States

loading · load error (`ErrorState`) · empty lane ("Nothing planned" / "Nothing actual") ·
no comparison data · dialog: new / edit / save error / field error / delete-confirm ·
block spanning midnight · viewing a non-today date (no "now" line).

### Interactions

`‹`/`›`/Today/date-pick/mini-calendar → change date (`?date=`). "Add block" → dialog for
the current date; `end` may fall on the next calendar day. Click a block → edit dialog;
delete from it. Future/past dates allowed (Q9). ISO week in the mini-calendar (D8).

### Responsive (D4 / VP9)

Rail (mini-calendar) drops below main `< wide` (`ScreenLayout`). The two lane tracks keep
a min width and scroll **inside `.tl-scroll`** if the viewport is tight — the page never
scrolls sideways. Toolbar wraps. Dialog is full-width-minus-margin on mobile.

### Accessibility (VP8)

- Each block is a `<button>` with an accessible name: `"<category> — planned/actual,
  <start>–<end>[, continues next day]"`.
- Lanes: `<h3>`/`aria-label` "Planned" / "Actual"; axis ticks not focusable.
- Planned vs actual conveyed by **lane label + border style**, not colour alone.
- Date toolbar controls labelled; `Dialog` traps focus (native `<dialog>`); the
  ends-next-day checkbox and category select wired via `<Field>`.
- Comparison table: real `<th>` headers; difference sign not colour-only (has `−` / `+`).

### New shared component

- **`MiniCalendar`** (`web/src/components/date/`) — month grid, **Monday-first** (D8),
  today ringed in `--brand`, selected day filled, `‹`/`›` month nav. `value` (ISO date) +
  `onChange`. Added to `design-system.md §4.13`. a11y: grid of buttons, `aria-label` per
  day, selected = `aria-pressed`.

## PLAN

### Files

```
web/src/components/date/MiniCalendar.tsx (+ .test.tsx, index.ts)
web/src/features/timeline/
  TimelineDay.tsx        — screen: data load, date param, dialog state, ScreenLayout
  TimelineGrid.tsx       — axis + two lanes + blocks + now-line
  BlockDialog.tsx        — create/edit/delete block form in a <Dialog>
  ComparisonCard.tsx     — §6 table in a <Card>
  timelineFormat.ts      — fmtMinute / fmtDuration / date helpers (from the old page)
  index.ts
  *.test.tsx
web/src/styles/timeline.css   — ui-tl-* (migrated from styles.css .tl-*), tokens only
web/src/styles/index.css      — @import timeline.css
web/src/App.tsx               — import TimelineDay from features/timeline
web/src/pages/Timeline.tsx    — DELETE (logic moved)
docs/design/design-system.md  — §4.13 MiniCalendar contract
docs/architecture/conventions.md — note web/src/features/ for feature screens
```

### Order

1. `MiniCalendar` + tests.
2. `timeline.css` (`ui-tl-*`): axis, lanes, block base, `--planned`/`--actual` treatments,
   category-colour custom property, now-line. Migrate from legacy `.tl-*`.
3. `timelineFormat.ts`; `TimelineGrid` (presentational — takes blocks, renders lanes).
4. `BlockDialog` (Dialog + primitives; same payload as today).
5. `ComparisonCard`.
6. `TimelineDay` — compose; `?date=` via `useSearchParams`; `ScreenLayout` + rail.
7. Wire `App.tsx`; delete old page; leave legacy `.tl-*` in `styles.css` untouched for now
   (other code may reference — check; the Board/Habits pages don't).
8. Tests → typecheck → build → Playwright (stub `/api/*`) → screenshots → QA.

### Tests (`pnpm test`)

- `MiniCalendar`: renders the right month, Monday-first, today ringed, selecting a day
  fires `onChange`, month nav.
- `TimelineGrid`: a planned block lands in the Planned lane with `--planned` styling and
  `top`/`height` proportional to its minutes; actual block → Actual lane, `--solid`;
  midnight block shows the ▼ marker; empty lane shows the hint; now-line only for today.
- `BlockDialog`: opens with defaults; rejects `end ≤ start` (or surfaces the API field
  error); edit pre-fills; kind locked on edit; delete calls `api.deleteBlock`.
- `TimelineDay`: changing the date (stepper + mini-calendar) refetches; load error →
  `ErrorState`; renders inside the shell.
- `ComparisonCard`: totals sum correctly; `Uncategorized` row; difference sign classes.

### Playwright verification

- `/timeline` with stubbed `api.timeline` / `api.comparison` / `api.listCategories`
  fixtures (planned + actual + one midnight block).
- Screenshots: desktop, mobile, dark.
- Confirm: two labelled lanes; planned dashed vs actual solid; block positions match the
  fixture times; mini-calendar in the rail (stacks below on mobile); "now" line on today;
  no console errors; no horizontal page scroll.
- Open the block dialog; keyboard-close it.
- Compare to `references/timeline.png` for spacing / type / category colour language —
  document the deliberate deviations (two lanes not one list; no checkboxes / tags /
  avatars / greeting / focus-mode — all excluded).

### Acceptance criteria

- [ ] View a chosen date's planned + actual blocks, positioned against the hours,
      planned and actual visually distinguishable, midnight-spanning correct (§5).
- [ ] Add / edit / delete planned and actual blocks; end may be next-day (§3, §4).
- [ ] See per-category planned time, actual time, and their difference for the date (§6).
- [ ] Renders in the app shell; mini-calendar rail; responsive; light + dark.
- [ ] a11y checklist passes; keyboard-only usable.
- [ ] `pnpm typecheck && pnpm test && pnpm build` green; Playwright screenshots captured.
- [ ] No excluded feature present (checkbox/tags/avatars/greeting/focus/week/month).

### Dependencies / blockers

- None. Backend timeline + comparison + categories + block CRUD APIs exist and are used
  by the current page. `MiniCalendar` is built here. `Dialog`, `Field`, `Input`, `Select`,
  `Checkbox`, `ScreenLayout`, `PageHeader`, `categoryColor` all exist.

## Status — ✅ COMPLETE (2026-09-04)

- [x] SPEC + PLAN approved (G1 decided 2026-09-04)
- [x] Implemented — `web/src/features/timeline/**`, `web/src/components/date/**`,
      `web/src/styles/timeline.css`; old `web/src/pages/Timeline.tsx` deleted
- [x] Tests green — 23 new (MiniCalendar 4, TimelineGrid 7, BlockDialog 5,
      ComparisonCard 3, TimelineDay 5+)
- [x] Browser-verified — Chromium (stubbed API): two lanes, planned dashed / actual
      solid, category colour, "now" line on today, mini-calendar rail, block dialog,
      no console errors, no h-scroll; desktop + mobile + dark screenshots
- [x] Visual QA — matches `references/timeline.png` spacing/type/category-colour
      language; deviations documented (two lanes not one list; no checkboxes / tags /
      avatars / greeting / focus-mode — all excluded)
- [x] Responsive QA — rail stacks below main `< wide`; lane tracks scroll inside
      `.tl2__scroll` at narrow widths; page never scrolls sideways
- [x] Accepted
- [ ] Committed — pending product owner

### QA fixes applied

- Legacy bare `label { flex-direction: column }` was breaking `.ui-field-check` and
  `.ui-field__label` inside the new Dialog → added explicit `flex-direction: row` /
  `display: block` overrides in `primitives.css`. (Affects every form using the
  primitives, not just Timeline.)

### Follow-ups (not blockers)

- `SplitButton` for "Add planned / Add actual" (currently one "Add block" with an
  in-dialog kind toggle).
- Consider stacking the two lanes vertically on very narrow screens instead of
  horizontal scroll.
- Legacy `.tl-*` classes in `styles.css` are now unused by any screen — remove at a
  cleanup phase.
