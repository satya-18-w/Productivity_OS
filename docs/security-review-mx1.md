# Security review — MX1 (categories as a shared module)

Date: 2026-09-04 · Reviewer: Claude (manual) · Scope: new `internal/categories`
module, its extraction from `internal/timeline`, the new `category_id` column and
validation path on `tasks` / `habits` / `goals`, migrations `000007`–`000008`, and
the `GET /api/categories/overview` composition handler in `cmd/server`.

> A module-boundary refactor (ADR-0009) plus four new cross-module FK relationships.
> Isolation-focused review per `CLAUDE.md`, with extra attention to the two new
> failure classes this milestone introduces: a category id validated by one module
> but leaking data from another account, and a composition endpoint that reads five
> services under one request.

## Verdict

**No findings.** Every new query is account-scoped; the validation path a supplied
`category_id` goes through cannot succeed for a category the caller does not own;
the composition endpoint takes the account id from the same source as every other
handler. Covered by automated tests, including per-module isolation cases added for
this milestone.

## Checked and OK

| Area | Finding |
|---|---|
| `categories` table ownership | Only `internal/categories` queries the `categories` table now. `internal/timeline`'s former `CreateCategory`/`RenameCategory`/`ArchiveCategory`/`ListActiveCategories`/`CountAssignableCategory` and the `LEFT JOIN categories` in `ListBlocksOverlapping` are gone; it reaches categories only through the `CategoryStore` interface (`AssignableToAccount`, `NamesForAccount`), both of which take `accountID` from the caller. |
| Category CRUD isolation | `Update`/`Archive` filter `WHERE account_id = $1 AND id = $2`; a foreign or unknown id returns `ErrNotFound` (→ 404), never leaks existence. Covered by `TestIsolation` in `internal/categories`. |
| `AssignableToAccount` (the shared checker) | `SELECT count(*) FROM categories WHERE account_id = $1 AND id = $2 AND archived_at IS NULL` — true only for an active category the caller owns. Every module that calls it (`timeline`, `tasks`, `habits`, `goals`) does so with the `accountID` from its own handler's `reqctx.IdentityFrom`, never from the request body. A foreign, archived, or unknown category id on task/habit/goal/block create-or-update returns `*ValidationError{"category_id": ...}` → 400, not a silent cross-account link. Covered by `TestTaskCategory`, `TestHabitCategory`, `TestGoalCategory`, `TestAddBlock_CategoryMustBeOwnedAndActive`, and the categories package's own `TestAssignableToAccount` matrix (owned-active / archived / foreign / unknown). |
| `NamesForAccount` (read-model labelling) | Returns only the given account's categories (active **and** archived, by design — an item keeps its label after the category is archived). Called with the same `accountID` the timeline handler already scoped `Timeline`/`Comparison` to; never exposed as an endpoint of its own, so there is no way to fetch another account's names through it. |
| New FKs | `tasks.category_id`, `habits.category_id`, `goals.category_id` → `categories.id ON DELETE RESTRICT`, nullable, each indexed `(account_id, category_id)`. RESTRICT is inert today (no hard delete exists anywhere in the API — `v1.md §2`) but guards the theoretical case, matching `time_blocks.category_id`. None of the three FKs can be satisfied by a foreign category id, because the FK only proves the category *exists somewhere*, not that the caller owns it — ownership is enforced application-side by `AssignableToAccount` before the write, and cross-account isolation tests confirm a foreign category is rejected rather than silently attached. |
| `CountByCategory` (per-module `categories.Counter`) | Each of `tasks`, `habits`, `goals`, `timeline` implements it as `WHERE account_id = $1 AND category_id IS NOT NULL [AND archived_at IS NULL] GROUP BY category_id` — scoped, no cross-account rows possible. Isolation covered by `TestCountByCategory` in all four packages (a second account's items never appear in the first account's counts). |
| `GET /api/categories/overview` composition handler | Lives in `cmd/server` (not inside `categories`, per ADR-0009's "categories stays pure" rule), takes `accountID` from `reqctx.IdentityFrom(ctx)` exactly like every other handler, and passes that one value into `categories.List` and all four `Counter` calls — no per-call trust boundary to get wrong. Mounted behind `read` (auth only, no CSRF needed for a `GET`). Covered end-to-end by `TestCategoriesOverview` in `cmd/server`. |
| Validation | Category `name` 1–60 trimmed, case-insensitive active-unique; `colour`/`icon` ≤ 40 chars, free-form keys the backend never interprets (no injection surface — never used in a query, path, or template, only stored/echoed as `text`). `category_id` on every consuming endpoint is parsed as a UUID before use; a malformed one → 400, never reaches SQL. |
| CSRF / method | Category writes (`POST`/`PATCH`/`archive`) and the habit's new `PUT /api/habits/{id}/category` all go through `write` (auth + CSRF); every `GET` (including the overview) goes through `read` (auth only). |
| Error exposure | `writeServiceError` in `categories`, `tasks`, `habits`, `goals` maps `*ValidationError` → 400 and the not-found sentinel → 404; anything else is a generic `500 INTERNAL` with detail logged server-side only (ADR-0002) — unchanged pattern, no new leak surface. |
| Resource use | `CountByCategory` and `NamesForAccount` are each one indexed, `GROUP BY`/scan query over a personal account's rows (N2 scale); the overview handler issues five such queries per request, all cheap and none N+1 against another table. |
| New surface | No new secrets, external calls, or dependencies. `sqlc.yaml` gained one block (`categoriesdb`) using the same overrides as every other module. |

## Re-verification

`go build` · `go vet` · `golangci-lint run` (0 issues) · `sqlc diff` (clean) ·
`gofmt -l` (clean) · `go test ./...` (all packages, including the new
`internal/categories` package and `cmd/server`'s first test) — all green. CP 1–4 also
walked live against a real server with curl (category create with colour/icon, task/
habit/goal category assignment, foreign/unknown category → 400, habit category
clear).
