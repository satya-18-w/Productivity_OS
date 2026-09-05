# Security review — M4 (habits and streaks)

Date: 2026-09-04 · Reviewer: Claude (manual) · Scope: `internal/habits`, migration
`000005`, the habit routes in `cmd/server`.

> Two account-owned tables (`habits`, `habit_completions`) behind the existing M1 auth.
> Isolation-focused review per `CLAUDE.md`.

## Verdict

**No findings.** Every read and write of the new tables is account-scoped, all writes
sit behind auth + CSRF, and a completion can only be created against a habit the caller
owns. Covered by automated tests.

## Checked and OK

| Area | Finding |
|---|---|
| Acting account | Every handler takes the account id only from `reqctx.IdentityFrom`. Habit and completion identifiers come from the path but every query also filters by `account_id`. |
| `habits` queries | `CreateHabit` inserts `account_id` from context; `SetHabitArchived` / `HabitBelongsToAccount` / `ListActiveHabits` / `ListArchivedHabits` all carry `WHERE account_id = $1`. Archive/unarchive of a missing-or-foreign id returns `ErrHabitNotFound` (→ 404). Covered by `TestHabitIsolation`. |
| `habit_completions` writes | `MarkComplete` / `UnmarkComplete` first call `assertOwned` (`HabitBelongsToAccount(account_id, habit_id)` → 0 ⇒ `ErrHabitNotFound`), then write with the `account_id` from context and the verified `habit_id`. There is no path that writes a completion without that ownership check (`TestHabitIsolation` asserts B cannot mark/unmark A's habit). |
| `habit_completions` reads | `ListCompletionDatesSince` is keyed by `habit_id`, and is only ever called for habit ids returned by `ListActiveHabits(account_id)`. The 400-day `since` bound also caps the row count returned. |
| Streak / date math | `currentStreak` and `countInRange` are pure functions over a date set; "today" comes from `AccountZone.Zone` → `timezone.Today`, i.e. the account's own timezone (ADR-0005). A future completion (Q9) is stored but ignored by `currentStreak` (`TestFutureCompletionDoesNotInflateStreak`). |
| Archive semantics (Q11) | Archiving only sets `archived_at`; `habit_completions` rows are never deleted. Unarchiving clears `archived_at` and the streak recomputes from the untouched history (`TestArchiveUnarchivePreservesHistory`). |
| CSRF / method | `POST` on habits and archive/unarchive, and `PUT`/`DELETE` on completions, are mounted through the `write` protector (auth + double-submit CSRF). `GET /api/habits` uses `read` (auth only). |
| Input validation | Habit name trimmed, 1–100 chars; the `{date}` path segment and `?date=` query are parsed by `timezone.ParseDate` (rejects non-`YYYY-MM-DD` → 400); path ids parsed as UUID (→ 404). Unknown JSON fields rejected (`DecodeJSON`). |
| Idempotency | `MarkComplete` uses `INSERT … ON CONFLICT (habit_id, on_date) DO NOTHING`; a double mark is a no-op, verified against the row count. The `UNIQUE(habit_id, on_date)` constraint is the backstop. |
| Error exposure | `writeServiceError` maps `*ValidationError` → 400 and `ErrHabitNotFound` → 404; anything else is a generic `500 INTERNAL`, detail logged server-side (ADR-0002). |
| FK / cascade | `habit_completions.habit_id → habits.id ON DELETE CASCADE`; both `account_id` FKs cascade from `accounts`. No orphan rows. |
| Resource use | `GET /api/habits` is `1 + N` queries (list + one bounded completion query per active habit). At the N2 scale (a handful of habits per account) this is fine. |
| New surface | No new secrets, external calls, or dependencies. |

## Re-verification

`go build` · `go vet` · `golangci-lint run` (0 issues) · `sqlc diff` (clean) ·
`go test ./...` (10 packages) · `pnpm typecheck` · `pnpm build` — all green.

---

## Addendum 2026-09-05 — `GET /api/habits/week` (batched "This Week" grid)

**Scope:** `habits.Service.Week` and the new `GET /api/habits/week?date=` route — a
pure optimization requested in `docs/left.md` Phase 6 (frontend was firing 7
parallel `GET /api/habits?date=` calls; same class of fix as the MX5 timeline range
endpoint). No new schema, no new `v1.md` scope — it returns exactly what `ListActive`
already computes per habit, batched over one ISO week instead of one call per day.

**Verdict: no findings.** `Week` reuses `ListActiveHabits`/`ListArchivedHabits`/
`ListCompletionDatesSince` — the same three account-scoped queries `ListActive`
already uses, called with the same `accountID` throughout. No new query shape, no
new table, no new cross-module dependency. The streak computation reuses the
existing pure `currentStreak` function unchanged; the only new logic is filtering
the same completion set down to the requested week's 7 dates for the `Completed`
field, which cannot broaden what account-scoped data was already fetched.

Re-verified: `go build/vet/test` (18 packages green, including `TestWeek`,
`TestWeek_StreakLookbackExceedsWeekWindow`, `TestHabitEndpoint_Week`),
`golangci-lint` (0 issues), `sqlc diff` (clean — no schema change, no new queries),
`gofmt -l` (clean). CP walked live via curl: a habit marked complete today,
`GET /api/habits/week?date=<today>` returned the correct Monday-first `week_start`,
all 7 `days`, the habit's `current_streak: 1` and `completed` containing today's
date, matching `docs/left.md`'s expected shape field-for-field; a missing `date`
parameter correctly rejected with `400`.
