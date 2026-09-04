# Screen — Tasks

**Reference:** `docs/design/references/tasks.png` (also panel 2 of `overall.png`)
**Purpose:** Create, view, edit, and change the state of tasks.
**Proposed route:** `/tasks`

> The **Kanban board** (`requirements` §8) is a *second view of the same tasks*. The
> reference shows a **list** view. Both are legitimate; the existing app has `/board`
> only. Proposed: `/tasks` (list) with a view toggle to the board, or keep `/board`
> separate. This is a routing decision (D10). This spec covers the **list** view.

---

## V1 scope alignment

Maps to `requirements` §7 (tasks) and §8 (board).

| Reference element | V1 status |
|---|---|
| Task rows: checkbox, title, category chip, due date, kebab | in scope. A task has title, optional description, optional due date, and one of `BACKLOG/TODO/IN_PROGRESS/DONE`. |
| Checkbox toggles "done" | in scope — maps to state `DONE` ⇄ previous state |
| Grouping: Overdue / Today / Upcoming / Completed | derivable from due date + state |
| Status tabs All / Today / Upcoming / Overdue / Completed | derivable filters |
| **Starred** tab / star affordance | **not V1** (no favourites/pin on tasks) |
| **Priority** chips High/Medium/Low | **not V1** — §7 scope boundary explicitly excludes priorities |
| Assignee avatars | **not V1** (P4) |
| Rail: "Task Stats" donut (Completed/In Progress/Overdue/Not Started) | derivable from state + due date |
| Rail: "Categories" list with counts | ⚠ V1 categories attach to **time blocks only**, *not* tasks (`requirements` core concepts: "Tasks, habits, and goals carry no category in V1"). **A task has no category.** So the category chips on task rows and the category breakdown are **out of V1 scope.** |
| Rail: "Priority Breakdown" donut | not V1 |
| Sort by Due Date / Filter / Group / View controls | Sort by due date is reasonable; the rest optional |
| "Add Task ▾" split button | in scope (create task) |

**Recommendation:** build the grouped list with checkbox + title + due date + state
control + kebab. **Drop (ratified — `design-system.md` §6.4):** priority, star, assignee,
and **task categories**. Extending categories to tasks would be a `requirements` change
(register item C1), not a design decision.

---

## Layout

- Shell (§4.1). Header: eyebrow "TASKS" + H1 "Tasks" + subtitle; title icon badge; header
  illustration.
- **View switcher** (§4.3): All / Today / Upcoming / Overdue / Completed. (No "Starred".)
- Right: **split primary** "＋ Add Task ▾" (§4.5).
- **KPI row** (§4.6) — 4 stat cards: Total Tasks (with "N completed"), In Progress,
  Overdue, Due This Week. Tinted green / blue / red / violet.
- **Toolbar row:** select-all checkbox, "Sort by: Due Date" dropdown; (optional Filter /
  Group / View on the right — defer).
- **Grouped list:** section per group with a **group header** (§4.8) — coloured left
  accent bar + label + count (Overdue red, Today green, Upcoming neutral, Completed
  green). Rows within: checkbox, title, due date (calendar icon; red if overdue;
  title struck-through + muted when completed), kebab.
- Right rail: "Task Stats" donut (§4.12), a quote card. (No Categories / Priority
  widgets.)

## Screen-specific components

- **Task row** — list row (§4.8): checkbox (§4.10) · title (Body-strong) · due chip
  (icon + date, `--danger` when overdue) · kebab. State beyond done/not-done is set via
  the kebab menu or the row's edit surface (Backlog / To Do / In Progress / Done, any
  direction — §7).
- **Group header** — §4.8 pattern.
- **Add-task form** (§4.16) — fields: **title** (required), **description** (optional),
  **due date** (optional, plain date). Nothing else.

## Interactions

- Checkbox → set `DONE` (and back to the prior state). Kebab → change state / edit /
  delete. Title click → edit. Tab / sort change the query (URL params).
- "＋ Add Task" → create form; optional dropdown for quick presets (defer).

## Responsive

- Right rail drops first; KPI row wraps 4→2→1; rows reflow (due chip wraps under title);
  view switcher scrolls.

## Cannot be inferred / ambiguous

- Whether the board and this list share a route with a toggle (D10).
- "Due This Week" definition (ISO week vs next 7 days).
- Whether completed tasks stay in their date group or all move to "Completed".
- Ordering within a group (V1: unspecified; reference sorts by due date).

## Design-system references

§4.1 shell · §4.2 header · §4.3 view switcher · §4.5 buttons · §4.6 KPI card ·
§4.8 list row + group header · §4.10 checkbox · §4.12 donut · §4.16 create/edit form ·
`requirements` §7–§8 · `visual-principles.md` VP1, VP7, VP10 · see also `board` (Kanban).
