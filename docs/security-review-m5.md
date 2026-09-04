# Security review — M5 (goals)

Date: 2026-09-04 · Reviewer: Claude (manual) · Scope: `internal/goals`, migration
`000006`, the goal routes in `cmd/server`.

> One account-owned table (`goals`) behind the existing M1 auth. Isolation-focused
> review per `CLAUDE.md`.

## Verdict

**No findings.** Every read and write is account-scoped; all writes sit behind auth +
CSRF. Covered by automated tests.

## Checked and OK

| Area | Finding |
|---|---|
| Acting account | Every handler takes the account id only from `reqctx.IdentityFrom`. The goal id comes from the path but every query also filters by `account_id`. |
| Queries | `CreateGoal` inserts `account_id` from context; `UpdateGoalFields` / `UpdateGoalProgress` / `DeleteGoal` / `ListGoals` all carry `WHERE account_id = $1`. Update/progress/delete of a missing-or-foreign id returns `ErrGoalNotFound` (→ 404). Covered by `TestGoalIsolation`. |
| Validation | Title trimmed, 1–200 chars; description ≤ 5000; `target_date` parsed by `timezone.ParseDate` (→ 400 on bad); `progress` checked against the four constants in the service **and** by the table `CHECK`. Path id parsed as UUID (→ 404). Unknown JSON fields rejected (`DecodeJSON`). |
| CSRF / method | `POST` / `PATCH` / `PUT` / `DELETE` on goals go through the `write` protector (auth + double-submit CSRF); `GET /api/goals` uses `read` (auth only). |
| Error exposure | `writeServiceError` maps `*ValidationError` → 400 and `ErrGoalNotFound` → 404; anything else is a generic `500 INTERNAL`, detail logged server-side (ADR-0002). |
| FK / cascade | `goals.account_id → accounts.id ON DELETE CASCADE`. A goal is linked to nothing else (`v1.md §10`), so there are no other FKs to reason about. |
| Resource use | `ListGoals` is one query, ordered, unbounded — a personal goal list is a handful of rows (N2 scale). |
| New surface | No new secrets, external calls, or dependencies. |

## Re-verification

`go build` · `go vet` · `golangci-lint run` (0 issues) · `sqlc diff` (clean) ·
`go test ./...` (11 packages) · `pnpm typecheck` · `pnpm build` — all green.
