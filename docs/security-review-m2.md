# Security review — M2 (categories and the timeline core)

Date: 2026-09-04 · Reviewer: Claude (manual) · Scope: `internal/timeline`,
`internal/platform/{timezone,reqctx}`, migrations `000002`–`000003`, the timeline routes
in `cmd/server`.

> M2 adds no authentication, session, or authorization *mechanism* — it adds two
> account-owned tables (`categories`, `time_blocks`) behind the existing M1 auth. This
> review is isolation-focused, per `CLAUDE.md` ("mandatory for milestones touching
> account isolation").

## Verdict

**No findings.** Every read and write of the new tables is account-scoped, cross-account
category references are rejected at the application layer, and all writes sit behind auth
+ CSRF. Isolation is covered by automated tests.

## Checked and OK

| Area | Finding |
|---|---|
| Acting account | Every timeline handler takes the account id only from `reqctx.IdentityFrom`, populated solely by the M1 auth middleware. No endpoint accepts an account selector; `category_id` / block ids come from the path or body but every query also filters by `account_id`. |
| Category queries | `CreateCategory` inserts `account_id` from context; `RenameCategory` / `ArchiveCategory` / `GetCategory` / `ListActiveCategories` / `CountAssignableCategory` all carry `WHERE account_id = $1`. Rename/archive of a missing-or-foreign id returns `ErrCategoryNotFound` (→ 404), never touches another account's row. Covered by `TestCategoryIsolation`. |
| Time-block queries | `CreateTimeBlock` inserts `account_id` from context; `UpdateTimeBlock` / `DeleteTimeBlock` / `GetTimeBlock` / `ListBlocksOverlapping` all filter by `account_id`. Edit/delete of a foreign block → `ErrBlockNotFound` (404). Covered by `TestEditBlock`, `TestDeleteBlock`, `TestComparison_Isolation`. |
| Cross-account category on a block | `assertAssignableCategory` runs `CountAssignableCategory(accountID, categoryID)` before insert/update; a category owned by another account (or archived, or unknown) yields 0 → `400 VALIDATION_ERROR`. The DB FK enforces existence only; ownership is enforced in the app layer + tested (`TestAddBlock_CategoryMustBeOwnedAndActive`), consistent with the V1 "no RLS" decision (ADR-0004). |
| CSRF / method | Category and block **writes** (`POST` / `PATCH` / `PUT` / `DELETE`) are mounted through the `write` protector (auth + double-submit CSRF). Reads (`GET /api/categories`, `/api/timeline`, `/api/comparison`) use `read` (auth only). |
| Timezone handling | The account's timezone is read via `account.Read` (scoped to the caller's own id) and resolved with `timezone.LoadLocation`. The client never supplies a zone for block math — it submits wall-clock `date` + `HH:MM` and the server converts (ADR-0005). A stored timezone was already IANA-validated at M1. |
| Input validation | Block times parsed server-side (`time.Date`, no arithmetic on attacker strings); category name trimmed and capped at 60 chars; malformed `date` / `HH:MM` / `category_id` → `400`. Unknown JSON fields rejected (`DecodeJSON`). |
| Error exposure | `writeServiceError` maps the timeline sentinels to 4xx; anything else falls through to the shared writer as a generic `500 INTERNAL` with detail logged server-side (ADR-0002). `ListBlocksOverlapping` and zone-resolution failures are wrapped and surface as generic 500s. |
| Resource use | Timeline / comparison queries are bounded to one date's blocks; the comparison aggregation is O(blocks-that-day) in memory. At the N2 scale (thousands of blocks per account, tens of accounts) this needs no pagination. |
| New surface | No new secrets, no new external calls, no new dependencies. `_ "time/tzdata"` embeds the zoneinfo DB (build-time only). |

## Re-verification

`go build` · `go vet` · `golangci-lint run` (0 issues) · `sqlc diff` (clean) ·
`go test ./...` · `pnpm typecheck` · `pnpm build` — all green.
