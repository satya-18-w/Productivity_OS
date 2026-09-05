# Security review — M6 (daily/weekly reviews + range endpoints)

Date: 2026-09-04 · Reviewer: Claude (manual) · Scope: new `internal/reviews`
module, migration `000009`, and the M6 Phase 1 range-read additions to
`internal/timeline` (`ComparisonRange`), `internal/habits`
(`CompletionCountsInRange`), and `internal/tasks` (`DoneCountInRange` + its new
`AccountZone` dependency).

> Two additions in one milestone: a new account-owned table pair
> (`daily_reviews`/`weekly_reviews`) with free-text JSON answers, and three new
> read endpoints that widen an existing single-date query into a caller-chosen
> range. Isolation-focused review per `CLAUDE.md`, with attention to the two
> failure classes a range parameter and a free-text JSON column each invite: an
> unbounded or attacker-influenced range, and unsanitized text reaching a query or
> a response verbatim.

## Verdict

**No findings.** Every review and range query is scoped by `account_id` taken only
from `reqctx.IdentityFrom`; free-text answers are stored and returned as opaque
JSON, never interpolated into SQL or evaluated; every reachable integer conversion
is bounds-checked first. Covered by automated tests, including isolation cases for
both reviews and all three range methods.

## Checked and OK

| Area | Finding |
|---|---|
| Review isolation | `GetDailyReview`/`UpsertDailyReview`/`GetWeeklyReview`/`UpsertWeeklyReview` all filter `WHERE account_id = $1`; the account id comes from the handler's `reqctx.IdentityFrom`, never from the query string or body. A caller with no saved review for a date/week gets a blank answer set, not another account's row (`TestDailyReview_Isolation`, `TestWeeklyReview_Isolation`, `TestReviewEndpoints_Isolation`). |
| Free-text answers | `answers` is stored as `jsonb`, round-tripped through `encoding/json` — never concatenated into a query string or rendered as HTML/template on the server. `filterKnown` drops any key outside the fixed `DailyPrompts`/`WeeklyPrompts` set before it reaches the database, so the stored JSON always has a bounded, known key set; only the *values* are free text, and those are opaque strings to both Postgres and Go. No script/markup interpretation happens server-side (`v1.md §11`: "free text only"). |
| Prompt set integrity | `DailyPrompts`/`WeeklyPrompts` are Go constants, not persisted or client-suppliable — a request cannot add, remove, or reorder prompts; the response `prompts` array is always the server's fixed list, decoupled from whatever `answers` a client submits. |
| ISO week bounds | `validateYearWeek` checks `iso_year ∈ [1, 9999]` and `iso_week ∈ [1, 53]` *before* any `int → int32` conversion (the DB column type); an out-of-range value is a 400 `VALIDATION_ERROR`, never reaches the query or wraps around a 32-bit boundary. The table's own `CHECK (iso_week BETWEEN 1 AND 53)` is a second, independent guard. |
| `ComparisonRange` / `CompletionCountsInRange` / `DoneCountInRange` isolation | Each takes `accountID` from its handler's `reqctx.IdentityFrom` and every underlying query is `WHERE account_id = $1`; a second account's blocks/habits/tasks never appear in the first account's range figures (`TestComparisonRange_Isolation`, `TestCompletionCountsInRange` isolation case, `TestDoneCountInRange` isolation case). |
| Range validation | All three range endpoints reject `to` before `from` with a 400 (`*ValidationError{"to": ...}`), and a malformed `from`/`to`/`date`/`year`/`week` is rejected before any query runs. None of the three range queries is unbounded in practice: `ComparisonRange`/`CompletionCountsInRange` scan an indexed, account-scoped table (`time_blocks`/`habit_completions`, both `(account_id, ...)`-indexed) and return at most the account's own rows; `DoneTaskCountInRange` uses the existing `task_transitions_account_state_at_idx` index. A caller can request an arbitrarily wide date range, but the result set is still bounded by what that one account owns (N2 scale) — no cross-account fan-out, no unbounded join. |
| `tasks.AccountZone` (new) | Structurally identical to `timeline.AccountZone`/`habits.AccountZone`; wired to the same `accountZone` adapter in `cmd/server`, which itself resolves via `account.Service.Read` (already account-scoped) — no new trust boundary. |
| CSRF / method | Review writes (`PUT /api/reviews/daily`, `PUT /api/reviews/weekly`) go through `write` (auth + CSRF); every new range `GET` goes through `read` (auth only), matching every other read endpoint in the codebase. |
| Error exposure | `writeServiceError` in `reviews` maps `*ValidationError` → 400; anything else falls through to `httpx.WriteError`'s default `500 INTERNAL` with detail logged server-side only (ADR-0002) — unchanged pattern. |
| New surface | No new secrets, external calls, or dependencies. `sqlc.yaml` gained one block (`reviewsdb`) using the same overrides as every other module; `answers jsonb` maps to `[]byte`, handled entirely by `encoding/json` in the service, never scanned as a Go struct sqlc could get wrong. |

## Re-verification

`go build` · `go vet` · `golangci-lint run` (0 issues, after fixing two `gosec`
G115 int→int32 findings on the now-bounds-checked ISO year/week conversions and one
`staticcheck` S1016 simplification) · `sqlc diff` (clean) · `gofmt -l` (clean) ·
`go test ./...` (all packages, including 15 new tests in `internal/reviews` and the
new range/isolation cases in `timeline`/`habits`/`tasks`) — all green. CP 1–3 also
walked live against a real server with curl (range comparison across a 3-day
window, habit range counts, task throughput, and a full daily+weekly review
save/reload/edit round-trip).
