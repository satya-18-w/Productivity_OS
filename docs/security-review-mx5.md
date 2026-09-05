# Security review — MX5 (Timeline range-read endpoint)

Date: 2026-09-05 · Reviewer: Claude (manual) · Scope: new `GET /api/timeline/range`
endpoint and `timeline.Service.TimelineRange`. No schema change, no new cross-module
dependency, no new write path.

> The narrowest possible surface added so far in this expansion: a single new
> read-only method that batches an already-reviewed one (`Timeline`) over a bounded
> date range, reusing the same `blocksOverlapping` query path already covered by
> `Comparison`/`ComparisonRange`/`DailyActualTotals`'s own reviews. The isolation
> question this raises: does batching multiple days into one response ever let a
> result belonging to one date range, or one account, leak into another? That is
> this review's focus.

## Verdict

**No findings.** `TimelineRange` takes the same single `accountID` its caller
resolved from `reqctx.IdentityFrom`, passes it unchanged into the one
`blocksOverlapping` call and the one `NamesForAccount` call it makes, and every
`DayTimeline` in its output is built by re-filtering that same account-scoped block
set per day in Go — there is no per-day database round-trip, and therefore no
per-day place for a scoping mistake to hide. This is a strictly smaller surface than
`ComparisonRange`'s already-reviewed identical shape (`docs/security-review-m7.md`).

## Checked and OK

| Area | Finding |
|---|---|
| Single account id, one source | `getTimelineRange` reads `accountID(r)` once and passes it into the one `TimelineRange` call for the request — no parameter anywhere accepts an account id from the query string or body. |
| Batching does not widen scope | `TimelineRange` calls `s.blocksOverlapping(ctx, accountID, start, end)` exactly once for the whole `[from, to]` window — the same account-scoped query every other read path in this module already uses. The per-day split afterward is pure Go slicing of an already-correctly-scoped result set; it cannot introduce another account's data because none is ever fetched. |
| Range bound enforced before any query | `daysInRange(from, to) > maxRangeDays` (62) is checked immediately after the `to`-before-`from` check, before `s.zone.Zone` or any database call — an oversized range is rejected cheaply, not after doing the expensive part of the work. |
| Category inheritance reused correctly | `TimelineRange` resolves `CategoryName` via the same `NamesForAccount` map already used by `Timeline`/`Comparison`, and inherits `blocksOverlapping`'s MX-TL category-inheritance resolution (a task-linked block's category) for free — no separate, possibly-inconsistent resolution logic was written for the range path. |
| Day-boundary correctness | Each day's bucket uses `timezone.DayWindow(d, loc)` and the same overlap test (`StartsAt.Before(dayEnd) && EndsAt.After(dayStart)`) `Timeline`/`Comparison` already rely on — `TestTimelineRange_MidnightBlockAppearsOnBothDays` confirms a midnight-spanning block appears correctly on both days, matching `Timeline`'s already-proven behavior. `TestTimelineRange_MatchesPerDayTimeline` directly cross-checks one day of range output against a standalone `Timeline` call for byte-for-byte equality. |
| Error exposure | `writeServiceError` (already shared with `EditBlock`/`DeleteBlock`) maps the new `*ValidationError` cases (`to` before `from`, range too large) to 400 with the field name — no new error class, no internal detail leaked. |
| New surface | No migration, no new table, no new cross-module dependency, no new write path. Purely a read composition over data every existing endpoint in this module already exposes. |

## Re-verification

`go build` · `go vet` · `golangci-lint run` (0 issues) · `sqlc diff` (clean — no
schema change) · `gofmt -l` (clean) · `go test ./...` — all 18 packages green,
including 6 new tests (`TestTimelineRange_MatchesPerDayTimeline`,
`TestTimelineRange_MidnightBlockAppearsOnBothDays`, `TestTimelineRange_ToBeforeFrom`,
`TestTimelineRange_ExceedsMaxDays`, `TestTimelineRange_Isolation`,
`TestTimelineRangeEndpoint`). CP 1 walked live against a running server with curl:
blocks created across a week, `GET /api/timeline/range?from=2026-08-31&to=2026-09-06`
returned exactly 7 days with blocks correctly placed and the exact JSON shape
`docs/left.md` specified; an inverted range and a >62-day range both correctly
rejected with `400`.
