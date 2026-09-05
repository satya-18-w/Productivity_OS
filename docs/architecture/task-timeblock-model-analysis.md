# Task ↔ Time Block model analysis

Date: 2026-09-05 · Author: Claude (analysis only — no code changed by this document)
Scope: `internal/tasks`, `internal/timeline` (owns `time_blocks`), `internal/categories`
(read-only, for the inheritance rule). No other domain read or touched.

> Requested by the product owner: Tasks and Time Blocks currently feel disconnected.
> This document determines exactly how the current implementation differs from a
> proposed Task ↔ TimeBlock linkage model, whether that model is compatible with the
> **authoritative** V1 requirements as written today, and — per explicit instruction —
> stops short of implementation if it is not, naming precisely what would need to
> change first.

## 0. Bottom line

**The proposed model is not compatible with `docs/requirements/v1.md` as currently
written, and requires a requirements amendment before it can be built.** This is not
a defect in the current implementation — it is new product scope (a Task ↔ TimeBlock
relationship does not exist anywhere in the current requirements, ADRs, or database).
Per the explicit instruction governing this analysis, **implementation does not
proceed** until the exact sections identified in §15 are amended and approved, the
same gate every prior expansion in this codebase (MX1, MX3, MX3-follow-up, MX4) went
through. §16 lists exactly what to approve to unblock this.

The rest of this document is the full technical analysis so that approval (or
targeted rejection) can be made with the complete picture, and so implementation can
start immediately once approved without re-deriving any of this.

## 1. Current domain model (as built)

### 1.1 Task (`internal/tasks`, table `tasks`)

```
tasks:
  id            uuid PK
  account_id    uuid NOT NULL FK -> accounts(id) ON DELETE CASCADE
  title         text NOT NULL
  description   text
  due_date      date                              -- plain calendar date, no time
  state         text NOT NULL CHECK (...)         -- NOTE: column is `state`, not `status`
  category_id   uuid FK -> categories(id) ON DELETE RESTRICT   -- nullable (MX1)
  goal_id       uuid FK -> goals(id) ON DELETE SET NULL         -- nullable (MX3)
  priority      text CHECK (HIGH|MEDIUM|LOW)                    -- nullable (MX3-follow-up)
  created_at    timestamptz NOT NULL DEFAULT now()
  updated_at    timestamptz NOT NULL DEFAULT now()

task_transitions:                                  -- audit trail, not a named export entity
  id, task_id, account_id, from_state, to_state, at
```

`internal/tasks/tasks.go` — `Task`, `TaskInput`, `State` (`BACKLOG|TODO|IN_PROGRESS|DONE`),
`Priority` (`HIGH|MEDIUM|LOW`), `Service` interface (`CreateTask`, `UpdateTask`,
`MoveTask`, `DeleteTask`, `Board`, `CountByCategory`, `DoneCountInRange`,
`ProgressByGoal`). No knowledge of time blocks anywhere in the module.

**Naming note (flag, not a defect):** the proposed model in the request calls the
state field `status`; the shipped schema and every reference to it in code, tests,
and the API (`{"state": "..."}`) calls it `state`, matching `v1.md §7`/`§8` exactly.
This analysis assumes the proposed model is describing the same field under a
different label, not requesting a rename — renaming a already-shipped, already
API-consumed field would be an unrelated breaking change and is explicitly out of
this analysis's scope. If a rename is actually wanted, that is a separate decision
with its own migration and frontend-contract cost.

### 1.2 Time Block (`internal/timeline`, table `time_blocks`)

```
time_blocks:
  id            uuid PK
  account_id    uuid NOT NULL FK -> accounts(id) ON DELETE CASCADE
  kind          text NOT NULL CHECK (planned|actual)   -- fixed at creation, immutable
  starts_at     timestamptz NOT NULL
  ends_at       timestamptz NOT NULL CHECK (ends_at > starts_at)
  category_id   uuid FK -> categories(id) ON DELETE RESTRICT   -- nullable
  created_at    timestamptz NOT NULL DEFAULT now()
```

`internal/timeline/timeline.go` — `Block`, `BlockInput`, `BlockKind`, `Service`
interface (`AddBlock`, `EditBlock`, `DeleteBlock`, `Timeline`, `Comparison`,
`ComparisonRange`, `DailyActualTotals`, `CountByCategory`, `ListAllBlocks`). No
`task_id` column, no knowledge of tasks anywhere in the module. `CategoryStore` is
the only cross-module dependency (`AssignableToAccount` + `NamesForAccount`, wired to
`categories.Service` — ADR-0009's established pattern).

### 1.3 Category (`internal/categories`, read-only for this analysis)

A flat label with `colour`/`icon` (both opaque, presentation-only, ADR-0009 §"Decision").
`AssignableToAccount(ctx, accountID, categoryID) (bool, error)` is the only check any
consumer module performs — it does **not** interpret colour, and neither would this
change. The request's "category colour is presentation-only and must never affect
business logic" rule is **already fully satisfied** by the existing implementation;
nothing about this proposal touches it.

### 1.4 What's proposed, restated precisely

- `time_blocks` gains a nullable `task_id` FK to `tasks(id)`.
- A block with `task_id` set inherits the task's category; it must not carry a
  contradictory `category_id` of its own.
- A block with `task_id` NULL (standalone — Sleep, Lunch, Break, Morning Routine) keeps
  its own optional `category_id`, exactly as today.
- No change to `kind` (planned/actual), no change to the four-state task board, no
  Spaces concept (explicitly excluded, and not present anywhere in this codebase).

## 2. Ownership / account isolation

No change to the isolation *mechanism* — `account_id` still comes only from
`reqctx.IdentityFrom` (ADR-0004), never the client, in both modules, unchanged. The
new surface is a **new cross-module validation**: a `task_id` a caller attaches to a
block must belong to the *same account* the block belongs to. This is exactly the
shape already proven three times in this codebase (`CategoryChecker` in
`tasks`/`habits`/`goals`, `GoalChecker` in `tasks`) — a new `TaskChecker` interface
defined in `timeline` (the consumer), satisfied structurally by `tasks.Service`,
wired in `cmd/server`. No new isolation *risk* — the pattern's isolation properties
are already covered by three existing security reviews (`docs/security-review-mx1.md`,
`-mx3.md`) that reached "no findings" on the identical shape.

## 3. Database relationships

```
tasks (1) ---- (0..n) time_blocks      via time_blocks.task_id, nullable
categories (1) ---- (0..n) tasks       via tasks.category_id, nullable   [existing]
categories (1) ---- (0..n) time_blocks via time_blocks.category_id, nullable, but only
                                        meaningful when task_id IS NULL  [existing column,
                                                                          new constraint]
```

One task can have many blocks (planned *and* actual — e.g., a "Write report" task
might have one planned block for tomorrow and, later, an actual block recording when
it really happened). One block links to at most one task. This mirrors the
already-established goal ↔ task shape exactly (MX3: "many tasks per goal, one goal
per task" — here it is "many blocks per task, one task per block").

**Category inheritance, structurally enforced.** Rather than storing a possibly-stale
copy of the task's category on the block (denormalization that could drift if the
task's category is later changed), the recommended design is:

- `time_blocks.category_id` stays nullable exactly as today.
- A new `CHECK (task_id IS NULL OR category_id IS NULL)` constraint makes "a
  task-linked block never carries its own category" a database-level invariant, not
  just an application convention — consistent with how `ends_at > starts_at` is
  already enforced at the DB layer rather than only in Go.
- A task-linked block's *effective* category for display/analytics is resolved by
  joining to `tasks.category_id` at read time. This is the only way to guarantee the
  inherited category never drifts out of sync with the task's actual category — if
  the task's category changes, every block linked to it reflects that immediately,
  with no update-cascade code needed anywhere.

This is a design recommendation, not yet approved — see §16.

## 4. Migrations

One new forward-only migration (continuing the existing sequence at `000014`):

```sql
-- 000014_timeblock_task_link.up.sql
ALTER TABLE time_blocks ADD COLUMN task_id uuid REFERENCES tasks (id) ON DELETE SET NULL;
ALTER TABLE time_blocks ADD CONSTRAINT time_blocks_category_xor_task
    CHECK (task_id IS NULL OR category_id IS NULL);
CREATE INDEX time_blocks_account_task_idx ON time_blocks (account_id, task_id);
```

```sql
-- 000014_timeblock_task_link.down.sql
DROP INDEX IF EXISTS time_blocks_account_task_idx;
ALTER TABLE time_blocks DROP CONSTRAINT time_blocks_category_xor_task;
ALTER TABLE time_blocks DROP COLUMN task_id;
```

Purely additive: every existing `time_blocks` row gets `task_id = NULL` automatically
(every block that exists today is, and remains, standalone) — zero backfill, zero data
loss, zero rows touched destructively (ADR-0003's forward-only convention, unchanged).

**`ON DELETE SET NULL` is a proposed default, not yet confirmed** — see §8 and §16.
It mirrors MX3's already-approved precedent for `tasks.goal_id` exactly ("deleting a
goal clears the link on its tasks without deleting them"), applied symmetrically here
("deleting a task clears the link on its blocks without deleting them — a block is a
record of when something happened/was planned, and that record has value independent
of whether the originating task still exists").

## 5. API changes (minimum required)

| Endpoint | Change |
|---|---|
| `POST /api/tasks`, `PATCH /api/tasks/{id}` | No change needed for the link itself — a task never carries a `block_id`; the reference lives on the block. Unaffected. |
| `GET /api/board` | Unaffected. |
| `POST /api/blocks` (create planned or actual) | `blockRequest` gains optional `task_id`. If set: validated against the caller's own tasks (`TaskChecker`); `category_id` must be absent/null in the same request (reject with `400 VALIDATION_ERROR` on both present — see §7). |
| `PUT /api/blocks/{id}` (update) | Same `task_id`/`category_id` mutual-exclusion rule. Clearing `task_id` (unlinking) is allowed and, once cleared, `category_id` may then be set on a subsequent edit. |
| `DELETE /api/blocks/{id}` | Unaffected — deleting a block never touches its task. |
| `GET /api/timeline?date=` | Response block shape gains `task_id`; a task-linked block's `category_id`/`category_name` in the response reflect the **inherited** value (resolved server-side), not a stored one. |
| `GET /api/comparison` (single date and range) | Per-category totals must bucket a task-linked block under its task's category, not "Uncategorized" — see §13 (analytics impact). This is a behavioural change to an existing, already-shipped endpoint, not just an additive one. |
| `GET /api/export` | `Block` export shape gains `task_id` (raw, for the consumer to join if needed) — matching the same completeness standard just applied to `Task`/`Habit` in the MX3-follow-up pass. |

**Not proposed** (matches "expose the minimum," no unsupported features): no new
`GET /api/tasks/{id}/blocks` listing endpoint, no embedding a task's linked blocks on
the task read model, no bidirectional state sync (moving a task to `DONE` does not
touch any block; completing an actual block does not touch the task's state). If any
of these are wanted, they are separate, later decisions.

## 6. Service / store changes

**`internal/timeline`:**
- New `TaskChecker` interface (mirrors `CategoryChecker` exactly):
  ```go
  type TaskChecker interface {
      AssignableToAccount(ctx context.Context, accountID, taskID uuid.UUID) (bool, error)
  }
  ```
  `tasks.Service` needs a new `AssignableToAccount` method to satisfy it structurally
  (mirrors `categories.Service.AssignableToAccount` and `goals.Service.AssignableToAccount`,
  both already shipped — same one-line SQL `count(*) WHERE account_id=$1 AND id=$2`).
- A bulk category-inheritance lookup, mirroring `CategoryStore.NamesForAccount`'s
  batch shape (avoids N+1 when a day's timeline has many task-linked blocks):
  ```go
  // CategoriesForTasks resolves each of the caller's tasks to its category id (nil if
  // uncategorized), for blocks that inherit a category from a linked task.
  type TaskCategories interface {
      CategoriesForTasks(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]*uuid.UUID, error)
  }
  ```
  (Could be folded into `TaskChecker` as one two-method interface, or kept separate —
  an implementation-time style choice, not a decision this analysis needs to force.)
- `AddBlock`/`EditBlock` gain `TaskID *uuid.UUID` handling: validate via `TaskChecker`,
  reject if `TaskID != nil && CategoryID != nil` (mutual exclusion, §7).
- `blocksOverlapping` (the shared internal helper behind `Timeline`, `Comparison`,
  `ComparisonRange`, `DailyActualTotals` — confirmed via `readmodel.go`) must resolve
  each task-linked block's inherited category before bucketing — this is the one
  piece of real logic complexity this change introduces, and it is shared by every
  read path, so it is written once.
- `CountByCategory` (feeds `GET /api/categories/overview`) must also resolve inherited
  categories, or a task-linked block's contribution to that category's block count
  silently disappears — a real regression if missed. See §13.

**`internal/tasks`:**
- New `AssignableToAccount(ctx, accountID, taskID uuid.UUID) (bool, error)` method on
  `Service` (to satisfy `timeline.TaskChecker`) and a `CategoriesForTasks`-shaped
  method (or the category piece could live entirely in `timeline` if it already holds
  every task's category via a bulk fetch — implementation detail).
- No schema change to `tasks` itself. No behavioural change to task CRUD, board, or
  transitions.
- `DeleteTask`: no code change required if `ON DELETE SET NULL` is used (the database
  does the work) — matches how `goals.DeleteGoal` needed no special-case code for
  MX3's identical pattern.

## 7. Validation

| Rule | Enforced where | Failure |
|---|---|---|
| `task_id`, if present, must be a valid UUID | HTTP layer (`blockRequest` parsing) | `400`, field `task_id`, "must be a UUID" |
| `task_id`, if present, must belong to the caller's account | Service layer (`TaskChecker.AssignableToAccount`) | `400 VALIDATION_ERROR`, field `task_id`, "task not found" |
| `task_id` and `category_id` must not both be set on the same request | Service layer (mirrors the existing `assertAssignableCategory`/`assertAssignableGoal` pattern) | `400 VALIDATION_ERROR`, field `category_id`, "cannot set a category on a task-linked block; it inherits the task's category" |
| Start/end validation | Unchanged — `validateBlockTimes` (`ends_at > starts_at`), already enforced at both the Go and DB layer | Unchanged |
| Kind fixed at creation | Unchanged | Unchanged |

No change to any existing validation rule; this is additive validation only.

## 8. Timezone handling

**No new timezone concern.** `time_blocks.starts_at`/`ends_at` are already
`timestamptz`, already converted from wall-clock via `AccountZone` at the HTTP layer
per ADR-0005, unaffected by adding `task_id`. `tasks.due_date` is already a plain
`timezone.Date` with no time component. A "list blocks for a task" query, if it
existed (it is not proposed — §5), would need no date-bucketing at all since it would
filter by `account_id + task_id`, not by date. Category inheritance resolution
(joining a block's task to get its category) is a plain UUID join with no
timezone-sensitive logic whatsoever. **This section of the analysis is included per
the request's checklist, and its finding is "no impact."**

## 9. Planned vs. actual semantics

`kind` remains immutable and orthogonal to `task_id`. A task may have:
- a **planned** block (I intend to work on this task 2–3pm tomorrow), and, later and
  independently,
- an **actual** block (I actually worked on it 2:15–3:30pm),

with no requirement that one exist for the other, and no automatic conversion between
them (`v1.md §4`: "No 'convert plan to actual'" — this proposal does not reopen that;
a task_id is set explicitly on whichever block(s) the user chooses to link, planned or
actual or both or neither). No change to the planned/actual distinction itself.

## 10. Category inheritance (the core new rule)

Restated precisely as the recommended, DB-enforced design (§3):

- Standalone block (`task_id IS NULL`): `category_id` behaves exactly as today —
  optional, freely settable, its own value.
- Task-linked block (`task_id IS NOT NULL`): `category_id` is always `NULL` in
  storage (enforced by the `CHECK` constraint — the caller cannot set it, and any
  request attempting to would be rejected at validation before it ever reaches SQL,
  §7); its effective category for every read path (timeline view, comparison
  reports, categories-overview counts, export) is resolved by joining to the linked
  task's `category_id`.
- If the user wants a task-linked block to show a *different* category than its
  task, the only path is to change the task's category (which then applies to every
  block linked to it) or unlink the block and set its own category directly. There is
  no per-block override — this is the direct implementation of "should normally
  inherit/use the Task's category rather than allowing contradictory category
  assignment."

## 11. Standalone time blocks

Fully preserved, zero behavioural change. Sleep, Lunch, Break, Morning Routine, or any
other block with no natural task continue to work exactly as they do today: created
with `task_id` omitted, optionally categorized directly, listed/edited/deleted with no
new constraint beyond what already exists. The new `CHECK` constraint is satisfied
trivially (`task_id IS NULL` makes the OR always true regardless of `category_id`).

## 12. Deletion / update behavior

| Action | Behavior |
|---|---|
| Delete a task that has linked blocks | **Proposed default (unconfirmed, see §16):** blocks survive, `task_id` set to `NULL` at the DB layer (`ON DELETE SET NULL`) — a block is a record of a time period, independently valuable even if its originating task is gone. Mirrors MX3's identical, already-approved `goal_id` precedent exactly. |
| Delete a block that references a task | No effect on the task whatsoever — deleting a block never touches `tasks`. Already true by construction (the FK direction is block → task, not the reverse). |
| Edit a task's category while blocks are linked to it | Every linked block's *effective* category changes immediately (it was never stored on the block — §10) — no cascade code needed, this falls out of the read-time-join design for free. |
| Edit a task's title/description/due date/priority/state while blocks are linked | No effect on any block. The link is a reference, not a data copy; nothing about a task's other fields propagates anywhere. |
| Unlink a block from its task (clear `task_id` on update) | Allowed; the block becomes standalone and may then have its own `category_id` set on a subsequent edit. |
| Re-link a standalone block to a task | Allowed; if the block currently carries its own `category_id`, the same mutual-exclusion validation applies — the request must clear `category_id` in the same call that sets `task_id`, or it is rejected (§7). |

## 13. Analytics impact

This is the one area with genuine, easy-to-miss regression risk if implemented
carelessly — flagged explicitly because it is the kind of thing that would pass a
naive test suite and silently under-count in production:

- **`Comparison`/`ComparisonRange`/`DailyActualTotals`** (`v1.md §6`, §13's "Planned
  vs actual" and "Time by category" reports) currently bucket every block by its own
  `category_id`, with `NULL` falling into the explicit "Uncategorized" bucket (Q8).
  Once task-linked blocks always store `category_id = NULL`, **every task-linked
  block would silently fall into "Uncategorized" in every report unless the read path
  is changed to resolve the inherited category first.** This is not optional
  follow-up work — it is required for the reports to remain correct at all once this
  ships, and must land in the same change, not a later one.
- **`CountByCategory`** (feeds `GET /api/categories/overview`'s per-category block
  count) has the identical problem: a task-linked block would silently stop counting
  toward its (inherited) category's block total unless the query also resolves
  through the task. Same requirement: fix in the same change.
- **`internal/reports`** (the five `v1.md §13` reports) and **`internal/export`**
  consume `timeline.Service`'s methods, not `time_blocks` directly — if the fixes
  above land inside `timeline`, `reports` and `export` inherit the correct behavior
  automatically with no code change of their own (composition, ADR-0002's dependency
  rule working as intended). This is a point in favor of the read-time-join design
  over any alternative that would require every consumer to know about the
  inheritance rule itself.

## 14. Edge cases

| Case | Resolution |
|---|---|
| A task has 5 linked blocks (2 planned, 3 actual) across different dates | All fine — no cardinality limit proposed or needed; `time_blocks_account_task_idx` supports an efficient "all blocks for this task" query if ever needed later. |
| A block's `task_id` points at a task that was deleted after the block was created, before `ON DELETE SET NULL` is confirmed as the policy | Cannot happen with `SET NULL` (the FK guarantees referential integrity always) — this row is exactly why the DB constraint, not just app logic, matters. |
| Two accounts, one owns the task, the other owns the block being created | Rejected at validation — `TaskChecker.AssignableToAccount` checks `account_id` match, identical to the existing `CategoryChecker`/`GoalChecker` cross-account rejection tested in `TestTaskGoalLink` and `TestTaskCategory`. |
| Client sends both `task_id` and `category_id` as the *same* effective value the task already has (i.e., "no real contradiction") | Still rejected — the rule is structural ("a task-linked block never stores its own category"), not "reject only if different," to keep the invariant simple and the `CHECK` constraint meaningful. Flagged as a possible UX friction point worth confirming (§16) — an alternative is to silently ignore a matching `category_id` rather than reject it, at the cost of the DB constraint becoming "ignore instead of reject," which is a materially different (and weaker) guarantee. |
| A task-linked block whose task has `category_id = NULL` (task itself is uncategorized) | The block's effective category is also `NULL` → "Uncategorized," which is correct and requires no special-casing beyond the join already described. |
| Midnight-spanning blocks with a linked task | No interaction — the task link is orthogonal to the existing midnight-spanning handling (`v1.md §3`/`§4`/`§5`, already correct); nothing about this proposal touches that logic. |
| Overlapping blocks, one linked to a task, one not, same time range | No interaction — overlap is explicitly allowed and unflagged today (`v1.md §3`/`§4`); this proposal does not change that. |

## 15. Backwards compatibility / migration concerns

- **Database:** purely additive migration (§4) — no existing row is touched, no data
  loss possible, no backfill required. Zero risk to existing data.
- **API:** additive on every endpoint except `GET /api/comparison` and
  `GET /api/categories/overview`, whose **response values** for accounts that start
  using task-linked blocks would change (a previously-"Uncategorized" bucket now
  correctly attributes to the task's category) — this is a correctness fix relative
  to the new model, not a breaking shape change (the JSON shape is unchanged, only
  which bucket a given block's seconds land in, and only for blocks the user
  deliberately links). No existing block is retroactively affected, since no existing
  block has a `task_id`.
- **Frontend contract:** every response shape gains one new field (`task_id`,
  nullable) rather than removing or renaming anything — safe for a frontend that
  hasn't been built against this yet to adopt incrementally. This is exactly why §16
  is worth resolving now, per the request's own framing ("before continuing frontend
  implementation") — once the frontend is built against a block shape *without*
  `task_id`, adding it later is still additive and safe, but building the linking UI
  itself would be rework if done twice.
- **No V1 non-goal is reopened.** Checked explicitly against `v1.md`'s non-goals list
  and `design-system.md §6.4`: this proposal does not introduce recurring tasks,
  subtasks, a live timer, Spaces, or any dashboard/analytics beyond the five fixed
  reports. It is a narrow relational addition between two entities that already exist.

## 16. Requirements gate — exact sections requiring revision

Per the instruction governing this analysis, the following are the **precise**
authoritative-document changes needed before implementation, named exactly so they
can be approved (or rejected/amended) without further research:

1. **`docs/requirements/v1.md` §3 (Day planning / planned blocks)** — currently:
   *"The user can add a planned block for a chosen date with a start time, an end
   time after the start, and an optional category."* Needs an added bullet: *"The
   user can add a planned block that references a task instead of, or in addition
   to being categorized directly"* plus a scope-boundary amendment stating the
   inheritance rule (§10) and that a task-linked block does not carry its own
   category.
2. **`docs/requirements/v1.md` §4 (Activity logging / actual blocks)** — same
   amendment, mirrored for actual blocks.
3. **`docs/requirements/v1.md` §5 (Timeline view)** — note that a task-linked
   block's displayed category is the task's category (a display-semantics
   clarification, not new behavior beyond §3/§4).
4. **`docs/requirements/v1.md` §6 (Planned vs actual comparison)** and **§13
   (Reports)** — note that per-category totals attribute a task-linked block to its
   inherited category (the analytics-impact fix, §13 of this document, needs to be
   a named, intentional behavior in the requirements, not just an implementation
   detail).
5. **`docs/requirements/v1.md` §7 (Tasks)** — add: *"A task may have zero or more
   linked time blocks (planned or actual); deleting a task does not delete its
   linked blocks"* (mirroring the existing MX3 goal-linkage bullet's style exactly)
   plus the delete-behavior default from §12/§8 of this document, flagged the same
   way MX3's goal-delete default was flagged: **"a build-time default: deleting a
   task clears the link on its blocks rather than blocking the delete; flagged for
   confirmation, not yet product-owner approved"** unless approved outright now.
6. **Domain-concepts intro (`v1.md`, top of file)** — the **task** bullet and the
   **time block** bullet both need one clause each describing the optional link, matching
   the style already used for the habit-target and goal-linkage amendments earlier
   this expansion.
7. **No new ADR is strictly required.** ADR-0009 already established the
   generalizable pattern this reuses exactly (a domain module holding a narrow
   structural interface into another module's `Service`, wired in `cmd/server`,
   with per-module isolation tests). This is an application of that decision, not a
   new architectural choice. If the product owner wants the category-inheritance
   *design* itself (§3/§10 — DB-enforced mutual exclusion + read-time join, as
   opposed to some other shape) recorded as a durable decision, a short ADR
   addendum or a note under ADR-0009 would be reasonable but is not mandatory.

## 17. What to approve to unblock implementation

A minimal, unambiguous approval would answer:

1. **Approve the core linkage:** a time block may optionally reference a task
   (yes/no).
2. **Approve the inheritance rule as designed:** a task-linked block never stores its
   own category; it always inherits the task's (yes / no — propose an alternative).
3. **Approve the delete-default:** deleting a task clears `task_id` on its linked
   blocks via `ON DELETE SET NULL`, without deleting the blocks (yes/no).
4. **Confirm no additional API surface is wanted** beyond §5's minimum (no
   `GET /api/tasks/{id}/blocks`, no bidirectional state sync) for this pass.

Once answered, `docs/requirements/v1.md` gets amended exactly as in §16 (this agent
proposing the precise wording, product owner approving it in place — the same
process already used for MX3/MX3-follow-up/MX4), and then the full implementation in
§4–§7 proceeds in one milestone, following the exact phased/checkpoint/security-review
pattern every prior milestone in `planning.md` has used.
