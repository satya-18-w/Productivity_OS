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

---

## Phase 4 — Tasks (list) — Status: ✅ COMPLETE (2026-09-04)

Route `/tasks` → `TasksScreen` (`web/src/features/tasks/`). Backend: `api.board()`
flattened to a task list; `api.moveTask` / `createTask` / `updateTask` / `deleteTask`.

- [x] `PageHeader` (eyebrow "Tasks" + title + subtitle + **Add task** primary).
- [x] **KPI row** — 4 `StatCard`s: Total (+ "N completed"), In progress, Overdue, Due
      this week (**ISO week** Mon–Sun, D8). No deltas / sparklines (§13 exclusion).
- [x] **Filter tabs** (`SegmentedControl`): All / Today / Upcoming / Overdue / Completed —
      **no "Starred"**. `?filter=` URL param.
- [x] **Grouped list** — `ListGroupHeader` (coloured accent bar + count): Overdue (danger),
      Today (success), Upcoming, No due date, Completed (success). Empty groups hidden.
      Within a group: dated groups sorted by due date; others newest-first (Q5).
- [x] **`TaskRow`** — `Checkbox` (done ⇄ TODO), title (→ edit), state chip, due chip
      (`--danger` + "· overdue" when overdue; "Today" when due today), **kebab `Menu`**
      (Edit · Move to <other 3 states> · Delete).
- [x] **`TaskDialog`** — `Dialog` + primitives, fields **title / description / due date
      only**. No priority / category / assignee / status field (§7 exclusions — asserted
      absent in tests).
- [x] Rail: compact "By status" card (Backlog / To do / In progress / Done counts).
- [x] Empty + error (retry) + loading states.
- [x] **New shared primitive `Menu`** (`components/ui/Menu.tsx`) — WAI-ARIA menu-button:
      click / Enter / Space / ↓ opens, arrow keys move, Enter selects, Esc / outside
      closes, focus returns to trigger. Added to `design-system.md §4.9a`.
- [x] Responsive — KPI row 4 → 2 → 1; filter tabs scroll; rail stacks below; no page
      h-scroll. Light + dark verified.
- [x] Tests — `taskGroups` (5, pure), `Menu` (5), `TaskRow` (6), `TaskDialog` (5),
      `TasksScreen` (6). Full suite green.
- [x] Browser-verified — Chromium, stubbed board; groups / KPIs / filter / kebab / add
      dialog; matches `references/tasks.png` **minus** priority/category/assignee/star and
      the priority-breakdown/category-list rail widgets (all excluded).
- [ ] Committed — pending product owner.

### Deferred
- Rail **Task Stats donut** (kept a plain "By status" count list — chart choice waits for
  the Reports spec / R1; P3 — the numbers are the point).
- "Sort by" control (only "due date" is meaningful in V1; it's the default within groups).
- Bulk select / bulk actions (reference-only).

### Bug fixed 2026-09-04 (found during a Timeline reference-accuracy audit, not a Tasks-specific pass)
The row `Checkbox` was unclickable by mouse/touch — its decorative visual box painted on
top of the real `<input>` and absorbed every click. See `design-system.md §4.10`.

### Follow-up — Category on Task (2026-09-05)

`v1.md §2` (ADR-0009) already lets a task carry a category; the backend has fully
supported `category_id` on tasks since that amendment (`internal/tasks`), but nothing in
the frontend consumed it until now. Wired up end-to-end, real data, both screens that
render a `Task` (Tasks and Board share the entity):

- [x] `api.ts` — `Task`/`NewTask` gain `category_id: string | null`. The API returns
      `category_id` only, no `category_name` (unlike a time block) — resolved client-side
      via `categoryNameFor()` (`taskGroups.ts`) against `api.listCategories()`.
- [x] `TaskDialog` — a Category `Select` (same "— none —" + list pattern as
      `BlockDialog`'s), shared by both create and edit.
- [x] `TaskRow` (Tasks list) and `TaskCard` (Board) — a `Chip` with the deterministic
      `categoryColor()` dot + resolved name, next to the existing state/due-date meta,
      shown only when a category is set.
- [x] `TasksScreen` and `BoardScreen` now fetch `api.listCategories()` alongside the
      board, once per screen, passed down to the dialog and every row/card.
- [x] **Still correctly absent** (per `v1.md §7`'s scope boundary + this task's explicit
      instruction): priority, assignees, tags, milestones. Note: `v1.md §7` was
      separately amended 2026-09-05 to add **priority** to tasks (backend already
      supports it, `internal/tasks` `Priority` field) — not built here, out of scope for
      this pass; flagged for a future, deliberately-scoped pass.
- [x] Tests — new coverage in `TaskDialog`/`TaskRow`/`TaskCard`/`TasksScreen`/
      `taskGroups.test.ts` (category select, resolved chip, chip absent when unset).
- [ ] Committed — pending product owner.
