# Tasks + Board — Reference-vs-Actual Gap Plan (2026-09-04)

Reference: `docs/design/references/tasks.png`. Actual: `/tmp/opencode/pg-tasks.png`,
`/tmp/opencode/pg-board.png`. Specs: `screens/tasks.md` (Phase 4 ✅), `screens/board.md`
(Phase 5 ✅). V1: `requirements/v1.md` §§7–8. Fixed V1 task model for this pass:
**title / description / due-date only**, states `BACKLOG/TODO/IN_PROGRESS/DONE`, board =
4 fixed columns, newest-first.

Grep-verified absent from `features/tasks/`, `features/board/`, `styles/tasks.css`,
`styles/board.css`: priority, star, tags, assignees/avatars, categories on tasks,
recurring, due time (only hits: `weekStart`, `dragstart`, "get started", and the
`TaskDialog` test that asserts those fields are absent).

## Deviations kept (reference shows them; V1 forbids them)

| Reference element | Kept out | Justification |
|---|---|---|
| Starred tab / star affordance | ✅ out | tasks.md: no favourites/pin on tasks |
| Priority chips High/Medium/Low + Priority Breakdown donut | ✅ out | v1.md §7 scope boundary; design-system §6.4 reference-only |
| Category chips on rows + Categories rail widget | ✅ out | Per this pass's fixed scope (see note below); tasks.md: task carries no category |
| Assignee avatars | ✅ out | P4; v1.md §7 "No assignee" |
| Global search, notification bell, sidebar Spaces, Dashboard/Notes/Calendar | ✅ out | Shell-level; not this pass's scope; design-system §6.4 |
| KPI sparklines, trend deltas, motivational sublabels ("Keep going!") | ✅ out | v1.md §13 (no range comparison); D6 forbids adaptive encouragement |
| Task Stats donut → plain "By status" count list | ✅ kept as-is | Numbers are the point (P3); chart choice waits on Reports R1 |
| Split "Add Task ▾" → plain "Add task" primary | ✅ kept as-is | Quick presets deferred; create path exists |
| Sort/Filter/Group/View toolbar, bulk select | ✅ out | Sort deferred (due-date is already the within-group default); bulk is reference-only |
| Quote card in rail | ✅ out | D6: when in doubt, omit |
| Keyboard drag, drop-position indicator (board) | ✅ out | Kebab "Move to …" is the full keyboard/touch path; ordering isn't V1 (Q5) |
| State chip on task rows (not in reference) | ✅ kept | Harmless V1 addition: surfaces the 4-state model the reference hides |

## Gaps to build (Phase B)

1. **Checkbox uncheck restores the previous state, not always TODO** — tasks.md:
   "Checkbox toggles 'done' … maps to state `DONE` ⇄ previous state". Actual
   (`TasksScreen.toggleDone`) sends `TODO` on every uncheck, so unchecking a task
   that was `BACKLOG`/`IN_PROGRESS` silently re-files it as `TODO`. Fix: session-level
   map of last non-DONE state per task id (frontend-only, no `api.ts` change); fall
   back to `TODO` when unknown (e.g. task already DONE at first load). Test: check
   an `IN_PROGRESS` task → `DONE`, then uncheck → `IN_PROGRESS`.
2. **None other.** Grouping (Overdue/Today/Upcoming/No due date/Completed), KPI row,
   filter tabs + `?filter=` URL param, kebab (Edit · Move to ×3 · Delete), title→edit,
   `TaskDialog` (title/description/due-date only), loading/error/empty states,
   board DnD + kebab move + same-column no-op, responsive (KPIs 4→2→1, board inner
   scroll, no page h-scroll) all match spec in the screenshots and code.

## Backend gaps noticed (list only — not implemented)

- `api.ts` `Task`/`NewTask` carry no category: amended v1.md §2 (ADR-0009, 2026-09-04)
  now says a category "may be assigned to … a task", but neither the client types nor
  the fixed scope of this pass include it. If categories-on-tasks lands, `TaskDialog`,
  row/card chips, and the negative-assertion test need revisiting. (Working tree
  already has unrelated uncommitted backend category work; untouched.)
- Board "newest-first" (Q5) is rendered in backend order; the frontend assumes the
  `/api/board` column order satisfies it — worth a backend assertion, not a frontend fix.
