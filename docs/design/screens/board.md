# Screen — Board (Kanban)

**Reference:** none (no dedicated mock; visual language taken from `references/tasks.png`
card styling + the ratified design system).
**Requirement:** `docs/requirements/v1.md §8` (+ §7 for the task model).
**Route:** `/board` (D10 — separate route from `/tasks`, same task model).

---

## V1 scope

`v1.md §8`: all tasks in **four fixed columns** `BACKLOG → TODO → IN_PROGRESS → DONE`
(that order); move a task from any column to any other → sets its state; reflects
create/edit/delete from §7.

**Scope boundary (build none of these):** no column create/rename/reorder/delete; one
board per account; no WIP limits, swimlanes, filters, or saved views. Task ordering within
a column is not a V1 requirement — columns list newest-first, no manual reorder (Q5).

Also excluded (reference-only): priority chips, category chips, assignee avatars, cover
images, card counts beyond the column total.

---

## Phase 5 — Board (Kanban) — Status: ✅ COMPLETE (2026-09-04)

`web/src/features/board/` — `BoardScreen` / `BoardColumn` / `TaskCard`. Reuses
`TaskDialog`, `Menu`, `Badge`, `taskGroups` helpers from Phase 4. Data: `api.board()`;
mutations: `api.moveTask` / `createTask` / `updateTask` / `deleteTask`.

- [x] `PageHeader` (eyebrow "Board" + title + subtitle) + **Add task** primary.
- [x] **Four fixed columns** in the required order (`STATE_ORDER`); column head =
      label + count `Badge`; `<section role="region" aria-label="<state> — N tasks">`.
- [x] **`TaskCard`** — `draggable`; title (→ edit), 2-line description clamp, due chip
      (`--danger` + "· overdue" / "Today"), **kebab `Menu`** = Edit · Move to <other 3
      states> · Delete.
- [x] **Move a task**: (a) **native HTML drag-and-drop** card → column
      (`dragstart` sets `text/task-id`; column `dragover`/`drop`; `--dragover` highlight);
      (b) **the kebab "Move to …" menu** — the keyboard- and touch-accessible path (native
      DnD is pointer-only). No-op if dropped on its current column.
- [x] `TaskDialog` (shared) — title / description / due date only.
- [x] Loading + error (retry) states. No rail (the board is wide).
- [x] Responsive — the 4 columns scroll **inside `.board2__scroll`** below ~laptop; the
      page never scrolls sideways (VP9). Light + dark verified.
- [x] a11y — columns are labelled regions; cards have accessible names; the kebab menu is
      the non-drag operable path (WAI-ARIA menu-button); DnD is a pointer enhancement only.
- [x] Tests — `TaskCard` (4), `BoardColumn` (3, incl. simulated drop), `BoardScreen` (5,
      incl. kebab move + current-column no-op). Full suite green.
- [x] Browser-verified — Chromium: 4 columns in order + counts, kebab move (hides current
      state), overdue styling, dark, mobile (columns scroll, page doesn't); no console
      errors.
- [ ] Committed — pending product owner.

Old `web/src/pages/Board.tsx` deleted.

### Deferred
- Keyboard drag (grab/move/drop with arrow keys) — the "Move to" menu already gives a
  full keyboard path; a true keyboard-DnD is a later enhancement.
- Drop-position indicator between cards (ordering isn't a V1 concern — Q5).
