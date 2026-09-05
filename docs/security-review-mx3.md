# Security review — MX3 (habit target descriptor; goal-task linkage + derived progress)

Date: 2026-09-04 · Reviewer: Claude (manual) · Scope: `habits.target` (Phase 1);
`tasks.goal_id`, `goals.AssignableToAccount`, `goals.ProgressReader` /
`tasks.ProgressByGoal`, and the new `PATCH /api/habits/{id}` and goal-linkage
surface on `POST`/`PATCH /api/tasks` (Phase 2).

> MX3 adds one new cross-module dependency direction (`tasks` → `goals`, mirroring
> the existing `tasks`/`habits`/`goals` → `categories` pattern) and one new
> HTTP-layer composition (`goals.Handler` merging in a `tasks.Service`-backed
> `ProgressReader`, mirroring `reports`' composition but scoped to a single field
> pair rather than a whole module). The isolation question this raises: does
> resolving a `goal_id` across module boundaries, or merging per-goal task counts
> into a list response, ever let one account see or affect another's data? That is
> this review's focus.

## Verdict

**No findings.** Every new query is scoped by the one `accountID` taken from
`reqctx.IdentityFrom` in the handler that owns the request; `goal_id` validation
and `ProgressByGoal`/`AssignableToAccount` both re-derive and re-check that same id
independently at each hop rather than trusting a caller-supplied value. `ON DELETE
SET NULL` at the DB layer is scoped by the foreign key itself (a task's `goal_id`
can only ever point at a goal id that existed, and clearing it never touches rows
outside the deleted goal's own linked set). Covered by automated isolation tests on
both the `habits` and `tasks`/`goals` sides.

## Checked and OK

| Area | Finding |
|---|---|
| `habits.target` (Phase 1) | Plain nullable `text` column, never used in a `WHERE` clause, never interpreted — display-only per `v1.md §9`. `UpdateHabit` scopes its `UPDATE` by `account_id = $1 AND id = $2` (`UpdateHabitFields`, `:one` — zero rows updated maps to `ErrHabitNotFound`, not silently returning another account's row). `validateTarget` bounds length (100 chars) before it ever reaches SQL. Covered by `TestUpdateHabit`'s isolation case (cross-account edit attempt returns `ErrHabitNotFound`). |
| `tasks.GoalChecker` / `goals.AssignableToAccount` (Phase 2) | `CountAssignableGoal` is `WHERE account_id = $1 AND id = $2` — a `goal_id` belonging to another account, or a nonexistent one, returns `0` and is rejected as `*ValidationError{"goal_id": "goal not found"}` before any write. This is the same shape and same call site (`assertAssignableGoal`, modeled on `assertAssignableCategory`) already proven safe by `CategoryChecker`. |
| `CreateTask` / `UpdateTask` goal validation | Both validate `GoalID` via `assertAssignableGoal(ctx, accountID, ...)` using the caller's own `accountID` — never a value from the request body — before the insert/update executes. `TestTaskGoalLink` exercises: own goal accepted; another account's goal rejected; a random nonexistent uuid rejected; clearing on update; changing on update; a foreign goal rejected on update too. |
| `ON DELETE SET NULL` (schema-level) | `tasks.goal_id uuid REFERENCES goals(id) ON DELETE SET NULL` — Postgres enforces this at the FK level; a delete of goal `G` can only null out rows where `goal_id = G`, which by construction were already scoped to whichever account created them (a task's own `account_id` is never touched by this cascade). `TestTaskGoalDeleteClearsLink` and `TestGoalEndpoints_TaskProgress` both confirm the linked tasks survive, unchanged except for the cleared `goal_id`, and that no other account's data is touched (single-account fixtures, but the FK's `ON DELETE` clause has no account-scoping mechanism to bypass in the first place — it operates purely on `goal_id` equality). |
| `tasks.ProgressByGoal` → `goals.ProgressReader` composition | `ProgressByGoal(ctx, accountID)` is `WHERE account_id = $1 AND goal_id IS NOT NULL GROUP BY goal_id` — every row it can return is already scoped to the caller. `goals.Handler.list` calls it with the same `accountID` it used for `ListGoals`, from the same single `reqctx.IdentityFrom` read at the top of the handler (mirroring `reports.Handler`'s single-account-id-per-request pattern, `docs/security-review-m7.md`). A goal from account A can never be annotated with account B's task counts, because the map lookup (`done[g.ID]`) only ever contains entries `ProgressByGoal` returned for the *same* `accountID` used to list the goals. |
| Wiring order (`cmd/server/main.go`) | `goalsSvc` is constructed before `tasksSvc` (so `tasks.NewService` can take it as a `GoalChecker`); `goals.NewHandler` is constructed after `tasksSvc` (so it can take it as a `ProgressReader`) — no import cycle, since both directions are local structural interfaces (`tasks.GoalChecker`, `goals.ProgressReader`), not direct package imports of each other's concrete types (ADR-0009 pattern). |
| Error exposure | `writeServiceError` in `tasks` maps the new `goal_id` `*ValidationError` the same way it already maps `category_id` — 400 with the field name, no internal detail leaked. `goals.Handler.list` maps a `ProgressByGoal` error through the existing generic `httpx.WriteError` path (500, detail server-side only) — unchanged pattern, no new error class introduced. |
| New surface | One migration (`000011_task_goal_link`, additive: one nullable column + one index, no data migration, no destructive change). No new secrets, external calls, or trust boundaries. |

## Re-verification

`go build` · `go vet` · `golangci-lint run` (0 issues) · `sqlc diff` (clean) ·
`gofmt -l` (clean) · `go test ./...` — all 17 packages green, including
`TestCreateHabit_WithTarget`, `TestUpdateHabit`, `TestHabitEndpoints_TargetAndEdit`
(Phase 1) and `TestTaskGoalLink`, `TestTaskGoalDeleteClearsLink`,
`TestGoalEndpoints_TaskProgress` (Phase 2). CP 1 and CP 2 both walked live against a
running server with curl: a habit created and edited with a target descriptor
(streak/completion state unaffected); 3 tasks linked to a goal, 2 marked `DONE`,
`GET /api/goals` read back `done_tasks: 2, total_tasks: 3`; a foreign/unknown
`goal_id` rejected with `400 VALIDATION_ERROR`; the goal then deleted, all 3 tasks
confirmed to survive on the board with `goal_id: null`.

---

## Addendum 2026-09-05 — MX3-follow-up (`tasks.priority`)

**Scope:** `tasks.priority` (migration `000012_task_priority`), plus a
completeness fix noticed while touching this code: `internal/export`'s own
`Task`/`Habit` response shapes were missing `goal_id`/`priority` and `target`
respectively (both added by MX3 proper and MX3-follow-up but never wired into the
M8 export bundle).

**Verdict: no findings.** `priority` is a plain nullable label with a `CHECK`
constraint at the DB layer (`'HIGH'|'MEDIUM'|'LOW'`) and app-level validation
(`validPriority`) mirroring `tasks.State`'s existing pattern exactly — it is never
used in a `WHERE` clause, never affects query results, and carries no isolation
surface beyond the task row it already belongs to (already covered by every
existing task-isolation test). The export fix is a pure read-composition change:
`export.TasksReader`/`HabitsReader` were already account-scoped via `tasks.Board`/
`habits.ListAll`; adding two/one more field(s) to the response struct doesn't touch
the query or the account-scoping at all.

Re-verified: `go build/vet/test` (18 packages green, including `TestTaskPriority`
and `TestTaskEndpoints_Priority`), `golangci-lint` (0 issues), `sqlc diff` (clean),
`gofmt -l` (clean). CP walked live via curl: created a task with `priority: HIGH`,
edited to `LOW`, cleared it (confirmed `null` via `GET /api/board`), and confirmed
an invalid value (`URGENT`) is rejected with `400 VALIDATION_ERROR`.

---

## Addendum 2026-09-05 — export completeness fix (goal derived progress)

**Scope:** `internal/export`'s `Goal` shape was missing `done_tasks`/`total_tasks`
(flagged as a known gap in `docs/export-format.md` since the MX3-follow-up pass, not
fixed then). Closed now: `export.TasksReader` gained `ProgressByGoal` (the exact
method signature `goals.Handler`'s live list endpoint already uses), `export.Export`
gained `GoalDoneTasks`/`GoalTotalTasks` maps, and `export/http.go`'s `goalBody`
gained the two fields, merged in `toExportBody` the same way `goals.Handler.list`
already merges them.

**Verdict: no findings.** `ProgressByGoal(ctx, accountID)` is already account-scoped
(reviewed with no findings under the original MX3 review above) — export calls it
with the exact same `accountID` it uses for every other reader in the same request,
so a goal from one account can never be annotated with another account's task
counts. No new query, no new table, no new cross-module dependency — `tasks.Service`
was already wired into `export.NewService` as a `TasksReader`; this only widens the
interface it must satisfy, which it already did structurally.

Re-verified: `go build/vet/test` (18 packages green, including new
`TestExport_GoalProgress` and an added assertion in
`TestExport_RoundTripCompleteness` confirming a goal with no linked tasks exports
`0`/`0`), `golangci-lint` (0 issues), `sqlc diff` (clean — no schema change),
`gofmt -l` (clean). CP walked live via curl: linked 2 tasks to a goal, marked 1
`DONE`, confirmed `GET /api/goals` and `GET /api/export` returned byte-for-byte
identical `done_tasks: 1, total_tasks: 2` for the same goal.
