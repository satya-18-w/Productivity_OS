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
