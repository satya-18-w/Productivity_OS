# Screen — Goals

**Reference:** `docs/design/references/goals.png` (also panel 4 of `overall.png`)
**Purpose:** Define goals; set a manual progress state; view the list.
**Proposed route:** `/goals`

---

## V1 scope alignment

Maps to `requirements` §10 (goals).

| Reference element | V1 status |
|---|---|
| Goal rows: title, description, target date, kebab | in scope. A goal has title, optional description, optional target date, and a manual progress state. |
| **Progress bar + %** (70%, 40%, 58% …) | **not V1.** §10: "Progress is one of **four states**, set manually — **no percentage**, no numeric target, no progress derived from tasks, habits, or time." |
| **"12 / 20 tasks"** sub-count | **not V1.** §10: "A goal is not linked to any other entity in V1." |
| Status chip "On Track / At Risk / Completed / Not Started" | ⚠ **label mismatch.** V1 states are exactly: **not started · in progress · achieved · abandoned.** "On Track" / "At Risk" imply a derived health signal that V1 does not compute. Use the four V1 labels. |
| Category chips on goals ("Health", "Personal") + category tabs + "Goal Categories" widget | ⚠ **goals carry no category in V1** (core concepts). Out of scope. |
| KPI row: Total Goals, On Track, At Risk, Completed | "Total" and a per-state count are fine with V1 labels ("In Progress", "Achieved"); "At Risk" is not a V1 state |
| Rail: "Goal Progress" donut, "Upcoming Milestones", "Goal Categories" | donut by **state** is fine; **milestones are not V1** (§10: "No milestones, key results, or check-in history") |
| "＋ New Goal" | in scope |
| Header illustration + quote cards | decorative (VP3) |

**Recommendation:** build a list of goal rows: title · description · target date · a
**progress-state control** with the four V1 states · kebab (edit / delete). Optionally a
KPI strip of counts per state and a donut-by-state in the rail. **Drop:** %, progress
bars, task counts, categories, milestones, "On Track/At Risk".

---

## Layout

- Shell (§4.1). Header: eyebrow "GOALS" + H1 "Goals" + subtitle; header illustration.
- **View switcher** (§4.3): keep only "All Goals" + filter by **state** (Not Started /
  In Progress / Achieved / Abandoned) — **not** by category.
- Right: **primary** "＋ New Goal".
- **KPI row** (§4.6, optional): Total Goals; In Progress; Achieved; (4th: Not Started or
  Abandoned). Tinted neutrally or by `--goal` violet.
- **Goal list:** stacked wide **goal rows** (card-like, hairline-separated or individually
  carded), each with a category-hue-free left accent (use `--goal` or state colour):
  - glyph tile · title (Body-strong) · description (one–two lines) · target date chip
    (calendar icon) · **progress-state chip** (§4.9 status chip, mapped to V1 labels) ·
    kebab.
- Right rail: "Goal Progress" donut by state (§4.12), a quote card.

## Screen-specific components

- **Goal row** — reuse `web/src/styles.css` `.goal-card` / `.goal-head` / `.goal-meta` /
  `.progress-chip` which **already encode the four V1 states**
  (`progress-NOT_STARTED / IN_PROGRESS / ACHIEVED / ABANDONED`). Prefer these over the
  reference's "On Track/At Risk". Remove the `%` and task-count elements.
- **Progress-state control** — the existing `.goal-progress-select` (a `<select>` of the
  four states) or a chip menu. Any state → any state.
- **New-goal form** (§4.16) — fields: **title** (required), **description** (optional),
  **target date** (optional).

## Interactions

- Change progress state inline (select / menu). Title or kebab → edit. Kebab → delete.
- State filter narrows the list.

## Responsive

- Rows reflow: meta chips wrap below the title. KPI row wraps. Right rail drops first.

## Cannot be inferred / ambiguous

- Whether to keep any KPI strip at all (adds value but not required).
- Ordering of goals (by target date? creation? — unspecified in V1).
- Icon/glyph per goal — the reference assigns one; V1 has no goal icon field. Use a
  generic goal glyph or none.

## Design-system references

§4.1 shell · §4.2 header · §4.3 view switcher · §4.5 buttons · §4.6 KPI card ·
§4.9 status chip · §4.12 donut · §4.16 create/edit form · existing `.goal-*` /
`.progress-*` classes in `web/src/styles.css` · `requirements` §10 ·
`visual-principles.md` VP3, VP7, VP10.
