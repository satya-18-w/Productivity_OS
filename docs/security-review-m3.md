# Security review — M3 (tasks and the Kanban board)

Date: 2026-09-04 · Reviewer: Claude (manual) · Scope: `internal/tasks`, migration
`000004`, the task routes in `cmd/server`.

> M3 adds no authentication or authorization *mechanism* — two account-owned tables
> (`tasks`, `task_transitions`) behind the existing M1 auth. Isolation-focused review per
> `CLAUDE.md`.

## Verdict

**No findings.** Every read and write of the new tables is account-scoped, all writes
sit behind auth + CSRF, and the transition log is written only inside the same
account-scoped transaction as the move it records. Covered by automated tests.

## Checked and OK

| Area | Finding |
|---|---|
| Acting account | Every handler takes the account id only from `reqctx.IdentityFrom`. No endpoint accepts an account selector; task ids come from the path but every query also filters `account_id`. |
| `tasks` queries | `CreateTask` inserts `account_id` from context; `GetTaskState` / `UpdateTaskFields` / `UpdateTaskState` / `DeleteTask` / `ListTasks` all carry `WHERE account_id = $1`. Edit/move/delete of a missing-or-foreign id returns `ErrTaskNotFound` (→ 404) and touches no row. Covered by `TestTaskIsolation`. |
| `task_transitions` writes | `RecordTransition` is only ever called inside `CreateTask` / `MoveTask`, in the same transaction, with the `account_id` taken from context and the `task_id` that was just verified to belong to that account (`GetTaskState` in the same tx). There is no endpoint that writes a transition directly. |
| `task_transitions` reads | None in M3. The `(account_id, to_state, at)` index is in place for M6/M7 range queries, which will filter by `account_id`. |
| Transaction integrity | `CreateTask` (insert + creation transition) and `MoveTask` (state check + update + transition) each run in one `pool.Begin` / `Commit` with `defer Rollback`. A same-state move commits an empty transaction (no rows), which is harmless. |
| CSRF / method | `POST` / `PATCH` / `PUT` / `DELETE` on tasks are mounted through the `write` protector (auth + double-submit CSRF). `GET /api/board` uses `read` (auth only). |
| Input validation | Title trimmed, 1–200 chars; description ≤ 5000; `due_date` parsed by `timezone.ParseDate` (rejects non-`YYYY-MM-DD`); `state` checked against the four constants (`validState`) before any DB call, and again by the table `CHECK`. Unknown JSON fields rejected (`DecodeJSON`). |
| Error exposure | `writeServiceError` maps `*ValidationError` → 400 and `ErrTaskNotFound` → 404; anything else falls through to the shared writer as a generic `500 INTERNAL`, detail logged server-side (ADR-0002). |
| FK / cascade | `task_transitions.task_id → tasks.id ON DELETE CASCADE` and both `account_id` FKs cascade from `accounts`. Deleting a task removes its transitions (verified in `TestDeleteTask`); deleting an account removes both. No orphan rows. |
| Resource use | `Board` is one `ListTasks` query per request, grouped in memory — O(tasks per account). At the N2 scale (a personal task list) no pagination is needed. |
| New surface | No new secrets, external calls, or dependencies. |

## Re-verification

`go build` · `go vet` · `golangci-lint run` (0 issues) · `sqlc diff` (clean) ·
`go test ./...` (9 packages) · `pnpm typecheck` · `pnpm build` — all green.
