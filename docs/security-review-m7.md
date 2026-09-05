# Security review — M7 (reports)

Date: 2026-09-04 · Reviewer: Claude (manual) · Scope: new `internal/reports`
module and its five `GET /api/reports/*` routes in `cmd/server`; the one new
timeline primitive it depends on, `timeline.DailyActualTotals`.

> A pure read-only composition module — the first in the codebase that depends on
> three sibling modules' published `Service` interfaces directly rather than a
> narrow single-purpose checker (`categories.Counter`, `CategoryChecker`). The
> question this raises: does composing three services in one module widen any
> account-isolation surface? Isolation-focused review per `CLAUDE.md`, with that
> question as the specific focus.

## Verdict

**No findings.** `reports` introduces no new persistence, no new writes, and no new
trust boundary — every figure it returns is scoped by the one `accountID` its
handler takes from `reqctx.IdentityFrom`, passed unchanged into every dependency
call. Composing three services does not compose their isolation risk; each
dependency call is independently account-scoped by the callee, and `reports` never
merges data across two different account ids in one response. Covered by automated
tests, including explicit isolation cases for all 5 reports and the new
`DailyActualTotals` primitive.

## Checked and OK

| Area | Finding |
|---|---|
| Single account id, one source | `reports.Handler` reads `accountID` exactly once per request (`accountID(r)`, `reqctx.IdentityFrom`) and passes that same value to every `Service` method call within the request. There is no code path where a report could be assembled from two different account ids, and no parameter anywhere accepts an account id from the request body/query/path. |
| Composition does not defeat callee scoping | `reports.Service` calls `timeline.Service.ComparisonRange`/`DailyActualTotals`, `habits.Service.CompletionCountsInRange`, and `tasks.Service.DoneCountInRange` — each of those is independently `WHERE account_id = $1` at the SQL layer (verified in M6's and this milestone's own review). `reports` adds no query of its own; it cannot broaden what those calls already return. |
| `DailyActualTotals` (new) | `WHERE account_id = $1` via the existing `blocksOverlapping` helper (unchanged from `Comparison`/`ComparisonRange`); per-day totals only sum blocks already scoped to the caller. `*ValidationError` on `to` before `from`, guarding before any query runs. Isolation covered by `TestDailyActualTotals_Isolation`. |
| Range validation | All 5 report methods validate `to` not before `from` themselves (`reports.ValidationError`, not a re-exported error from a dependency) before calling any dependency — a malformed or inverted range never reaches a query. All 5 HTTP routes reject a missing or malformed `from`/`to` with 400 before calling the service. |
| `TaskThroughput`'s zone dependency | Converts the date range to instants via the same `AccountZone`/`accountZone` adapter every other date-taking module already uses (`account.Service.Read` → account-scoped) — no new trust boundary. |
| Read-only surface | `Handler.Mount` takes only a `read` protector (auth only) — there is no `write` path to protect with CSRF because `internal/reports` performs no writes (`v1.md §13`: "The user can view"). Confirmed by grep: no `INSERT`/`UPDATE`/`DELETE`, no sqlc block, no connection pool held by the module at all. |
| Error exposure | `writeServiceError` in `reports` maps `*ValidationError` → 400; anything else falls through to `httpx.WriteError`'s default `500 INTERNAL`, detail logged server-side only (ADR-0002) — unchanged pattern. |
| New surface | No new secrets, external calls, dependencies, or database objects. `reports` is the first module with no sqlc block and no migration — by design (a pure composition layer per its own M7 decision table). |

## Re-verification

`go build` · `go vet` · `golangci-lint run` (0 issues) · `sqlc diff` (clean — no
schema change introduced, confirming `reports` owns no table) · `gofmt -l` (clean)
· `go test ./...` (all packages, including 11 new tests in `internal/reports` and 5
new `DailyActualTotals` tests in `internal/timeline`) — all green. CP 1–3 also
walked live against a real server with curl, using an `America/New_York` account
and a real block spanning the March 2025 DST transition: all five reports over the
full month returned figures consistent with 2 real hours elapsed (not 3), including
the transition day itself in the daily-actual breakdown.
