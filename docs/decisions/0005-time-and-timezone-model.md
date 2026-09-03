# 0005 — Time & Timezone Model

**Status:** Accepted — 2026-09-03
**Applies to:** V1 (and beyond unless superseded)

## Context

N4 is a hard correctness requirement: instants stored unambiguously; "a date" and "a
week" always resolved in the account's timezone; blocks that span midnight and reports
whose range crosses a daylight-saving transition must produce correct totals.
Time-dependent behaviour appears in M2 (timeline, planned-vs-actual), M4 (streaks), M6
(daily / weekly reviews), and M7 (reports). Several product questions about time
behaviour (Q4, Q7, Q9) are open and are **not** resolved by this ADR.

## Decision

- Each account has a **timezone** stored as an **IANA time zone name** (e.g.
  `Asia/Kolkata`).
- Every instant is stored in PostgreSQL as `timestamptz`. The application passes
  timezone-aware `time.Time` values; the storage representation is UTC.
- The **server process runs in UTC** (`TZ=UTC`); no code relies on the host's local
  time.
- **All** calendar reasoning — which date an instant falls on; the bounds of a date,
  an ISO week, or a date range in the account's timezone — is done in Go, in a single
  `internal/platform/timezone` package, using `time.LoadLocation(account.Timezone)`.
  This package is the only place timezone math lives.
- Time-bounded queries filter by an **instant range `[start, end)`** computed in Go by
  that package. Queries do not perform timezone conversion; there is no scattered
  `AT TIME ZONE` in SQL.

## Alternatives considered

- **Bucket dates / weeks in SQL with `AT TIME ZONE`.** Works, but spreads timezone
  logic across every time-bounded query, is harder to unit-test, and makes DST-crossing
  ranges fragile. Centralizing in one tested Go package is more maintainable.
- **Store a fixed UTC offset instead of an IANA name.** Rejected: offsets change with
  DST; the IANA name is the correct primitive and lets `time` compute the right offset
  for any instant.
- **Per-time-block timezone.** Rejected: not in the requirements; N4 specifies an
  account-level timezone.
- **Store local wall-clock times without a zone.** Rejected: ambiguous around DST
  transitions and midnight; violates N4.

## Consequences

- One well-tested module owns the hardest correctness surface in the product; every
  feature that needs date / week logic calls it rather than re-deriving it.
- Time-bounded list queries take a computed `[start, end)` parameter — a minor,
  consistent shape constraint on the query layer (ADR-0003).
- A DST-observing account (e.g. `America/New_York`) and a non-DST account (e.g.
  `Asia/Kolkata`) must both be covered by the test matrix (ADR-0007), including a
  report range that crosses a DST boundary.
- Changing an account's timezone changes how historical instants bucket into dates /
  weeks going forward; the M2 / M6 specs define the user-facing behaviour.

## Deferred (not resolved by this ADR)

- The default timezone at account creation — product open question **Q4**.
- Whether overlapping-block "total time" is a sum of durations or wall-clock coverage —
  **Q7**.
- Whether past-tense data may be recorded for dates that have not yet occurred — **Q9**.

These are product questions consumed by the relevant milestone specifications.

## Revisit conditions

- A requirement for per-entity (not per-account) timezones appears (V2+).
- The Go standard library's timezone handling proves insufficient for a needed case →
  evaluate a dedicated date library, still centralized in the same package.

## Related documents

- `docs/requirements/v1.md` — N4; open questions Q4, Q7, Q9
- `docs/architecture/conventions.md` — timezone package usage
- ADR-0002, ADR-0003, ADR-0007
