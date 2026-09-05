# Security review — MX-TL (Task / Time Block linkage)

Date: 2026-09-05 · Reviewer: Claude (manual) · Scope: `time_blocks.task_id`, the new
`timeline.TaskChecker` cross-module dependency (`timeline` → `tasks`, a new
direction), and the category-inheritance read paths (`blocksOverlapping`,
`CountByCategory`) that resolve a task-linked block's effective category.

> This adds one new cross-module dependency direction to the codebase
> (`timeline` → `tasks`, alongside the existing `timeline` → `categories`), and
> changes two already-shipped, already-tested read paths (`Comparison`/
> `ComparisonRange`/`DailyActualTotals` via `blocksOverlapping`, and
> `CountByCategory`) to resolve an inherited category instead of a block's own. The
> isolation question this raises: does resolving a `task_id` across the module
> boundary, or changing how an existing read path buckets data, ever let one
> account see or affect another's data? That is this review's focus.

## Verdict

**No findings.** Every new query is scoped by the one `accountID` the calling
handler took from `reqctx.IdentityFrom`; `task_id` validation re-derives and
re-checks that same id independently rather than trusting a caller-supplied value,
identical in shape to `CategoryChecker`/`GoalChecker` (both already reviewed with no
findings, `docs/security-review-mx1.md`, `-mx3.md`). The category-inheritance
resolution reads only the caller's own tasks (`CategoriesForTasks(ctx, accountID)`)
and can never merge in another account's category data, because the map it returns
is itself already scoped.

## Checked and OK

| Area | Finding |
|---|---|
| `tasks.AssignableToAccount` | `CountAssignableTask` is `WHERE account_id = $1 AND id = $2` — a `task_id` belonging to another account, or a nonexistent one, returns `0` and is rejected as `*ValidationError{"task_id": "task not found"}` before any write. Same call site pattern (`assertAssignableTask`, modeled on `assertAssignableCategory`) already proven safe twice. `TestAddBlock_TaskLink` covers: own task accepted; another account's task rejected; a random nonexistent uuid rejected. |
| `AddBlock`/`EditBlock` mutual-exclusion validation | `validateCategoryTaskExclusive` runs before either cross-module check, so a request with both fields set is rejected without ever touching the database — no path where a contradictory state could be written even transiently. `TestAddBlock_TaskAndCategoryMutuallyExclusive` and `TestEditBlock_TaskLink` cover create and update. |
| DB-level enforcement | `CHECK (task_id IS NULL OR category_id IS NULL)` is a second, independent guarantee below the application layer — even a future code path that forgets the app-level check cannot write a contradictory row. Defense in depth, matching the existing `ends_at > starts_at` precedent. |
| `tasks.CategoriesForTasks` scoping | `SELECT id, category_id FROM tasks WHERE account_id = $1` — every row returned is already the caller's own; `blocksOverlapping`/`CountByCategory` look up a block's `TaskID` in this map, so a task belonging to another account could never appear even if a `task_id` from another account somehow reached this code (which it cannot, per the validation above). Two independent layers would have to fail simultaneously for a cross-account leak, and neither can. |
| `blocksOverlapping` inheritance resolution | Only calls `CategoriesForTasks` when at least one row in the window actually has a `task_id` set (`taskCategoriesIfNeeded`) — a query-count optimization, not a correctness-relevant branch; both branches produce identical, correctly-scoped results. `TestComparison_TaskLinkedBlockInheritsCategory` and `TestTimeline_TaskLinkedBlockShowsInheritedCategoryName` confirm the resolved value is correct and that an uncategorized task correctly yields a nil (Uncategorized) category, not an error. |
| `CountByCategory` attribution fix | `CountBlocksByTask` is `WHERE account_id = $1 AND task_id IS NOT NULL` — same account-scoping shape as `CountBlocksByCategory`. `TestCountByCategory_TaskLinkedBlockCountsTowardInheritedCategory` proves a task-linked block is neither double-counted nor silently dropped. |
| `ON DELETE SET NULL` (schema-level) | `time_blocks.task_id uuid REFERENCES tasks(id) ON DELETE SET NULL` — Postgres enforces this at the FK level; deleting task `T` can only null out rows where `task_id = T`, which by construction were already scoped to whichever account created them. `TestDeleteTask_ClearsBlockLink` confirms the block survives, unchanged except for the cleared `task_id`, reverting cleanly to standalone/Uncategorized. |
| Wiring order (`cmd/server/main.go`) | `goalsSvc`/`tasksSvc` are now constructed before `timelineSvc` (so `timeline.NewService` can take `tasksSvc` as a `TaskChecker`) — no import cycle, since the dependency is a local structural interface (`timeline.TaskChecker`), not a direct import of `tasks`'s concrete types (ADR-0009 pattern, now applied a fourth time). |
| Error exposure | `writeServiceError` in `timeline` maps the new `task_id` `*ValidationError` the same way it already maps `category_id` — 400 with the field name, no internal detail leaked. No new error class introduced. |
| New surface | One migration (`000014_timeblock_task_link`, additive: one nullable column, one CHECK constraint, one index — no data migration, no destructive change). No new secrets, external calls, or trust boundaries. |

## Re-verification

`go build` · `go vet` · `golangci-lint run` (0 issues) · `sqlc diff` (clean) ·
`gofmt -l` (clean) · `go test ./...` — all 18 packages green, including 8 new tests
in `internal/timeline` (`TestAddBlock_TaskLink`,
`TestAddBlock_TaskAndCategoryMutuallyExclusive`, `TestEditBlock_TaskLink`,
`TestDeleteTask_ClearsBlockLink`, `TestComparison_TaskLinkedBlockInheritsCategory`,
`TestTimeline_TaskLinkedBlockShowsInheritedCategoryName`,
`TestCountByCategory_TaskLinkedBlockCountsTowardInheritedCategory`) and 2 new tests
in `internal/tasks` (`TestAssignableToAccount`, `TestCategoriesForTasks`). CP 2/3
walked live against a running server with curl: a category-carrying task linked to
an actual block with no category of its own; setting both fields on one request
rejected with `400`; `GET /api/timeline` and `GET /api/comparison` both correctly
attributed the block's time to the task's category (`Deep Work`, `5400` actual
seconds); `GET /api/categories/overview` counted the task-linked block toward that
category's block total; the task then deleted, the block confirmed to survive with
`task_id: null` and `category_id: null` (reverted cleanly to standalone/
Uncategorized).

---

## Addendum 2026-09-05 — `GET /api/tasks/{id}/blocks` (reverse lookup)

**Scope:** a new read-only endpoint, `timeline.Service.BlocksForTask`, requested by
the frontend after it needed to show a task's scheduled blocks (the reverse of the
task→category display already reviewed above). Required reopening a `v1.md §7` line
written during the original MX-TL pass that had explicitly excluded this endpoint —
done via a plain requirements amendment, not a silent code addition (the line now
explains the reversal and why).

**Verdict: no findings.** `BlocksForTask` calls `tasks.AssignableToAccount` first,
with the caller's own `accountID` — a `taskID` belonging to another account or that
doesn't exist returns `ErrTaskNotFound` (404) before any block query runs, identical
in shape to every other cross-module ownership check in this codebase. The
subsequent `ListBlocksByTask` query is itself `WHERE account_id = $1 AND task_id =
$2` — even if the ownership check were somehow bypassed, the query could not return
another account's rows. The category-inheritance resolution reuses
`CategoriesForTasks`, already reviewed above, with no new logic.

Re-verified: `go build/vet/test` (18 packages green, including `TestBlocksForTask`,
`TestBlocksForTask_EmptyAndNotFound`, `TestBlocksForTaskEndpoint`), `golangci-lint`
(0 issues), `sqlc diff` (clean — no schema change, `time_blocks_account_task_idx`
from the original MX-TL migration already supports this query), `gofmt -l` (clean).
CP walked live via curl: two blocks (one planned, one actual, different dates)
linked to a category-carrying task; `GET /api/tasks/{id}/blocks` returned both,
correctly ordered, both showing the task's inherited category; an unknown task id
correctly rejected with `404`.
