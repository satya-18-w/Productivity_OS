# Security review — M8 (data export)

Date: 2026-09-04 · Reviewer: Claude (manual) · Scope: new `internal/export`
module and its `GET /api/export` route; the five new "list everything" read
methods it depends on (`categories.ListAll`, `timeline.ListAllBlocks`,
`habits.ListAll`/`AllCompletions`, `reviews.ListDaily`/`ListWeekly`).

> An export endpoint is, by construction, the single highest-value target for a
> cross-account data leak in the whole application: one request, one response,
> every entity type the account owns. This review's entire focus is that one
> question, checked independently for each of the six sources `export` composes
> and for the composition itself.

## Verdict

**No findings.** Every one of the six read calls `export.Service.Export` makes is
independently scoped by the single `accountID` the handler takes from
`reqctx.IdentityFrom`; none of the five new "list everything" methods drops the
`account_id` filter that every other query in the codebase already carries. Two
isolation tests — one at the service layer, one through the live HTTP endpoint —
confirm a second account's export is entirely empty after the first account has
data in every entity type. Also verified live: a real download for a seeded
account, followed by an export for a brand-new account returning all-empty arrays.

## Checked and OK

| Area | Finding |
|---|---|
| Single account id, six calls, one source | `export.Handler.export` reads `accountID` exactly once (`accountID(r)`) and passes that same value into all six `Service.Export` calls within `service.go`. There is no parameter anywhere — query, path, or body — that could substitute a different account id; `GET /api/export` takes no parameters at all. |
| The five new "list everything" methods | Each is `WHERE account_id = $1` with no additional filter that could be bypassed, and none accepts an account id as an argument distinct from the one baked into the query at the call site: `ListAllCategories`, `ListAllBlocks`, `ListAllHabits`, `ListAllCompletions`, `ListAllDailyReviews`, `ListAllWeeklyReviews` — all take a single `accountID uuid.UUID` argument, all filter by it, none joins across accounts. `habit_completions` is scoped directly by its own `account_id` column (not only transitively via `habit_id`), so a completion cannot surface through a mismatched habit/account pairing even in principle. |
| Reused "everything" methods (no new query) | `tasks.Board` and `goals.ListGoals` were already account-scoped (M3/M5) and unchanged here — reused as-is, not re-implemented, so no new surface to get wrong. |
| Composition does not defeat callee scoping | Same reasoning as M7's review: `export` adds no query of its own, so it cannot broaden what its six dependencies already return. Verified explicitly this time with both a unit-level isolation test (`TestExport_Isolation`) and an HTTP-level one (`TestExportEndpoint_Isolation`) — the second account's export is `[]` for every array, not just a spot-checked subset. |
| Read-only surface | `Handler.Mount` takes only a `read` protector — no write path exists to protect with CSRF, matching `internal/reports`. Confirmed by grep: `internal/export` contains no `INSERT`/`UPDATE`/`DELETE`, no sqlc block, no connection pool. |
| Response headers | `Content-Disposition: attachment` is set from a server-computed filename (`productivity-os-export-<UTC date>.json`) — no user input reaches the header, so there is no header-injection surface. |
| Excluded entity: `task_transitions` | Deliberately not exported (documented in `docs/export-format.md`) — it is an internal audit trail, not a `v1.md §14`-named entity. Excluding it is a scope decision, not a leak risk either way (it is already `account_id`-scoped like everything else). |
| Error exposure | Any error from a dependency call reaches `httpx.WriteError` directly (export has no `*ValidationError` of its own — the endpoint takes no parameters to validate); unchanged 500-with-server-side-logging pattern (ADR-0002). |
| New surface | No new secrets, external calls, or dependencies. `sqlc.yaml` unchanged — the five new queries live in modules whose sqlc block already existed; `internal/export` itself has no sqlc block, matching `internal/reports`. |

## Re-verification

`go build` · `go vet` · `golangci-lint run` (0 issues) · `sqlc diff` (clean) ·
`gofmt -l` (clean) · `go test ./...` (all packages, including 4 new tests in
`internal/export` and 4 new "list everything" tests across
`categories`/`timeline`/`habits`/`reviews`) — all green. CP 1–3 also walked live
against a real server with curl: seeded one of every entity (including an archived
habit and its completion, and both review kinds) for one account, downloaded
`/api/export` and confirmed every entity present with the correct
`Content-Disposition` header, then confirmed a freshly registered second account's
export returns every array empty.
