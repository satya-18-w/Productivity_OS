# Security review — MX4 (Notes module)

Date: 2026-09-05 · Reviewer: Claude (manual) · Scope: new `internal/notes` module and
its four `/api/notes*` routes; the `export.NotesReader` composition wiring it into
`GET /api/export`.

> The simplest module in the codebase — single table, no cross-module dependency, no
> checker interface, no derived computation. The only genuinely new surface is the
> export composition, which follows an already-proven pattern (`GoalsReader`). The
> isolation question this raises: does anything about a note's plain-text, no-linkage
> shape create a gap the existing per-module review pattern wouldn't catch? That is
> this review's focus.

## Verdict

**No findings.** Every query in `internal/notes/queries.sql` is `WHERE account_id =
$1 [AND id = $2]` — identical shape to every other single-table module in the
codebase (`goals`, `categories`). `UpdateNoteFields`/`DeleteNote` are `:execrows`,
mapping zero affected rows to `ErrNoteNotFound` rather than silently succeeding or
leaking another account's row count. There is no linkage to any other entity, so
there is no cross-module validation surface to get wrong (no `CategoryChecker`- or
`GoalChecker`-equivalent needed) — the smallest possible attack surface of any module
built so far.

## Checked and OK

| Area | Finding |
|---|---|
| Account scoping | `CreateNote` takes `accountID` from the caller (ultimately `reqctx.IdentityFrom`, never the request body) and writes it directly; `UpdateNoteFields`/`DeleteNote`/`ListNotes` all filter `WHERE account_id = $1`. No query in the module omits this predicate. |
| Not-found vs. cross-account leak | `UpdateNote`/`DeleteNote` return `ErrNoteNotFound` (→ 404) when the `execrows` count is 0 — this is the same outcome whether the note doesn't exist at all or belongs to another account, so no response ever distinguishes "not yours" from "doesn't exist" (no account-enumeration side channel). Covered by `TestNoteIsolation`. |
| Input validation | Title required, bounded to 200 chars; body bounded to 20,000 chars — both enforced server-side in `validateInput` before any query runs (`TestNoteValidation`). No HTML/script content is treated specially since notes are plain text end-to-end with no rendering concern at the API layer (rendering safety is the frontend's responsibility, standard for every text field in this codebase). |
| Hard delete | `DeleteNote` is a real `DELETE`, not a soft-delete flag — matches the approved scope ("no trash"); there is no `deleted_at` column or filtered-list logic to get wrong. |
| Export composition (`NotesReader`) | `export.Service.Export` calls `s.notes.ListNotes(ctx, accountID)` with the exact same `accountID` used for every other reader in the same call (`reports`' and the rest of `export`'s own established single-account-id-per-request pattern, `docs/security-review-m7.md`). A note from one account can never appear in another account's export bundle, because `ListNotes` itself is already account-scoped — the composition layer adds no new scoping logic to get wrong. |
| Error exposure | `writeServiceError` in `notes` maps `*ValidationError` → 400 and `ErrNoteNotFound` → 404, identical to every other module's pattern; anything else falls through to `httpx.WriteError`'s default 500 with server-side-only detail. |
| New surface | One migration (`000013_notes`, additive: one new table, no destructive change, no data migration). No new secrets, external calls, or trust boundaries. One new `sqlc.yaml` block, scoped only to `internal/notes/queries.sql`. |
| Scope note (not a finding) | `docs/design/design-system.md` §6.4 separately ratified Notes as "must not be implemented" — a frontend-side governance conflict, not a security or isolation issue. Recorded in `planning.md`'s MX4 section; the frontend agent still needs to sync that document. |

## Re-verification

`go build` · `go vet` · `golangci-lint run` (0 issues — one `staticcheck` S1016 finding
fixed during review, a struct-literal-to-conversion simplification with no behavioural
change) · `sqlc diff` (clean) · `gofmt -l` (clean) · `go test ./...` — all 18 packages
green, including `TestNoteLifecycle`, `TestNoteValidation`, `TestNoteIsolation`,
`TestNoteEndpoints` (new in `internal/notes`) and the extended
`TestExport_RoundTripCompleteness`/`TestExport_Isolation` (now asserting notes appear
only for the owning account). CP 1–2 walked live against a running server with curl: a
note created, edited, listed, and deleted (list empty afterward, confirming the hard
delete); a second note created and confirmed present in the `notes` array of a live
`GET /api/export` response.
