# 0009 — Categories as a shared, cross-cutting module

**Status:** Accepted — 2026-09-04
**Applies to:** V1.x onward (design-driven expansion — see `planning.md` Appendix A)

## Context

V1 shipped categories as an internal concern of `internal/timeline`: a flat label a
**time block** may reference, with no colour or icon carrying product meaning
(`v1.md §2`). The approved design (`docs/design/references/`) makes categories a
top-level organising concept — "Spaces" in the navigation — that groups **tasks,
habits, goals, notes, and calendar events** as well as time blocks, each category
carrying a colour and an icon, with per-category item counts.

That requires:
1. tasks / habits / goals / notes / events to carry a `category_id`;
2. categories to have `colour` and `icon`;
3. a single owner for the `categories` table that every other domain module can depend
   on without importing its internals (ADR-0002 boundary rule).

The `v1.md §2` scope boundary and the domain-concepts note ("Tasks, habits, and goals
carry no category in V1") are amended alongside this ADR.

## Options considered

### Option A — Keep categories inside `timeline`; other modules import `timeline`
- Pros: no move.
- Cons: `tasks`, `habits`, `goals`, `notes` would each depend on the whole `timeline`
  module for one small concern — the exact coupling ADR-0002 forbids. `timeline` becomes
  a god-module.

### Option B — Duplicate a lightweight category concept per module
- Pros: perfect isolation.
- Cons: five `categories` tables, no shared identity, no "one Work category across the
  app", contradicts the design.

### Option C — Extract `internal/categories` as its own module
- Pros: one owner, one table, one published interface; every other module depends only
  on a small `categories` interface; matches the design's mental model.
- Cons: a real refactor of `timeline`; a cross-module FK (`time_blocks.category_id` →
  `categories.id`) now crosses a module boundary at the DB level.

## Decision

**Option C.** `internal/categories` becomes its own domain module.

- It **owns the `categories` table** and every query against it. Schema still lives in
  the single `db/migrations/` directory (ADR-0003); the category `queries.sql` moves to
  `internal/categories/`.
- Its published interface (`categories.Service`, one interface at the package root per
  ADR-0002) covers: create, rename, set colour/icon, archive, list; plus
  `AssignableToAccount(ctx, accountID, categoryID)` for other modules to validate a
  supplied `category_id`.
- `categories` gains `colour` and `icon` columns (both nullable text — an icon/colour
  *key*, not a hex value the backend interprets).
- `tasks`, `habits`, `goals`, and (later) `notes`, `events` each gain a **nullable
  `category_id`** FK to `categories(id)` with `ON DELETE RESTRICT` (categories are never
  hard-deleted — `v1.md §2` "no hard delete" — so RESTRICT only guards the theoretical
  case, matching `time_blocks`).
- A module that needs to validate a `category_id` holds a small
  `CategoryChecker` interface (just `AssignableToAccount`), wired to `categories.Service`
  in `cmd/server`. No module imports another's internals.
- **Cross-table item counts** (a category's task/habit/goal/block totals) are *not*
  computed inside `categories` — that would require it to know every other table's
  schema. Instead each domain module implements a `categories.Counter` interface
  (`CountByCategory(ctx, accountID) (map[uuid.UUID]int, error)`), and a thin composition
  handler in `cmd/server` assembles `GET /api/categories/overview`. `categories` stays
  pure (its own table only).
- Cross-account isolation rules (ADR-0004) apply unchanged: every category query is
  scoped by `account_id`; `AssignableToAccount` returns false for a category the caller
  does not own.

## Consequences

- Positive: one shared `Work` category across tasks, habits, goals, notes, blocks;
  `timeline` shrinks to just time blocks; the boundary rule holds — modules depend on a
  4-method interface, not each other.
- Negative / accepted: a one-time refactor of `internal/timeline` and its tests; a DB
  FK now crosses a module boundary (accepted — it is one database, and the *code* still
  only crosses via the interface); the counts endpoint has a small amount of wiring
  ceremony in `cmd/server`.
- Follow-up: `internal/notes` and `internal/calendar` (Appendix A) will add their own
  `category_id` + `Counter` implementation when built. `time_blocks` category handling
  moves from an in-package query to the `CategoryChecker` call.

## Links

- Requirements: `docs/requirements/v1.md` §2 (amended with this ADR)
- Principles: `docs/product/principles.md` E2 (module boundaries), P5 (user owns data)
- Related ADRs: 0002 (module architecture), 0003 (persistence), 0004 (isolation)
- Plan: `planning.md` Appendix A · milestone **MX1**
