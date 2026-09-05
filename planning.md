# Productivity OS — Implementation Plan

The single working plan. Build top to bottom. Each phase has sub-tasks and one or more
**checkpoints** — an observable result that proves the phase works. Stop at each
checkpoint. Tick boxes as we go.

- **Scope:** full V1 (§1–§14 of `docs/requirements/v1.md`), milestones M1–M8 in
  `docs/roadmap.md` order.
- **Architecture:** fixed by ADR-0001…0007 in `docs/decisions/`. Do not re-litigate.
- **Rules:** `CLAUDE.md` (hard constraints, module boundaries, isolation, migrations).

---

## Build-time decisions (defaults — product owner may change any)

| Topic | Decision | Source |
|---|---|---|
| Default timezone at registration | Browser-detected IANA name sent by the client; server validates; fallback `UTC` | Q4 (default) |
| Password policy | 6–128 chars; ≥1 lowercase, ≥1 uppercase, ≥1 special (non-alphanumeric); no denylist | **Q6 resolved 2026-09-04** — `v1.md` |
| Session lifetime | 30 days absolute; no idle timeout; no "remember me" | ADR-0004 deferred |
| Session token | 32 bytes from `crypto/rand`, base64url, stored as-is | ADR-0004 |
| CSRF | `SameSite=Lax` cookie + double-submit token header on state-changing requests | ADR-0004 deferred |
| Argon2id params | 19 MiB memory, time=2, parallelism=1, 16-byte salt, 32-byte key | ADR-0004 deferred |
| Login rate limiting | In-memory limiter keyed on email + client IP (5 failures / 15 min) | ADR-0004 deferred |
| Cookie | `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`, name `session` | ADR-0004 |
| Go version | `go 1.26.0` in `go.mod` (`x/crypto` requires it); toolchain auto-manages | ADR-0002 |
| UUID type | `github.com/google/uuid` for account/session ids (added — was not in the original dep list) | — |
| Frontend router | `react-router-dom` v7 (added — 3 routes with guards) | ADR-0006 |
| Frontend data fetching | native `fetch` + a thin typed client wrapper; TanStack Query only if a real need appears | ADR-0006 deferred |
| Frontend asset packaging | `go:embed` the built `web/dist` into the binary | ADR-0001/0006 deferred |

---

## Environment notes

- **Docker** available. Dev Postgres 16 runs on host port **5433** (5432 was already
  taken). Databases `productivity_os` (dev) and `productivity_os_test` (integration)
  exist in the container.
- **Go: `go.mod` requires `go 1.25.0`** — `pgx/v5 v5.10.0` needs it; the toolchain
  auto-upgraded (go1.26.8 fetched). CI must use Go ≥ 1.25.
- **sqlc** not on PATH — the Makefile runs it via `go run …@v1.27.0` (first run fetches).
- `migrate` v4 is used as an embedded library (`cmd/migrate`), no CLI needed.
  `node` 22 + `pnpm` 10 present for the frontend. `golangci-lint` 2.12.2 present.
- Copy `.env.example` → `.env` before running anything.

---

## Target repository layout

```
cmd/server/            process entry point, wiring
internal/
  platform/
    config/            env config loading
    postgres/          pgxpool lifecycle, health
    httpx/             server, middleware, error envelope, request id
    timezone/          IANA validation (M1); date/week bucketing (M2)
  account/             account + session domain module (M1)
    <module>.go        the one public interface
    service.go         behaviour, transaction boundaries
    store.go           sqlc-generated query wrappers
    queries.sql        hand-written SQL for sqlc
db/migrations/         golang-migrate, forward-only
web/                   React + TS + Vite SPA
  src/
  dist/                built assets (embedded)
Makefile
docker-compose.yml     Postgres only
.env.example
sqlc.yaml
.golangci.yml
.github/workflows/ci.yml
```

---

# MILESTONE 1 — Skeleton and authentication

Covers `v1.md §1`, N1, N3, N5 (shell). A user can register, log in, log out, view and
change their timezone, and change their password, with strict per-account isolation.

## Phase 1 — Repo skeleton & local dev

- [x] 1.1  `go mod init` (`github.com/satya-18-w/productivity-os`); Go pinned at `1.22.2` in `go.mod`
- [x] 1.2  Directory layout (`cmd/server`, `internal/platform/{config,httpx}`; rest added per phase)
- [x] 1.3  `docker-compose.yml` — Postgres 16 only, named volume, healthcheck
- [x] 1.4  `.env.example` — `DATABASE_URL`, `PORT`, `ENV`, `SESSION_TTL`, `SHUTDOWN_GRACE`
- [x] 1.5  `Makefile` — db-up/down/reset, migrate(+create/down), sqlc(+diff), run, build, test, lint, web-dev/build
- [x] 1.6  `sqlc.yaml`, `.golangci.yml`, `.gitignore`
- ⛔ **CP 1a** `make db-up` — BLOCKED: Docker not available in this environment (WSL integration off). See notes.
- [x] **CP 1b** `make lint` clean (`golangci-lint run` → 0 issues; `go vet` clean; `gofmt` clean)

## Phase 2 — HTTP server & platform middleware  ✅ COMPLETE

- [x] 2.1  `platform/config` — env load + validation; fatal on missing `DATABASE_URL`, bad `ENV`, bad durations
- [x] 2.2  `platform/httpx` — `http.Server` + `ServeMux`; context-driven graceful shutdown with bounded grace
- [x] 2.3  Middleware: request ID (echoed on `X-Request-Id`) → `slog` request logging → panic recovery
- [x] 2.4  Error envelope `{ "error": { code, message, fields? } }`; `APIError` + `ValidationError`; ≥500 logged with request id, returned as generic `INTERNAL`
- [x] 2.5  `GET /healthz` → 200
- [x] 2.6  `DecodeJSON` — malformed / oversized (>1 MiB) / trailing-data / unknown-field body → 400, never 500
- [x] 2.7  Tests: recoverer (no leak), error writer, `DecodeJSON` table, middleware chain order, server serve + drain
- [x] **CP 2a** server serves `GET /healthz` → 200; `Server.Run` returns nil on shutdown and drains an in-flight request (proven by `TestServer_*`)
- [x] **CP 2b** `TestRecoverer_PanicBecomesGeneric500` — panic → 500 envelope, no secret/stack in body, server keeps serving
- [x] **CP 2c** `go test ./...` green; `golangci-lint` 0 issues

## Phase 3 — Database: pool, migrations, schema  ✅ COMPLETE

- [x] 3.1  `platform/postgres.Open` — ping-verified `pgxpool` from `DATABASE_URL`; `Healthy` for readiness
- [x] 3.2  `golang-migrate` v4 via embedded FS (`db/`); `cmd/migrate` (up/down/drop) backs `make migrate` and the deployed one-shot
- [x] 3.3  Migration `000001` — `accounts` (uuid pk, `citext` email unique, password_hash, timezone, created_at) + `sessions` (token pk, account_id fk ON DELETE CASCADE, created_at, expires_at, last_seen_at); indexes on `account_id` and `expires_at`; all instants `timestamptz`
- [x] 3.4  `GET /readyz` → 200 when the pool pings, 503 otherwise; no account data
- [x] 3.5  `sqlc.yaml` targets `db/migrations` + `internal/account/queries.sql` (queries land in Phase 4)
- [x] 3.6  `internal/platform/pgtest.Pool` — migrates `TEST_DATABASE_URL`, truncates every table per test, skips when unset
- [x] **CP 3a** migrations apply from empty; down→up cycles; `up` idempotent (integration tests + manual)
- [x] **CP 3b** unreachable DB → exit `1`, `"database unreachable: ... SQLSTATE 3D000"`
- [x] **CP 3c** stopping Postgres flips `/readyz` to 503 while `/healthz` stays 200 and the server keeps serving

## Phase 4 — Account module: registration & login  ✅ COMPLETE

- [x] 4.1  `internal/account.Service` interface: `Register`, `Authenticate`, `ResolveSession`, `EndSession`, `EndAllSessions`, `ChangePassword`, `SetTimezone`, `Read` (all implemented; handlers for the last four land in Phases 5–6)
- [x] 4.2  `internal/platform/password` — Argon2id, PHC string format (params embedded), constant-time verify; decoy-hash on the unknown-email path
- [x] 4.3  `internal/account/queries.sql` → `sqlc generate` → `internal/account/accountdb/` (uuid + citext overrides); `sqlc diff` clean
- [x] 4.4  `Register` — validate email / password (Q6 policy) / IANA timezone (`internal/platform/timezone`); case-insensitive uniqueness via `citext` + `23505`; account + session in one tx (`inTx` DBTX seam)
- [x] 4.5  `Authenticate` — lookup by normalized email, verify hash; unknown email and wrong password both return `ErrInvalidCredentials`
- [x] 4.6  `POST /api/accounts` → 201 + `Set-Cookie` (HttpOnly, SameSite=Lax, Secure in prod); 409 `EMAIL_ALREADY_REGISTERED`; 400 `VALIDATION_ERROR` with `fields`
- [x] 4.7  `POST /api/sessions` → 200 + `Set-Cookie`; 401 `INVALID_CREDENTIALS` with byte-identical bodies
- [x] 4.8  `internal/platform/ratelimit` in-memory fixed-window limiter; 6th failed login (5/15min per email+IP) → 429 `RATE_LIMITED`
- [x] 4.9  36 tests: service integration (register/dup/validation/default-tz/authenticate/resolve/expiry/logout/change-password/set-tz) + HTTP (201/409/400/malformed/no-enumeration/success/rate-limit) + password + ratelimit units
- [x] **CP 4a** register → 201 + `session` cookie; same email different case → 409 (curl + tests)
- [x] **CP 4b** login good → 200 + cookie; bad → 401, unknown-email == wrong-password body (curl + tests)
- [x] **CP 4c** `go test ./...` green; `golangci-lint` 0 issues

*Infra: `pgtest` now provisions a **database per test package** (`pos_test_<pkg>`) so integration
tests across packages don't deadlock on shared tables.*

## Phase 5 — Sessions & auth middleware

- [x] 5.1  `ResolveSession` — token → session row; rejected if missing or `now >= expires_at`; bumps `last_seen_at`
- [x] 5.2  `Handler.RequireAuth` middleware — `session` cookie → `ResolveSession` → `account.Identity` in ctx (only source of the acting account); else `401 UNAUTHENTICATED`
- [x] 5.3  `DELETE /api/sessions/current` → 204, deletes the row, clears session + CSRF cookies
- [x] 5.4  `Handler.RequireCSRF` — readable `csrf_token` cookie issued with each session; unsafe methods require a matching `X-CSRF-Token` header (constant-time) or `403 FORBIDDEN`
- [x] 5.5  `authflow_test.go` — no/garbage/expired cookie → 401; logout kills the cookie; write without CSRF → 403
- [x] **CP 5a** `GET /api/account` returns the caller's profile with the cookie, 401 without (curl + test)
- [x] **CP 5b** after logout the same cookie returns 401
- [x] **CP 5c** a session past `expires_at` returns 401

## Phase 6 — Account management endpoints  ✅ COMPLETE

- [x] 6.1  `GET /api/account` → `{ email, timezone }`, caller's own (`IdentityFrom(ctx)`), auth-only (no CSRF on a read)
- [x] 6.2  `PUT /api/account/timezone` — IANA-validated → 204; invalid or empty → `400 VALIDATION_ERROR`
- [x] 6.3  `PUT /api/account/password` — verify current, validate new, replace hash + `EndAllSessions` in one tx → 204 + cookies cleared; wrong current → 401; weak new → 400
- [x] 6.4  Tests: timezone round-trip + invalid; password change ends every session incl. current; wrong current → 401
- [x] **CP 6a** change timezone via API, re-`GET` shows the new value (curl + test)
- [x] **CP 6b** change password → 204, prior cookie dead, re-login required (curl + test)

## Phase 7 — Isolation & hardening  ✅ COMPLETE

- [x] 7.1  `isolation_test.go` — read-is-caller-own, timezone/password/logout don't cross, per-endpoint (7 tests)
- [x] 7.2  `account_id` in body → 400 (unknown field); in query / `X-Account-Id` header → ignored
- [x] 7.3  All access via `sqlc` parameterized queries; injection-shaped email/password inert (`TestIsolation_SQLInjectionShapedValuesAreInert`)
- [x] 7.4  `TestHardening_InternalErrorDoesNotLeak` — pool closed mid-request → `500 INTERNAL`, generic body, detail logged by request id; malformed/oversized bodies → 400
- [x] 7.5  Security review — `docs/security-review-m1.md`. Bundled skill couldn't run (no git remote); manual review done. **F1** (hash session tokens), **F2** (per-IP throttle on both auth endpoints), **F3** (limiter map sweep), **F6** (best-effort `last_seen_at`) fixed. F4/F5/F7 accepted or deferred to ADR-0008.
- [x] **CP 7a** isolation suite green for every M1 endpoint
- [x] **CP 7b** security review complete; all fixed findings re-verified green

## Phase 8 — Frontend shell  ✅ built (interactive check pending — no browser in this env)

- [x] 8.1  `web/` — Vite 7 + React 19 + TS scaffold; `src/api.ts` typed client (same-origin `fetch`, CSRF token read from cookie → `X-CSRF-Token`); `react-router-dom` v7 added
- [x] 8.2  Register screen — email, password (min 12), timezone pre-filled from `Intl.DateTimeFormat().resolvedOptions().timeZone`; field-level errors
- [x] 8.3  Login screen — email, password; generic error, 429 message
- [x] 8.4  Authenticated shell — email + timezone, change-timezone form, change-password form, log out; no product features
- [x] 8.5  Route guards in `App.tsx` — auth state from `GET /api/account` on load (`auth.tsx`); logged-out → `/login`, logged-in → `/` from auth pages
- [x] 8.6  `web/embed.go` `//go:embed all:dist`; `httpx.SPA` serves assets + falls back to `index.html`; `cmd/server` mounts it at `/`. `web/dist/.gitkeep` (via `web/public/`) keeps it compiling pre-build
- [x] 8.7  `styles.css` — single stylesheet, max-width card, one media query, dark-mode; targets 375 / 1280 (not visually verified — see below)
- [~] **CP 8a/8b** register → shell → change tz / password → logout: **API paths all verified** via curl + Go tests; the **browser click-through needs a human** (no browser here). `make web-build && PORT=8090 ./bin/server` serves it; `/`, `/login`, `/assets/*.js` all 200, `/api/*` unaffected.
- [~] **CP 8c** responsive: CSS written for it; **needs a visual check at 375 / 1280** (or a Playwright test — a deliberate dep decision).

## Phase 9 — CI  ✅ written (green pending first push — no remote configured)

- [x] 9.1  `.github/workflows/ci.yml` — `backend` job (Postgres 16 service; `go build`, `go vet`, `golangci-lint` v2.13.2, `sqlc diff`, `go test`) + `frontend` job (`pnpm typecheck`, `pnpm build`)
- [x] 9.2  Runs on push to `main` and every PR
- [~] **CP 9** locally equivalent commands all pass (`sqlc diff` clean, tests green with CI-style `TEST_DATABASE_URL`); **goes green once pushed to GitHub** — repo has no remote yet

### M1 done when

CP 1a–9 all met · isolation suite green · security review clean · CI green · a person can
do the whole §1 flow through the browser.

---

# MILESTONE 2 — Categories and the timeline core

Covers `v1.md §2, §3, §4, §5, §6` and completes **N4**. Depends on M1. This is the
product's core — plan a day, log what happened, see the gap.

## M2 decisions (defaults — product owner confirms before Phase 4)

| Topic | Default | Rationale / source |
|---|---|---|
| **Module structure** | One `internal/timeline` module owns categories + blocks + the read models. | In V1 categories label *only* time blocks (`v1.md`: tasks/habits/goals carry none). One bounded context; one public interface (ADR-0002). |
| **Q7** — overlapping-block totals | **Sum of block durations.** Two overlapping 1h blocks = 2h. | Deterministic arithmetic (`v1.md §6`); wall-clock-covered needs interval merging and hides double-booking. Also settles the same question for M7. |
| **Q8** — no-category blocks | **Explicit "Uncategorized" bucket** in the comparison. | Otherwise logged time silently vanishes from the per-category view. Also settles M7. |
| **Q9** — future-dated `actual` blocks | **Allowed.** No "not after today" check. | Future *planning* is already in V1; this is a personal tool; a date check needs fragile "today in which tz at request time" logic. (Also unblocks M4/M6 the same way.) |
| **Midnight-spanning totals** | A block counts toward a date only for the portion of its `[start, end)` inside that date's window in the account tz. `23:00–02:00` → 1h to day N, 2h to day N+1. | This *is* N4's "blocks that span midnight produce correct totals". |
| **Category name uniqueness** | Unique active name per account, case-insensitive. Archived names don't block a new one. | Not stated in `v1.md`; avoids two active "Gym"s. |
| **Assigning an archived category** | Rejected on create/edit. Existing assignments keep it (`v1.md §2`). | `v1.md §2`: archived = "no longer offered when assigning". |
| **`date` query param** | `?date=YYYY-MM-DD`, resolved in the account's timezone. | ADR-0005. |

> Q7, Q8, Q9 are open questions in `v1.md`. Confirm the defaults (or amend `v1.md`)
> before Phase 4 ships — Phases 1–3 don't depend on them.

## Phase 1 — Timezone & calendar foundation (`internal/platform/timezone`) — completes N4  ✅ COMPLETE

- [x] 1.1  `Date` value type (Y/M/D) + `ParseDate` / `String` (`YYYY-MM-DD`) / `Today(loc)` / `DateAt`
- [x] 1.2  `LoadLocation(name)` cached (`sync.Map`); `Valid` now delegates to it; `_ "time/tzdata"` embedded so behaviour is host-independent
- [x] 1.3  `DayWindow(Date, *Location) (start, end)` — `[00:00, next-day 00:00)` instants, DST-aware (23h / 25h / ±30min days)
- [x] 1.4  `ISOWeekWindow(Date, *Location)` — `[Mon 00:00, next Mon 00:00)`
- [x] 1.5  `DateAt`, `ISOWeekAt(instant, *Location) (year, week)`
- [x] 1.6  `OverlapSeconds(bStart, bEnd, wStart, wEnd) float64` — block duration clipped to a window
- [x] **CP 1a** 10 tests green: Kolkata/UTC 24h; NY 23h spring-forward + 25h fall-back; Lord Howe ±30min; Chatham 45-min offset; `23:00–02:00` → 1h day N / 2h day N+1; ISO week `2021-01-01` → Monday `2020-12-28` (W53)
- [x] **CP 1b** `go test ./...` green across all packages; `golangci-lint` 0 issues

## Phase 2 — Categories (schema + service + API)  ✅ COMPLETE

- [x] 2.1  Migration `000002_categories` — `id`, `account_id` fk (cascade), `name`, `archived_at` null, `created_at`; partial unique index `(account_id, lower(name)) WHERE archived_at IS NULL`
- [x] 2.2  `internal/timeline` module + `Service` interface; `queries.sql` → `timelinedb` (2nd sqlc block); `internal/platform/reqctx` extracted so both modules share the request principal
- [x] 2.3  `CreateCategory` / `RenameCategory` / `ArchiveCategory` / `ListActiveCategories` — account id from `reqctx.IdentityFrom`; name trimmed, 1–60 chars; duplicate active name → `ErrCategoryNameTaken` (case-insensitive); archived name is reusable
- [x] 2.4  HTTP: `GET/POST /api/categories`, `PATCH /api/categories/{id}`, `POST /api/categories/{id}/archive` — wrapped in the account module's auth (+ CSRF on writes) via a `Protector` passed from `cmd/server`
- [x] 2.5  15 tests: create/list/order, dup (case-insensitive), archived-name-reuse, validation, rename + collision + not-found, archive + re-archive, cross-account isolation, HTTP 201/409/400/204/404
- [x] **CP 2** create / rename / archive / list green (service + HTTP tests); isolation green; `sqlc diff` + lint clean

## Phase 3 — Time blocks (schema + service + API)  ✅ COMPLETE

- [x] 3.1  Migration `000003_time_blocks` — `id`, `account_id` fk (cascade), `kind` CHECK `in ('planned','actual')`, `starts_at`/`ends_at` `timestamptz`, `category_id` null fk (RESTRICT), `created_at`; `CHECK (ends_at > starts_at)`; indexes `(account_id, starts_at)` and `(category_id)`
- [x] 3.2  `AddBlock` / `EditBlock` / `DeleteBlock` — account-scoped; `kind` set only at creation; `end > start`; a supplied `category_id` must be the caller's and active (`CountAssignableCategory`)
- [x] 3.3  Q9: no future-date gate on `actual` blocks (per the decision table)
- [x] 3.4  HTTP: `POST /api/blocks`, `PUT /api/blocks/{id}`, `DELETE /api/blocks/{id}` (RFC 3339 instants); `pgtype.UUID`/`pgtype.Text` ↔ domain helpers
- [x] 3.5  10 tests: planned + actual; `21:00–02:00` cross-midnight round-trip; kind / end-after-start validation; category must be owned + active (foreign / archived / unknown all rejected); edit + isolation; delete + isolation
- [x] **CP 3** cross-midnight planned block with a category → 201, bad range → 400 (curl); full service + isolation tests green; `sqlc diff` + lint clean

## Phase 4 — Timeline & comparison read models (§5, §6)  ✅ COMPLETE

- [x] 4.1  `Timeline(ctx, accountID, Date)` → planned + actual blocks overlapping the date's `DayWindow` in the account tz; each carries instants + category id/name. `AccountZone` interface keeps timeline decoupled from account (wired in `cmd/server`)
- [x] 4.2  `Comparison(ctx, accountID, Date)` → per-category `{plannedSeconds, actualSeconds, differenceSeconds}`, `OverlapSeconds`-clipped to the day window; explicit **Uncategorized** row (Q8, nil id); overlaps summed (Q7); named categories sorted, Uncategorized last
- [x] 4.3  HTTP: `GET /api/timeline?date=`, `GET /api/comparison?date=` — date via `timezone.ParseDate`, resolved in the account tz; missing/invalid → 400
- [x] 4.4  8 read-model tests: planned/actual split; midnight block on both days; per-category totals + difference; Uncategorized bucket; overlapping blocks summed; `23:00–02:00` → 1h day N / 2h day N+1; DST spring-forward counts 2 real hours not 3 wall-clock; isolation
- [x] **CP 4** comparison correct for a Kolkata normal day (planned 7200 / actual 10800 / diff 3600, curl) + DST + midnight cases (tests); full suite + `sqlc diff` + lint green

## Phase 5 — Frontend: categories  ✅ built (browser click-through pending)

- [x] 5.1  `api.ts` — category + block + timeline/comparison methods and types; `Categories` page — list active, create, rename inline, archive (with confirm)
- [x] 5.2  `AuthLayout` with a top nav (Account · Categories) wrapping the authenticated routes; `Shell` split into `pages/Account.tsx`; guards unchanged
- [~] **CP 5** routes serve (`/categories` → 200 SPA), API round-trips through the SPA origin, `tsc` + `vite build` clean — visual create/rename/archive needs a browser

## Phase 6 — Frontend: timeline & blocks  ✅ built (browser click-through pending)

- [x] 6.0  **Block API moved to wall-clock** — `{kind, date, start "HH:MM", end "HH:MM", ends_next_day, category_id}`; the Go handler converts to instants via `AccountZone` (ADR-0005: zero client-side tz math). Timeline response carries `start_minute` / `end_minute` / `from_prev_day` / `to_next_day` + `local_*` for the editor, all computed server-side.
- [x] 6.1  Date nav — prev / next / today + a `<input type=date>`
- [x] 6.2  `Timeline` page — 24-hour grid, Planned vs Actual lanes (colour-coded), blocks positioned by server-computed minutes, ▲/▼ markers for spill
- [x] 6.3  `BlockForm` — kind (locked on edit), date, start/end time inputs, "ends next day", category `<select>`; create / save / delete
- [x] 6.4  Responsive CSS — axis + 2 lanes, horizontal-scroll wrapper, single-column form under 520px
- [~] **CP 6a** create/edit/delete verified via API (NY DST day + cross-midnight block position correctly: `09:00–11:00` → `13:00Z–15:00Z`, minute 540–660; `22:00`→`01:30+1` → minute 1320–1440, `to_next_day`); browser interaction pending
- [~] **CP 6b** CSS written for 375 / 1280; visual pass pending (no browser)

## Phase 7 — Frontend: comparison + M2 wrap  ✅ COMPLETE

- [x] 7.1  Comparison table on the Timeline page — category · planned · actual · difference (colour-coded), Uncategorized row, Total footer; reloads with the date
- [x] 7.2  Full regression — `go test ./...` (8 pkgs), `pnpm typecheck` + `vite build`, `sqlc diff`, `golangci-lint` — all green
- [x] 7.3  `docs/security-review-m2.md` — isolation-focused; **no findings**. Every read/write of `categories` + `time_blocks` account-scoped; cross-account `category_id` rejected at the app layer + tested; writes behind auth + CSRF
- [x] 7.4  CI needs no change — `go test ./...` picks up `internal/timeline`, `sqlc diff` covers the new queries, migrations `000002`/`000003` run in the integration harness, the frontend job covers the new pages
- [x] **CP 7** comparison correct end-to-end (DSA planned 3h / actual 1.5h / diff −5400; Uncategorized actual 1h; named sorted, Uncategorized last); full suite + typecheck + `sqlc diff` + review clean

### M2 done when

CP 1–7 met · a person can manage categories, plan and log a day's blocks (including
across midnight), and see the timeline and the planned-vs-actual comparison ·
isolation tests green for `categories` and `time_blocks` · CI green.

**M2 status: backend + frontend + security review complete.** Remaining before full
sign-off (same as M1): a browser click-through of the timeline/categories/comparison UI,
and CI green on a first push. Q7 / Q8 / Q9 were built to the recommended defaults — if
the product owner amends any, revisit Phases 4 and 7.

---

# MILESTONE 3 — Tasks and the Kanban board

Covers `v1.md §7, §8`. Depends on M1. Independent of M2.

## M3 decisions (defaults — product owner confirms)

| Topic | Default | Rationale / source |
|---|---|---|
| **Module** | New `internal/tasks` module (own bounded context; tasks carry no category in V1). | ADR-0002. |
| **Q5** — ordering within a column | **Unspecified / no manual ordering.** Columns list tasks newest-first (`created_at desc`). No drag-to-reorder within a column — only drag *between* columns. | `v1.md §8` scope boundary: "Task ordering within a column is not a product requirement in V1." |
| Initial state | New tasks start in **`BACKLOG`**. | Not stated in `v1.md`; the leftmost column. |
| **Q10** — state-transition history | **Record every state change** as a `task_transitions` row `{task_id, account_id, from_state, to_state, at timestamptz}`, including creation (`from = NULL`). A same-state "move" records nothing. | `v1.md` needs "tasks that entered `DONE` that week" (M6) and "task throughput" (M7). Recording all transitions keeps every later interpretation open. |
| Q10 — "entered DONE in a range" (M7/M6 will confirm) | **Distinct tasks with ≥ 1 `→ DONE` transition whose `at` is in the range.** A task that bounces `DONE → IN_PROGRESS → DONE` in one range counts once. | Most meaningful "throughput"; the raw transition rows still allow a per-event count if the owner prefers. Confirm at M7. |
| Move interaction (frontend) | A **state `<select>` on each card** is the primary move control (accessible, works on mobile); HTML5 drag-and-drop between columns as a wide-screen enhancement. No DnD library. | `v1.md §8` ("move a task from any column to any other"); ADR-0006 (no libraries without need); N5 (375px). |
| Field bounds | title required ≤ 200 chars; description optional ≤ 5000; due date optional plain `date`. | `v1.md §7` ("plain date, no time, no reminder"). |

> Q5 and Q10 are open questions in `v1.md`. The defaults above stay within its scope
> boundaries; confirm them (or amend `v1.md`) — Phases 1–2 encode the transition-log
> design that M6/M7 depend on.

## Phase 1 — Tasks schema + service  ✅ COMPLETE

- [x] 1.1  Migration `000004_tasks` — `tasks` (+ `description`, `due_date date`, `state` CHECK, `created_at`/`updated_at`) + `task_transitions` (`task_id` + `account_id` fk cascade, `from_state` null, `to_state`, `at`); indexes `(account_id, created_at DESC)`, `(account_id, to_state, at)`
- [x] 1.2  `internal/tasks` module + `Service`; `queries.sql` → `tasksdb` (3rd sqlc block)
- [x] 1.3  `CreateTask` (→ `BACKLOG` + creation transition, one tx), `UpdateTask`, `MoveTask` (any→any; records a transition unless unchanged; one tx), `DeleteTask` (transitions cascade), `Board` (4 columns in fixed order, `created_at DESC`)
- [x] 1.4  Validation — title 1–200 (trimmed), description ≤ 5000, due date optional plain `date`, state ∈ 4
- [x] 1.5  7 tests — create + creation transition; validation; update + not-found; move through all states + same-state no-op + `DONE`-re-entry counts 2 + invalid state + not-found; delete + cascade; board shape/order/grouping; cross-account isolation (board / update / move / delete)
- [x] **CP 1** service tests green incl. transition log + isolation; full suite + `sqlc diff` + lint clean

## Phase 2 — Task + board HTTP API  ✅ COMPLETE

- [x] 2.1  `GET /api/board`, `POST /api/tasks`, `PATCH /api/tasks/{id}`, `PUT /api/tasks/{id}/state`, `DELETE /api/tasks/{id}` — auth (+ CSRF on writes) via the account `Protector`
- [x] 2.2  `GET /api/board` → `{ columns: [{ state, tasks: [...] }, …] }` in the fixed order `BACKLOG · TODO · IN_PROGRESS · DONE`
- [x] 2.3  9 HTTP tests — create → 201; validation (blank title, bad `due_date`) → 400; update → 204; move through all four → 204; bad state → 400; unknown/malformed id → 404; delete → 204; board shape + order + grouping
- [x] **CP 2** board returns 4 ordered columns; a task moves `BACKLOG → TODO → DONE` and lands in the right column (curl); tests + lint + `sqlc diff` clean

## Phase 3 — Frontend: task list + Kanban board  ✅ built (browser click-through pending)

- [x] 3.1  `api.ts` task methods + types; `Board` page at `/board`; nav gains "Board"
- [x] 3.2  Four columns; each task a card with title, due-date badge, 3-line description clamp; create / edit / delete via `TaskForm`
- [x] 3.3  Move: a state `<select>` on every card (primary, accessible) + HTML5 drag between columns (`text/task-id`, drop-target highlight) — no DnD library
- [x] 3.4  Responsive CSS — `repeat(4, minmax(190px, 1fr))` grid, `board-scroll` wrapper for mobile
- [~] **CP 3a** create + move through all four columns verified via API + 16 tests; browser drag/select interaction pending
- [~] **CP 3b** CSS written for 375 / 1280; visual pass pending (no browser)

## Phase 4 — M3 wrap  ✅ COMPLETE

- [x] 4.1  Full regression — `go test ./...` (9 pkgs), `pnpm typecheck` + `vite build`, `sqlc diff`, `golangci-lint` — all green
- [x] 4.2  `docs/security-review-m3.md` — isolation review of `tasks` + `task_transitions`; **no findings** (transition log written only inside the account-scoped move tx; all queries `account_id`-scoped; cascades verified)
- [x] 4.3  CI needs no change — `go test ./...` picks up `internal/tasks`, `sqlc diff` covers the queries, migration `000004` runs in the integration harness, the frontend job covers the new page
- [x] **CP 4** full suite + review green; `/board` serves the SPA

### M3 done when

CP 1–4 met · a person can create, edit, delete tasks and move them across the four
fixed columns · every state change is recorded in `task_transitions` · isolation tests
green · CI green.

**M3 status: backend + frontend + security review complete.** Pending sign-off (as M1/M2):
browser click-through + first CI run. Q5 (no manual ordering) and Q10 (record all
transitions; throughput = distinct tasks with a `→ DONE` in range) built to the
recommended defaults — M7 confirms the throughput interpretation.

---

# MILESTONE 4 — Habits and streaks

Covers `v1.md §9`. Depends on M1. Q9 and Q11 resolved in `v1.md` (2026-09-04).

## M4 decisions (defaults — product owner confirms)

| Topic | Default | Rationale / source |
|---|---|---|
| **Module** | New `internal/habits`. | ADR-0002; habits carry no category. |
| Completion model | A completion is a row `{habit_id, account_id, on_date}` with `UNIQUE(habit_id, on_date)`. Mark = idempotent insert; unmark = delete. No quantity/note (`v1.md §9`). | `v1.md §9`. |
| **Q9** future-dated completions | Allowed — no date gate. A future completion does **not** extend the current streak (streak anchors on today/yesterday). | Q9 resolution. |
| **Q11** archive/unarchive | Archive sets `archived_at`; completions are untouched. Unarchive clears it; the streak resumes from stored completions. | Q11 resolution. |
| Streak algorithm | In the account tz: `today = Today(loc)`. Anchor = `today` if completed, else `today-1` if completed, else streak 0. Then count consecutive completed dates backwards from the anchor. Completions after `today` are ignored for the current streak. | `v1.md §9` ("most recent day is today or yesterday; one missed date resets to zero; no grace days"). |
| Habit name | required, trimmed, 1–100 chars. No rename in V1 (`v1.md §9` has create + archive, not edit). | `v1.md §9`. |
| Extra on the card | `last_30_days` completed-count — display only, derived from stored data (not a `v1.md` requirement; a small convenience). | — |
| Per-date view | `GET /api/habits?date=YYYY-MM-DD` (default today) — each active habit carries `completed_on_date` for that date + `current_streak`. | `v1.md §9` ("see, for a chosen date, which habits were completed"). |

## Phase 1 — Habits schema + service  ✅ COMPLETE

- [x] 1.1  Migration `000005_habits` — `habits` + `habit_completions` (`UNIQUE(habit_id, on_date)`, both `account_id` fk cascade); indexes `(account_id, archived_at)`, `(habit_id, on_date)`
- [x] 1.2  `internal/habits` module + `Service`; `queries.sql` → `habitsdb` (4th sqlc block); `AccountZone` reused for the user's "today"
- [x] 1.3  `CreateHabit`, `ArchiveHabit` / `UnarchiveHabit` (completions untouched), `MarkComplete` (idempotent `ON CONFLICT DO NOTHING`), `UnmarkComplete`, `ListActive(accountID, viewDate)` → `HabitView{ streak, completedOnDate, last30Days }`, `ListArchived`
- [x] 1.4  `streak.go` — pure `currentStreak(set, today)` + `countInRange`; one 400-day-window query per habit, streak/30-day computed in Go
- [x] 1.5  10 tests — 2 pure (streak: consecutive, anchor today/yesterday, gap stops, future ignored, 100-run; range count) + 6 integration (create + validation; mark idempotent + unmark; streak via service + break; **archive hides / unarchive restores streak from preserved history (Q11)**; future completion doesn't inflate (Q9); isolation)
- [x] **CP 1** streak tests green (reset + anchor + future-ignored + Q11) + isolation; full suite + `sqlc diff` + lint clean

## Phase 2 — Habits HTTP API  ✅ COMPLETE

- [x] 2.1  `GET /api/habits?date=` (default today via `AccountZone`, returns active `habits` + `archived`), `POST /api/habits`, `POST /api/habits/{id}/{archive,unarchive}`, `PUT`/`DELETE /api/habits/{id}/completions/{date}` — auth (+ CSRF on writes)
- [x] 2.2  8 HTTP tests — create → 201; blank name → 400; mark ×3 (idempotent) → 204; bad date → 400; list `current_streak` + `completed_on_date`; unmark drops streak; archive→hidden / unarchive→back; unknown/malformed id → 404; `?date=` view-date state
- [x] **CP 2** create habit + mark 3 consecutive days (Asia/Kolkata) via curl → `current_streak: 3`, `completed_on_date: true`, `last_30_days: 3`

## Phase 3 — Frontend: Habits page  ✅ built (browser click-through pending)

- [x] 3.1  `api.ts` habit methods + types; `Habits` page at `/habits`; nav gains "Habits"
- [x] 3.2  Date nav (prev/next/today + date input); each active habit a row — a ✓ toggle for the selected date, name, 🔥 streak, `N/30d`, Archive
- [x] 3.3  Create-habit form; archive with a "history is kept" confirm; a "Show archived (N)" disclosure with per-habit Unarchive
- [x] 3.4  Responsive CSS (reuses `.rows` / `.date-nav`; habit-check button, streak/30d nowrap)
- [~] **CP 3a** toggle across days + streak update verified via API + 18 tests; browser interaction pending
- [~] **CP 3b** CSS written; visual pass pending (no browser)

## Phase 4 — M4 wrap  ✅ COMPLETE

- [x] 4.1  Full regression — `go test ./...` (10 pkgs), `pnpm typecheck` + `vite build`, `sqlc diff`, `golangci-lint` — all green
- [x] 4.2  `docs/security-review-m4.md` — **no findings**. Completions can only be written against a caller-owned habit (`assertOwned`); all queries `account_id`-scoped; archive keeps history; mark idempotent
- [x] 4.3  CI needs no change — `go test ./...` picks up `internal/habits`, `sqlc diff` covers the queries, migration `000005` runs in the integration harness, the frontend job covers the new page
- [x] **CP 4** full suite + review green

### M4 done when

CP 1–4 met · a person can create habits, mark/unmark completions per date, archive and
unarchive, and see the current streak for each active habit · isolation tests green ·
CI green.

**M4 status: backend + frontend + security review complete.** Pending sign-off (as
M1–M3): browser click-through + first CI run.

---

# MILESTONE 5 — Goals

Covers `v1.md §10`. Depends on M1. No open questions. The smallest milestone.

**Decisions:** new `internal/goals` module. Progress states `NOT_STARTED` (default) ·
`IN_PROGRESS` · `ACHIEVED` · `ABANDONED`. Title 1–200, description ≤ 5000, target date
optional plain `date`. No linkage to any other entity (`v1.md §10`).

## Phase 1 — Goals backend + API  ✅ COMPLETE

- [x] 1.1  Migration `000006_goals` — `goals` (+ `description`, `target_date date`, `progress` CHECK/default `NOT_STARTED`, `created_at`/`updated_at`); index `(account_id, created_at DESC)`
- [x] 1.2  `internal/goals` module + `Service`; `queries.sql` → `goalsdb` (5th sqlc block)
- [x] 1.3  `CreateGoal` (→ `NOT_STARTED`), `UpdateGoal`, `SetProgress` (validated), `DeleteGoal`, `ListGoals` (newest first)
- [x] 1.4  HTTP: `GET /api/goals`, `POST /api/goals`, `PATCH /api/goals/{id}`, `PUT /api/goals/{id}/progress`, `DELETE /api/goals/{id}` — auth (+ CSRF on writes)
- [x] 1.5  13 tests — lifecycle (create → cycle all 4 states → edit → delete), validation, isolation (update/progress/delete), HTTP (201 / 400 / 204 / 404)
- [x] **CP 1** create a goal, cycle all four progress states, edit, delete (curl + tests); isolation green

## Phase 2 — Frontend: Goals page + M5 wrap  ✅ COMPLETE

- [x] 2.1  `api.ts` goal methods + types; `Goals` page at `/goals`; nav gains "Goals"
- [x] 2.2  Goal list — each a card: title, `🎯 target date`, description, colour-coded progress chip + a progress `<select>`; create / edit / delete via `GoalForm`
- [x] 2.3  `docs/security-review-m5.md` — **no findings**. Full regression green (11 Go packages, `pnpm typecheck` + `vite build`, `sqlc diff`, lint)
- [x] **CP 2** `/goals` serves the SPA; create → `NOT_STARTED`, set progress → `ACHIEVED` verified via curl; full suite + review green

### M5 done when

CP 1–2 met · a person can create, edit, delete goals and set any of the four progress
states · isolation tests green · CI green.

**M5 status: complete.** Pending sign-off (as M1–M4): browser click-through + first CI run.

---

# MILESTONE 6 — Daily and weekly reviews

Covers `v1.md §11, §12`. Depends on M2, M3, M4. Q1, Q2, Q9, Q10 resolved in `v1.md`
(Q1/Q2 with **placeholder prompt wording** — the product owner replaces the text).

## M6 decisions

| Topic | Default | Rationale / source |
|---|---|---|
| Module | New `internal/reviews` — owns only the review answers + the fixed prompt sets. | ADR-0002. |
| Prompt sets | Go constants (`dailyPrompts`, `weeklyPrompts`), each `{key, text}`. Not in the DB. Q1/Q2 wording is placeholder. | `v1.md §11/§12` ("fixed, non-editable, defined in the review SPEC"). |
| Answer storage | `answers jsonb` — a `{promptKey: text}` map, saved/loaded whole. | Answers are only ever read/written as a full review. |
| Identity | daily → `UNIQUE(account_id, on_date)`; weekly → `UNIQUE(account_id, iso_year, iso_week)`. | `v1.md`. |
| Reference data | **Assembled by the frontend** from range-capable read endpoints — not by the reviews module. Daily reuses `/api/comparison?date=` + `/api/habits?date=`. Weekly uses new range endpoints (below), which M7 also needs. | Keeps `reviews` a thin, well-bounded module; avoids `reviews` composing four other services. |
| "Tasks that entered DONE in a week" (Q10) | Distinct tasks with ≥ 1 `→ DONE` transition whose `at` is in the ISO-week instant range (account tz). | Q10 resolution. |
| Editing | A `PUT` upserts the whole answer set. "Complete" and "edit" are the same operation. Empty answers allowed (a partially-filled review). | `v1.md §11` ("edit a previously completed review"). |

## Phase 1 — Range read methods (foundation for M6 + M7)  ✅ COMPLETE

- [x] 1.1  `timeline`: `ComparisonRange(ctx, accountID, from, to Date)` (per-category planned/actual/diff over `[from 00:00, to+1 00:00)` in account tz, sharing `Comparison`'s bucketing via a new `categoryTotals` helper) + `GET /api/comparison` accepts `?from=&to=` as an alternative to `?date=` (mutually exclusive; response carries `from`/`to` or `date`)
- [x] 1.2  `habits`: `CompletionCountsInRange(ctx, accountID, from, to Date)` → `[]RangeCount{HabitID, Name, Count}` (active + archived, zero-count habits included via a `LEFT JOIN ... FILTER`) + `GET /api/habits/range?from=&to=`
- [x] 1.3  `tasks`: `DoneCountInRange(ctx, accountID, fromInstant, toInstant time.Time)` → distinct tasks with a `→ DONE` in range (`count(DISTINCT task_id)`); new `tasks.AccountZone` on the handler resolves `?from=&to=` dates to instants + `GET /api/tasks/throughput?from=&to=`
- [x] 1.4  Tests — range totals across a DST boundary (`TestComparisonRange_DSTTransitionCountsRealElapsedTime`); distinct-task DONE count with bounce-in/out (`TestDoneCountInRange`); zero-count habit inclusion; `to`-before-`from` validation on all three; isolation on all three
- ✅ **CP 1** verified live via curl (fresh server, real DB): 3 actual blocks across days 15/16/20 → `?from=2025-06-15&to=2025-06-16` correctly sums 1h+2h=3h and excludes day 20; a habit marked on days 15+16 → `/api/habits/range` returns `count: 2`; a task moved to DONE today → `/api/tasks/throughput?from=today&to=today` returns `done_count: 1`. `go build`, `go vet`, `golangci-lint` (0 issues), `sqlc diff` (clean), `gofmt -l` (clean), `go test ./...` all green.

## Phase 2 — Reviews module (schema + service)  ✅ COMPLETE

- [x] 2.1  Migration `000009_reviews` — `daily_reviews` (id, account_id fk, on_date, answers jsonb default `{}`, created_at, updated_at; `UNIQUE(account_id, on_date)`) + `weekly_reviews` (…, iso_year, iso_week int with `CHECK (iso_week BETWEEN 1 AND 53)`, …; `UNIQUE(account_id, iso_year, iso_week)`). Applied to dev + test DBs. *(000007 = categories colour/icon, 000008 = entity `category_id` — both MX1.)*
- [x] 2.2  `internal/reviews` module + `Service`; `DailyPrompts`/`WeeklyPrompts` constants (Q1/Q2 placeholder wording verbatim from `v1.md`); `queries.sql` → `reviewsdb` (7th sqlc block, `answers` as `[]byte` JSON)
- [x] 2.3  `GetDaily(date)` / `SaveDaily(date, answers)` (upsert), `GetWeekly(year, week)` / `SaveWeekly(year, week, answers)`; unknown prompt keys dropped by `filterKnown`; a never-saved date/week returns a blank review (empty answers, no `ErrNotFound` — "not found" isn't a real state for a review, an unanswered date just has no answers yet)
- [x] 2.4  Tests (`internal/reviews/service_test.go`, 12 cases) — save + reload; upsert overwrites (edit replaces, not appends); unknown keys dropped; empty-answers save allowed; different weeks/years independent; invalid `iso_week` (0/54/-1) → `*ValidationError`; isolation (daily + weekly)
- ✅ **CP 2** save a daily review, reload it, edit it; isolation green

## Phase 3 — Reviews HTTP API  ✅ COMPLETE

- [x] 3.1  `GET /api/reviews/daily?date=` → `{date, prompts:[{key,text}], answers:{key:text}, updated_at?}`; `PUT /api/reviews/daily?date=` (body `{answers}`) → 204
- [x] 3.2  `GET`/`PUT /api/reviews/weekly?year=&week=` — same shape (`iso_year`/`iso_week` instead of `date`)
- [x] 3.3  HTTP tests (`internal/reviews/http_test.go`, 3 cases) — blank before save → prompts + `{}`; save → 204; reload → answers, unknown key gone; missing/bad `date` or `year`/`week` → 400; out-of-range week → 400; cross-account isolation
- ✅ **CP 3** verified live via curl (fresh server, real DB): blank daily review shows all 4 prompts; saved answers round-trip with `updated_at` set; weekly review round-trips independently; editing replaces the whole answer set (not a merge)

## Phase 4 — Frontend: Reviews page  *(owned by the parallel frontend agent — see `planning.md` Appendix A.4; not touched here)*

- [ ] 4.1  `api.ts` review + range methods; `Reviews` page at `/reviews`; nav gains "Reviews"
- [ ] 4.2  Daily / Weekly toggle; date picker (daily) or week stepper (weekly); a prompt + `<textarea>` per prompt; Save
- [ ] 4.3  Reference panel — daily: time-by-category + habits done that day; weekly: time-by-category, habit counts, tasks → DONE that week
- [ ] 4.4  Responsive 375 / 1280
- ✅ **CP 4** complete a daily review with the day's totals shown; view/edit a past one; same for a weekly review

The backend contract this phase needs is complete and stable as of Phase 3: `GET/PUT
/api/reviews/daily?date=`, `GET/PUT /api/reviews/weekly?year=&week=`, plus the M6
Phase 1 range endpoints (`GET /api/comparison?from=&to=`, `GET /api/habits/range?from=&to=`,
`GET /api/tasks/throughput?from=&to=`) for the reference panels.

## Phase 5 — M6 wrap (backend scope)

- [x] 5.1  Backend regression — `go build`, `go vet`, `golangci-lint` (0 issues), `sqlc diff` (clean), `gofmt -l` (clean), `go test ./...` all green. *(`pnpm typecheck` / `vite build` are the frontend agent's to run against Phase 4.)*
- [x] 5.2  `docs/security-review-m6.md` — isolation review of `daily_reviews` + `weekly_reviews`, the reviews HTTP surface, and the M6 Phase 1 range endpoints. **No findings** (2 gosec + 1 staticcheck fixed during review).
- [ ] 5.3  Confirm CI covers the new module — deferred, no CI run configured in this environment (same caveat as M1-M5)
- ✅ **CP 5** backend suite + review green

### M6 done when

CP 1–5 met · a person can complete, edit, and view daily reviews (per date) and weekly
reviews (per ISO week), each with that period's totals shown for reference · isolation
tests green · CI green.

**Backend (Phases 1, 2, 3, 5) is done as of 2026-09-04** — CP 1–3, 5 all verified
(tests + live curl); `go build/vet/lint/sqlc diff/test` all green. **Phase 4
(frontend) is the parallel frontend agent's** — the API contract it needs is
final and stable. M6 as a whole is not "done" until Phase 4 ships.

---

# MILESTONE 7 — Reports

Covers `v1.md §13`. Depends on M2, M3, M4, and M6 Phase 1 (the range read methods
this milestone reuses directly). Q7/Q8 were already resolved (M2); **Q10 resolved
2026-09-04** — see `v1.md` Resolutions: distinct-task count, `count(DISTINCT
task_id)` over `→ DONE` transitions in range, identical to what M6 already
implemented.

## M7 decisions

| Topic | Default | Rationale / source |
|---|---|---|
| Module | New `internal/reports` — read-only composition over `timeline`, `habits`, `tasks`, never writes. | ADR-0002; matches the `internal/categories` overview composition already done in MX1. |
| Dependency shape | Consumer-defined interfaces (`TimelineReader`, `HabitsReader`, `TasksReader`, `AccountZone`) naming the exact methods reports needs; `timeline.Service`/`habits.Service`/`tasks.Service` already satisfy them structurally — no adapter code. Return types reuse the producing module's exported structs (`timeline.CategoryTotals`, `timeline.DayTotal`) rather than duplicating near-identical shapes. | Same pattern as every other cross-module dependency (ADR-0002); avoids pointless duplication where the shapes are genuinely identical. |
| New capability needed | `timeline.DailyActualTotals(from, to)` — the one report (#5) nothing existing computes. Everything else (#1-#4) is a thin view over M6 Phase 1's `ComparisonRange` / `CompletionCountsInRange` / `DoneCountInRange`. | v1.md §13. |
| Range validation | Every report takes `(accountID, from, to Date)`; `reports` validates `to` before `from` itself (its own `*ValidationError`) rather than surfacing a foreign module's error type. | Consistent error-type ownership per module (ADR-0002). |
| "Time by category" vs "Planned vs actual by category" | Two distinct reports, same source (`ComparisonRange`) — #1 returns actual-only rows (`CategoryTimeRow`), #2 returns the full planned/actual/difference rows (`timeline.CategoryTotals`), so the two are genuinely different response shapes even though one query produces both. | `v1.md §13` lists them as two separate reports. |
| Habit completion rate | `CompletedDays / days in [from, to]` (inclusive day count), as a float `0..1`. | `v1.md §13` "the number of completed days and the completion rate". |
| Reads-only | No route needs `write`; `Handler.Mount` takes only a `read` protector. | `v1.md §13` "The user can view" — no report is ever edited. |

## Phase 1 — `internal/timeline.DailyActualTotals` (the one new read primitive)  ✅ COMPLETE

- [x] 1.1  `timeline`: `DayTotal{Date, ActualSeconds}`; `DailyActualTotals(ctx, accountID, from, to Date) ([]DayTotal, error)` — actual-kind blocks only, per-day `OverlapSeconds` against each day's `DayWindow` (reuses the same primitives as `Comparison`/`ComparisonRange`); `*ValidationError` if `to` before `from`
- [x] 1.2  Tests (`internal/timeline/readmodel_test.go`, 5 cases) — multi-day totals (planned excluded); a midnight-spanning block splits correctly across its two days; a DST-transition day's total reflects real elapsed seconds (2h, not 3); `to`-before-`from` validation; isolation
- ✅ **CP 1** daily actual totals over a multi-day window match a manual sum, including a DST-transition day — all 5 tests green

## Phase 2 — `internal/reports` module  ✅ COMPLETE

- [x] 2.1  `internal/reports` package: `reports.go` (`Service` interface, 5 methods, `ValidationError`, consumer interfaces `TimelineReader`/`HabitsReader`/`TasksReader`/`AccountZone`), `service.go` (pure composition, no SQL, no sqlc block — reports owns no table). `timeline.Service`/`habits.Service`/`tasks.Service` satisfy the consumer interfaces structurally — no adapter code needed.
- [x] 2.2  `TimeByCategory` / `PlannedVsActualByCategory` (both from one `ComparisonRange` call) · `HabitCompletion` (from `CompletionCountsInRange` + `daysInRange`) · `TaskThroughput` (from `DoneCountInRange`, converting the date range to instants via `AccountZone`) · `DailyActualTotals` (pass-through to `timeline`)
- [x] 2.3  Tests (`internal/reports/service_test.go`, 8 cases, real modules wired over one shared pool) — each report's figures against hand-built fixtures across tasks/habits/timeline; a one-month range crossing the March 2025 DST transition (`TestReports_DSTCrossingRange`); `to`-before-`from` → `*ValidationError` on all 5; isolation on all 5
- ✅ **CP 2** all 5 report methods return correct figures over a one-month range including a DST-crossing range — 8/8 tests green

## Phase 3 — Reports HTTP API + M7 wrap  ✅ COMPLETE

- [x] 3.1  `GET /api/reports/time-by-category?from=&to=`, `GET /api/reports/planned-vs-actual?from=&to=`, `GET /api/reports/habit-completion?from=&to=`, `GET /api/reports/task-throughput?from=&to=`, `GET /api/reports/daily-actual?from=&to=` — all `read`-only (no `write`/CSRF path exists — reports never writes), all `400` on missing/malformed/inverted range
- [x] 3.2  HTTP tests (`internal/reports/http_test.go`, 3 cases) — each route's response shape against known fixtures; range validation (missing/malformed/inverted) across all 5 routes; isolation
- [x] 3.3  Full regression — `go build/vet` clean, `golangci-lint` (0 issues), `sqlc diff` (clean — confirms `reports` introduced no schema change), `gofmt -l` (clean), `go test ./...` all green
- [x] 3.4  `docs/security-review-m7.md` — isolation review of the 5 report routes, the multi-service composition, and the new `DailyActualTotals` primitive. **No findings.**
- ✅ **CP 3** verified live via curl (fresh server, real DB, `America/New_York` account): a real block spanning the March 9 2025 spring-forward transition (01:00→04:00 local, only 2h real elapsed — confirmed at both block-creation and report layers); all 5 reports over the full March 2025 range returned figures consistent with that (time-by-category 7200s, planned-vs-actual 3600/7200/3600, habit-completion 2/31 days, daily-actual showing 7200s on the transition day itself); task-throughput separately confirmed to count a same-day `DONE` task; inverted range → 400. Full suite + review green.

### M7 done when

CP 1–3 met · the five reports (`v1.md §13`) are each queryable over a caller-chosen
range with correct, deterministic figures, including a DST-crossing range · isolation
tests green · `internal/reports` performs no writes and owns no table.

**M7 backend is done as of 2026-09-04.** `v1.md` Q10 fully resolved (was "partly").
No frontend report views exist yet — same split as M6: the parallel frontend agent
owns that, against the now-final API contract above.

*(Frontend report views are the parallel frontend agent's, same split as M6 Phase 4 —
the backend contract above is what it needs.)*

---

# MILESTONE 8 — Data export

Covers `v1.md §14`, principle P5. Depends on every entity. **Q3 resolved
2026-09-04** (product owner, `AskUserQuestion`): a single JSON document.

## M8 decisions

| Topic | Default | Rationale / source |
|---|---|---|
| Format | One `GET /api/export` response, `Content-Type: application/json`, `Content-Disposition: attachment` — a single downloaded `.json` file, one JSON object with a named array per entity type. | Q3 resolution. |
| Module | New `internal/export` — read-only composition over `categories`, `timeline`, `tasks`, `habits`, `goals`, `reviews`, exactly like `internal/reports` (MX1/M7 precedent). Owns no table, performs no writes, no sqlc block. | ADR-0002; matches `internal/reports`. |
| Scope | Exactly the entities `v1.md §14` names: categories, planned blocks, actual blocks, tasks, habits + completions, goals, daily reviews, weekly reviews. **Not** `task_transitions` (audit trail, not a named export entity) or session/account credential data. | `v1.md §14`; principle P5 (user owns their data, not our internals). |
| Missing "list everything" methods | `tasks.Board` and `goals.ListGoals` already return every row unbounded — reused as-is. `categories.List`, `habits.ListActive`, `reviews.Get*` are all either archived-filtered or single-key — none returns the full unbounded set export needs, so each of `categories`, `timeline`, `habits`, `reviews` gains one or two new read methods (Phase 1). | Reuse over duplication, but only where an existing method already means "everything". |
| Dependency shape | Consumer-defined interfaces (`CategoriesReader`, `TimelineReader`, `TasksReader`, `HabitsReader`, `GoalsReader`, `ReviewsReader`) naming exactly the methods export needs; every real service already satisfies them structurally. | Same pattern as `internal/reports`. |
| Documented schema | `docs/export-format.md` — the JSON shape, one section per array, field-by-field. | `v1.md §14` "documented format". |

## Phase 1 — "List everything" methods on categories / timeline / habits / reviews  ✅ COMPLETE

- [x] 1.1  `categories`: `ListAll(ctx, accountID) ([]Category, error)` — active + archived (new query `ListAllCategories`, no schema change)
- [x] 1.2  `timeline`: `ListAllBlocks(ctx, accountID) ([]Block, error)` — every planned + actual block, unbounded (new query `ListAllBlocks`)
- [x] 1.3  `habits`: `ListAll(ctx, accountID) ([]Habit, error)` — active + archived, raw records (not the computed `HabitView`); `HabitCompletion{HabitID, Date}` + `AllCompletions(ctx, accountID) ([]HabitCompletion, error)` — every completion, every habit
- [x] 1.4  `reviews`: `ListDaily(ctx, accountID) ([]DailyReview, error)` / `ListWeekly(ctx, accountID) ([]WeeklyReview, error)` — every saved review of each kind
- [x] 1.5  Tests (4 new cases, one per module) — each new method returns everything incl. archived where relevant, and is isolation-tested against a second account
- ✅ **CP 1** each new method returns the account's complete set for its entity, nothing more, nothing of another account's — all 4 new tests green, full suite (`go test ./...`) green

## Phase 2 — `internal/export` module  ✅ COMPLETE

- [x] 2.1  `internal/export` package: `export.go` (`Service`, `Export` struct reusing every producing module's own record types, 6 consumer interfaces), `service.go` (pure composition — 8 read calls across 6 modules, no SQL, no connection pool)
- [x] 2.2  `Export(ctx, accountID) (Export, error)` — gathers all six sources; splits `timeline.ListAllBlocks` into `PlannedBlocks`/`ActualBlocks` by `Kind`; flattens `tasks.Board`'s four columns into one list; stamps `ExportedAt`
- [x] 2.3  Tests (`internal/export/service_test.go`, 2 cases over a shared fixture) — round-trip completeness (every entity incl. an archived category, an archived habit — whose completion still appears — and both review kinds, each asserted present with correct content); isolation (a second account's export contains only its own one category, nothing else)
- ✅ **CP 2** an export gathered for a fixture account contains exactly the fixture's data — every entity type, nothing missing, nothing foreign — both tests green

## Phase 3 — Export HTTP API + M8 wrap  ✅ COMPLETE

- [x] 3.1  `GET /api/export` (`read`-only — no write exists) → the full JSON document, `Content-Disposition: attachment; filename="productivity-os-export-<date>.json"`
- [x] 3.2  `docs/export-format.md` — documented schema, one section per array, field-by-field
- [x] 3.3  HTTP tests (`internal/export/http_test.go`, 2 cases) — round-trip completeness via the endpoint (every array populated, correct content, headers present); isolation
- [x] 3.4  Full regression — `go build/vet` clean, `golangci-lint` (0 issues), `sqlc diff` (clean — no schema change), `gofmt -l` (clean), `go test ./...` all green
- [x] 3.5  `docs/security-review-m8.md` — isolation review of the export endpoint and its six-module composition. **No findings.**
- ✅ **CP 3** verified live via curl (fresh server, real DB): seeded one of every entity (incl. an archived habit + its completion, both review kinds) for one account; downloaded `/api/export` — every array present with correct content and the right `Content-Disposition` header; a brand-new second account's export returned every array empty. Full suite + review green.

### M8 done when

CP 1–3 met · a single `GET /api/export` download contains every entity `v1.md §14`
names for the caller's account, correctly and completely, and nothing from any other
account · `docs/export-format.md` documents the shape · isolation tests green ·
`internal/export` performs no writes and owns no table.

**M8 backend is done as of 2026-09-04.** `v1.md` Q3 fully resolved (single JSON
document, product owner). No export UI exists yet — same split as M6/M7: the
parallel frontend agent owns the download button, against this final API contract.

*(No export UI exists yet — same split as M6/M7: the parallel frontend agent owns
the download button, against this final API contract.)*

---

# V1 DONE

- [x] Every capability §1–§14 demonstrably works **(backend)** — M1–M8 all complete (2026-09-04); curl-verified + full test suites green for every milestone. **Frontend is the parallel agent's** — not all 14 capabilities have a UI yet (Reviews/Reports/Export ship no UI in this repo's history; earlier milestones' UI is being rebuilt by that agent against the design mockups).
- [x] Q1–Q11 all resolved and reflected in `v1.md` — Q10 (M7) and Q3 (M8) were the last two, resolved 2026-09-04.
- [ ] N1–N7 verified: one binary; runs on free tiers; isolation tested (✅ — every module has isolation tests, all green); timezone / DST / midnight covered (✅ — extensively tested, esp. M2/M6/M7); responsive at 375 + 1280 (frontend agent's to verify — no browser in this environment); a database restore performed successfully once (not attempted — needs a real deployment target)
- [ ] Deployment spike done → **ADR-0008** (hosting provider, database provider, backup mechanism) — **not started**; needs a product-owner decision on provider(s), this environment cannot spike a live deployment
- [ ] No V1 non-goal present in the product — holds for the backend as built; depends on what the frontend agent ships against the design mockups (see Appendix A.0: those mockups are a V2/V3 product)

## What actually remains to call V1 fully done

The backend is finished. What's left is **not backend coding work**:

1. **The frontend catching up** — Reviews, Reports, and Export have zero UI yet (M6–M8 Phase 4 in each milestone above); earlier milestones' UI is mid-rebuild by the parallel agent.
2. **A deployment decision** (ADR-0008) — hosting + DB provider + backup mechanism. Needs the product owner; this environment can't spike a real deploy.
3. **Browser-dependent verification** — responsive breakpoints, a real DB restore — needs a target environment and a human (or a browser tool) this session doesn't have.
4. **A non-goal audit once the frontend is closer to finished** — confirm nothing in `v1.md`'s non-goals list (OAuth, notifications, collaboration, gamification, etc.) has crept in from the design mockups, which visibly include several of them (Appendix A.0).

None of that is "implement the next thing" work for me right now — it's waiting on the frontend agent, a deployment decision, or a browser. **Two things *are* mine to plan and build next — see immediately below (higher priority) and Appendix A (lower priority, gated on approval).**

---
---

# NEXT — Contract reconciliation (found 2026-09-04, supersedes the "V1 DONE" backend claim above in one respect)

While writing this plan, `docs/left.md` and `planning.md` Appendix C (both written by the
parallel frontend agent, same day) surfaced something the M6–M8 work above didn't know
about: **the frontend was built screen-by-screen against its own mocked contracts**,
with an explicit swap point noted for each. Some of what I built this session matches;
some doesn't. Until this is reconciled, Reports, Daily Review, Weekly Review, and the
Habits "This Month"/"This Week" views **run on mock data despite the real backend
existing** — so "every capability §1–§14 demonstrably works" (V1 DONE, item 1) is not
quite true yet for those four screens.

This is not a product-scope question (nothing here is new scope, everything is already
approved V1) — it's an integration-contract question, squarely mine to resolve per
Appendix A.4's own rule ("coordinate the API contract, not the components").

## Reconciliation table

| Screen | docs/left.md expects | What I built (M6/M7) | Verdict |
|---|---|---|---|
| **Reports** | **One** `GET /api/reports?from=&to=` returning all 5 sub-reports in one payload, field names `seconds`/`planned_seconds`/`actual_seconds` (no `difference`), `completed_days`+`range_days` (not `completion_rate`), bare `task_throughput` int (not an object) | **Five** separate routes (`/api/reports/time-by-category` etc.), different field names, `completion_rate` float, `task_throughput` as `{from,to,done_count}` | **Reconcile — rebuild the HTTP layer as one combined endpoint matching `docs/left.md`'s shape exactly.** `reports.Service`'s 5 Go methods stay (good internal shape); only `internal/reports/http.go` changes. Nothing depends on the 5-route shape yet — safe to replace, not "widen". |
| **Daily review** | `GET` returns **404/null** if nothing saved; keys `wentWell`/`notPlanned`/`differently`/`grateful` (frontend explicitly says it will remap these itself) | 200 with a blank `answers: {}` if nothing saved; keys `went_well`/`not_as_planned`/`differently_tomorrow`/`grateful_for` | **Mostly fine as-is** — key names are the frontend's problem to remap (their own doc says so). The 404-vs-200 choice was deliberate (a review's non-existence isn't an error state) and is *more* correct than what the mock assumed; flag it to the frontend agent rather than change it — 200-with-blank is strictly easier for a frontend to handle than 404-then-catch. |
| **Weekly review reference data** | Category time / habit counts / tasks→DONE that week, via *some* endpoint | Already exists: `GET /api/comparison?from=&to=`, `GET /api/habits/range?from=&to=`, `GET /api/tasks/throughput?from=&to=` (M6 Phase 1) | **No backend gap — a wiring gap.** `planning.md`'s own M6 Phase 4 note already says this; the frontend's Weekly Review screen (Appendix C.3) still uses a mock. Nothing for me to build; worth restating clearly so it gets picked up. |
| **Habits "This Month" heatmap** | `GET /api/habits/history?from=&to=` → per-habit **array of completion dates**, active-only (or archived flagged), range bound ~92 days | Nothing — `habits.CompletionCountsInRange` (M6) returns a **count**, not the date list a heatmap needs | **Real gap — new endpoint needed.** |
| **Habits "This Week" grid** | `GET /api/habits/week?date=` — one call instead of 7× `GET /api/habits?date=` | ✅ **Built 2026-09-05** — see R2.3 below, no longer deferred. |
| **Category `PATCH` wipes colour/icon on a rename-only call** | (not in docs/left.md; found via re-reading my own MX1 design) | `Update` requires the full `{name,colour,icon}` triple every call — a rename-only PATCH silently clears colour/icon | **Real gap — my own design choice, worth fixing now before any screen depends on the current behaviour.** Change `Update` to a partial update (pointer fields; omitted = unchanged, explicit `""` = cleared) — more correct PATCH semantics anyway. |
| **Habit rename + archived-categories/unarchive** | Both explicitly "(c) not V1, confirm before building" | Neither exists | **Habit rename is now needed anyway** — MX3's approved target-descriptor needs *some* edit endpoint; build `PATCH /api/habits/{id}` (name + target, mirroring the existing category `Update`) as part of MX3, which supersedes this "not needed" note. Archived-categories/unarchive stays unbuilt — no approval yet, no MX3 need. |

## Phase plan

### Phase R1 — Reports: one combined endpoint  ✅ COMPLETE

- [x] R1.1 New `GET /api/reports?from=&to=` in `internal/reports/http.go`, calling all 5 existing `Service` methods (via a new `Service.Report` composition method) and assembling one response in `docs/left.md`'s exact shape
- [x] R1.2 `range_days` per habit via new `activeDaysInRange` (habit's `CreatedAt`/`ArchivedAt`, now exposed on `habits.Habit`; `reports.HabitsReader` widened with `ListAll`)
- [x] R1.3 Removed the 5 granular routes — `reports.Service`'s 5 Go methods stay (used internally by `Report()` and directly tested), only the HTTP surface consolidated
- [x] R1.4 Tests (`internal/reports/service_test.go` + `http_test.go`, 15 cases total) — combined shape field-for-field; `range_days` bounded by creation and by archiving (2 new cases); 366-day bound (new); DST-range case retained; isolation retained
- ✅ **CP R1** `GET /api/reports?from=2026-08-01&to=2026-09-04` verified live via curl — every field name and value matches `docs/left.md`'s expected shape exactly (`time_by_category[].seconds`, `planned_vs_actual` excludes Uncategorized, `habit_completion[].completed_days`+`range_days`, bare `task_throughput`, `daily_actual_totals[].seconds`). Full suite + lint + `sqlc diff` green.

### Phase R2 — Habit history + week

- [x] R2.1 `habits`: `History(ctx, accountID, from, to Date) ([]HabitHistoryEntry{HabitID, Name, Archived bool, Completions []Date}, error)` — one query (`HabitHistory`, a `LEFT JOIN` so a zero-completion habit still appears), ≤ 92-day bound enforced server-side (400 above it, `maxHistoryRangeDays`)
- [x] R2.2 `GET /api/habits/history?from=&to=` — response includes an `archived` flag per habit (docs/left.md left this open; chose to include archived habits with the flag rather than omit them, since a past range's heatmap should stay complete)
- [x] R2.3 ✅ **COMPLETE 2026-09-05** (was deferred as low-priority, built after a fresh `docs/left.md` re-read surfaced it as still open). `habits.WeekView{WeekStart, Days, Habits []HabitWeekEntry{HabitID,Name,CurrentStreak,Completed}, Archived []ArchivedHabitName{HabitID,Name}}`; `Service.Week(ctx, accountID, date)` resolves the ISO week containing `date` via `timezone.ISOWeekWindow`, reuses `ListActiveHabits`/`ListCompletionDatesSince` (same per-habit loop and `currentStreak` as `ListActive`, so the streak's 400-day lookback is unaffected by the 7-day display window) plus `ListArchivedHabits` for the name-only list. `GET /api/habits/week?date=<any-day-in-week>` — no migration, no new sqlc query, reuses everything `ListActive`/`History` already had.
- [x] R2.4 Tests (`internal/habits/service_test.go` + `http_test.go`, 6 new cases for history + 3 new for week: `TestWeek`, `TestWeek_StreakLookbackExceedsWeekWindow`, `TestHabitEndpoint_Week`) — completion dates correct within range, excludes out-of-range dates; zero-completion habit still listed; archived habit + flag included; 92-day bound enforced (and exactly-92 allowed); `to`-before-`from`; isolation; week batching matches `docs/left.md`'s exact shape; streak lookback exceeds the displayed week; missing/malformed `date` → 400
- ✅ **CP R2** verified live via curl — `GET /api/habits/history?from=2026-08-01&to=2026-09-04` returns exactly `docs/left.md`'s expected shape (per-habit `completions` date array); a > 92-day range → 400. This is `mockHabitHistory`'s (`habitData.ts`) real swap-in target. `GET /api/habits/week?date=2026-09-05` returned `week_start: "2026-08-31"`, all 7 `days`, and a marked habit's `current_streak`/`completed` correctly — matching `docs/left.md`'s shape field-for-field. Full suite + lint + `sqlc diff` green; addendum in `docs/security-review-m4.md` — no findings.

### Phase R3 — Category partial update

- [x] R3.1 New `categories.UpdateInput{Name, Colour, Icon *string}` — nil = unchanged, non-nil (incl. `&""`) = set to that value. `Update` now does a read-merge-write (`GetActiveCategory` then the existing full-replace `UpdateCategory` with merged values) rather than a new partial SQL statement — simplest correct option at this scale. New `categoryUpdateRequest` (pointer fields) replaces `categoryRequest` for the PATCH route only; `Create`/`categoryRequest` unchanged (full value semantics, as before).
- [x] R3.2 Updated existing MX1 tests for the new signature (`in(...)` → `updateName(...)` at 5 call sites); added `TestUpdate_PartialLeavesOmittedFieldsUnchanged` (service) and `TestCategoryEndpoints_PartialUpdate` (HTTP) — rename-only, colour-only, explicit-empty-clears, and a no-op empty `UpdateInput{}`
- ✅ **CP R3** verified live via curl — created a category with colour+icon, PATCH'd `{"name":"..."}` only, reloaded: colour and icon both survived unchanged. Full suite + lint + `sqlc diff` green.

### Phase R4 — Habit edit (folds into MX3, see below)

Covered by the MX3 section — `PATCH /api/habits/{id}` for name + target descriptor.

**Suggested order:** R1 and R2 unblock the most already-built frontend work (Reports
screen entirely, two Habits views) — do those first. R3 is a small, isolated fix. R4 is
part of MX3.

---
---

# APPENDIX A — Design-driven scope expansion (backend analysis)

> **Status: ANALYSIS ONLY — not approved, not started.** Derived from the 13 mockups in
> `docs/design/references/` (`overall`, `dashboard`, `tasks`, `habits`, `goals`, `notes`,
> `categories`, `analytics`, `calander`, `timeline`, `timeline-week`, `timeline-month`,
> `timeline-agenda`). The frontend is being built by a separate agent against these
> screens; **this appendix is the backend side only** — what the API must provide.

## A.0 — Verdict

The mockups depict a **V2/V3 product**, not V1. Most of what they add is already on the
**V2/V3 roadmap** (`docs/roadmap.md`: Calendar, Focus Sessions, Pomodoro, Quick Capture,
Notifications, Achievements, richer analytics, AI insights) — the design just makes that
vision concrete and pulls it forward.

**Nothing in this appendix may be built until the product owner:**
1. approves the specific scope items below, and
2. the corresponding change is written into `docs/requirements/v1.md` (or a new
   `v2.md`) and, where architectural, a new ADR.

Every item is tagged: **`BUILT`** (V1 done) · **`V1-PLANNED`** (M6–M8 above) ·
**`NEW`** (beyond approved scope — needs the two steps above) · **`OUT`** (explicit V1
non-goal — needs a deliberate reversal).

## A.1 — Screen → backend requirements

### Dashboard (`dashboard.png`) — `NEW` (roadmap has no "dashboard")
A single "today" summary across every module.
- `NEW` `GET /api/dashboard?date=` → `{ tasks:{done,total}, habits:{done,total}, focus:{tracked_minutes, goal_minutes}, goals:{on_track,total}, schedule:[…today's blocks…], recent_tasks, habit_week_grid, goals_progress, notes_recent, week_task_counts }`. Pure composition of other modules' read methods — **no new storage**, but needs everything else below to exist first.
- `NEW` "Focus Time 4h 20m **of 8h goal**" → a per-account **daily focus-minutes goal** (settings) + focus-session tracking (A.7).
- `NEW` "Today's Schedule" blocks show a **completion checkmark** → time blocks need a `done` flag (A.4).

### Tasks (`tasks.png`) — extends `internal/tasks`
- `BUILT` create / edit / delete / move across 4 states; `task_transitions`.
- `NEW` **priority** — `priority` enum `LOW|MEDIUM|HIGH` (nullable or default `MEDIUM`). *Conflicts `v1.md §7`: "no priorities". Needs amendment.*
- `NEW` **category** — `category_id` on `tasks` → `categories` (A.3). *Conflicts `v1.md §2`: "tasks carry no category in V1".*
- `NEW` **starred** — `starred bool` for the "Starred" tab.
- `NEW` **task time** — the mockup shows `09:00` next to tasks. Either a `due_at timestamptz` (time-of-day) replacing/augmenting `due_date`, or leave as a plain date. *Conflicts `v1.md §7`: "The due date is a plain date with no time".*
- `NEW` derived counts: `GET /api/tasks/stats?from=&to=` → `{by_state, by_priority, by_category, overdue, due_this_week}` (overdue = `due_date < today` and not `DONE`).
- `OUT` **assignees / avatars** on tasks → collaboration is a **permanent V1 non-goal** (Vision). Skip entirely.
- `V1-PLANNED` (M7) task throughput / "entered DONE in range".

### Habits (`habits.png`) — extends `internal/habits`
- `BUILT` create / archive / unarchive / mark / unmark / current streak / last-30.
- `NEW` **target descriptor** — `target text` ("30 minutes", "8 glasses", "At least 20 pages"). Display-only; the completion stays binary. *Soft-conflicts `v1.md §9`: "A completion has no quantity, partial value, or note" — this is on the habit, not the completion, so arguably fine; confirm.*
- `NEW` **icon** — `icon text` (an icon key).
- `NEW` **category** — `category_id` on `habits` (A.3). *Conflicts `v1.md §2`.*
- `NEW` **longest streak ever** — computed over full history (extend the streak query window or add a `max_streak` cache).
- `NEW` **weekly / monthly consistency %** and per-day completion counts over a range → `GET /api/habits/stats?from=&to=` (also feeds M6 weekly review + M7 reports + analytics heatmap).
- `NEW` "Your Streak" (an **aggregate** streak — days where ≥1 habit was completed, or "perfect days"). Define precisely, then compute.

### Goals (`goals.png`) — extends `internal/goals`
- `BUILT` create / edit / delete / four-state progress.
- `NEW` **task linkage + derived % progress** — `goal_id` on `tasks` (or a `goal_tasks` join); progress = `done_tasks / linked_tasks`. The mockup's "12 / 20 tasks", "70%". *Directly conflicts `v1.md §10`: "no percentage, no numeric target, no progress derived from tasks, habits, or time" and "A goal is not linked to any other entity in V1". Major amendment.*
- `NEW` **milestones** — a `goal_milestones` sub-table `{goal_id, title, target_date, done_at}`. *Conflicts `v1.md §10`: "No milestones, key results, or check-in history".*
- `NEW` **icon**, **category tags** (multiple — a `goal_categories` join). *Conflicts `v1.md §2`.*
- `NEW` **derived status** — `ON_TRACK | AT_RISK | AHEAD | COMPLETED | NOT_STARTED`, computed from `% progress` vs `time elapsed to target_date`. Distinct from the manual 4-state progress; decide whether both exist or the derived status replaces the manual one.
- `NEW` "67% success rate", "on track" counts → `GET /api/goals/stats`.

### Notes (`notes.png`) — **`NEW` module `internal/notes`** (roadmap: "Quick Capture" is V2; a full Notes feature is not on any roadmap)
- `NEW` migration `notes` — `{id, account_id, title, body text, icon, color, pinned bool, favorite bool, archived_at, deleted_at, created_at, updated_at}`.
- `NEW` `note_tags` — free-form multi-tags (`{note_id, tag}`), or a shared tags table.
- `NEW` `category_id` on notes (A.3).
- `NEW` CRUD + `POST .../pin`, `.../favorite`, `.../archive`, `.../trash` (soft delete), `.../restore`. Tabs: All / Pinned / Favorites / Archive / Trash.
- `NEW` body is markdown/rich text with checklists — backend stores text as-is; rendering is the frontend's job.
- `NEW` word / character count — derived, or computed client-side.
- `NEW` "Related Notes" — either derived from shared tags, or a manual `note_links` table. Decide.

### Categories (`categories.png`) — **architecture change: extract `internal/categories`**
- `BUILT` (as timeline-owned) name + archive + active list.
- `NEW` **promote categories to their own top-level module** that `timeline`, `tasks`, `habits`, `goals`, `notes`, `calendar` all depend on (through its published interface — ADR-0002). `internal/timeline` keeps only `time_blocks`. **This is an architecturally significant refactor → needs an ADR** (amend or supersede ADR-0002's "currently: none" module note; `time_blocks.category_id` FK now points at another module's table — decide: shared `categories` table owned by the `categories` module, every other module reads it via the interface, FKs allowed within the one DB).
- `NEW` `color` + `icon` columns. *Conflicts `v1.md §2`: "No colour or icon carries product meaning".*
- `NEW` `GET /api/categories` returns per-category **item counts** by type (`{tasks, notes, habits, goals, events}`) and a **`last_used_at`** (max of the referencing rows' timestamps, or a maintained column) for "Recently Used".
- `NEW` category `import` / `export`. Folds into M8 export or a small addition.

### Analytics (`analytics.png`) — **supersedes M7 Reports with a much larger surface**
- `V1-PLANNED` (M7) the five fixed reports (Time by category, Planned vs actual, Habit completion, Task throughput, Daily actual totals).
- `NEW` **date-range comparison** — "+12% vs last month" needs previous-period figures. *Conflicts `v1.md §13`: "no comparison between two ranges".*
- `NEW` **trend series** — tasks-completed + focus-time per day over a range (line/bar). *Conflicts `v1.md §13`: "no trend lines or forecasts".*
- `NEW` **habit-consistency heatmap** — per-day completion rate, range. (Feeds off A.1 habits stats.)
- `NEW` **productivity streak** — a derived "productive day" streak (define: a day with ≥ N tasks done, or ≥ M focus minutes, or ≥ 1 habit — pick a deterministic rule).
- `NEW` **deterministic "insights"** — e.g. "focus time peaked on Sep 17" (a fact, fine), "18% more productive this month" (needs a defined productivity score — riskier). *Brushes `v1.md §13`: "No unsupported AI-generated productivity claims" and the V3 "Productivity Insights" item. Keep strictly factual + explainable, or defer.*
- `NEW` per-tab report endpoints: `GET /api/analytics/{overview|productivity|habits|tasks|goals|time|categories}?from=&to=`.
- `NEW` "Export Report".

### Calendar / Events (`calander.png`, and the Week/Month/Agenda timeline views) — **`NEW` module `internal/calendar`** (roadmap: "Calendar" is V2)
- `NEW` migration `events` — `{id, account_id, title, starts_at, ends_at, all_day bool, category_id, icon, description, done_at, created_at}`. *Conflicts `v1.md §5`: "One date at a time; no week or month timeline."*
- `NEW` `GET /api/events?from=&to=&view=day|week|month|agenda` → events in range.
- `NEW` CRUD + a done toggle.
- **Design decision needed:** are "events" distinct from `time_blocks`, or do time blocks gain a title and become the calendar's content? The `timeline-*.png` views show titled, tagged, categorized, completable blocks in day/week/month/agenda — strongly implies **unifying `time_blocks` and events** into one "scheduled item" entity. This is the single biggest modelling decision in the whole expansion.
- `OUT` external calendar sync (Google) → V2 with its own spec; not now.

### Timeline (`timeline.png`, `-week`, `-month`, `-agenda`) — major extension of `internal/timeline`
- `BUILT` per-date planned/actual blocks + comparison.
- `NEW` blocks gain **title**, **icon**, **tags** (multi), **completion state** (`done_at`). *Conflicts `v1.md §3/§4`: a block has only start/end/category.*
- `NEW` **week / month / agenda range reads** — `GET /api/timeline?from=&to=&view=…`. *Conflicts `v1.md §5`: "no week or month timeline".*
- `NEW` per-view aggregate headers (tasks done, focus minutes, habits, goals for the week/month).
- `NEW` "current time" indicator — frontend only, no backend.

### Focus / Pomodoro (`dashboard.png`, `timeline.png` right rail) — **`NEW` module `internal/focus`** (roadmap: "Focus Sessions" + "Pomodoro" are V2)
- `NEW` migration `focus_sessions` — `{id, account_id, started_at, ended_at, planned_minutes, kind (FOCUS|SHORT_BREAK|LONG_BREAK), task_id null, block_id null, created_at}`.
- `NEW` `POST /api/focus/start`, `POST /api/focus/{id}/stop`, `GET /api/focus?from=&to=` (history + totals).
- `NEW` per-account **Pomodoro config** (work / short-break / long-break lengths, rounds) → settings (A.9).
- Feeds Dashboard "Focus Time", analytics "Focused Time", trends.

### Activity feed (`overall.png` #9 "Timeline — See your activity history", `dashboard` has none) — `NEW`
- The `overall.png` "Timeline" screen is an **activity log** ("Completed a task", "Added a new note", "Reached 7-day habit streak", "Created a new goal", "Updated category"). Distinct from the day-planner "Timeline".
- `NEW` migration `activity_events` — `{id, account_id, kind, subject_type, subject_id, summary, at}`, written by each module on notable actions. Or derive from existing `*_transitions` / `created_at` — cheaper but incomplete.
- `NEW` `GET /api/activity?before=&limit=`.
- **Naming clash:** the nav has both "Timeline" (day-planner) and, in `overall.png`, "Timeline" (activity history). Resolve the naming before building.

### Settings / Preferences (`overall.png` #10) — `NEW`
- `NEW` migration column or table `account_preferences` — `{theme (LIGHT|DARK|SYSTEM), accent_color, font_size, compact_mode bool, milestone_celebrations bool, daily_focus_goal_minutes, week_start, pomodoro_config jsonb}`.
- `NEW` `GET /api/preferences`, `PUT /api/preferences`.
- *Conflicts `v1.md §1`: "No profile fields beyond email, password, and timezone" — preferences aren't "profile" exactly, but it's an expansion of the account surface. Confirm.*

### Profile + Badges (`overall.png` #11) — `NEW` / `OUT`
- `NEW` profile fields — display name, role tags, bio/quote. *Conflicts `v1.md §1` directly.*
- `OUT` **badges / achievements** ("Consistency", "Goal Setter", "Early Bird") + "Show Milestone Celebrations" → **"Advanced gamification" is a V1 non-goal**; "Achievements" is V2 "carefully". Defer; if pulled forward, needs its own spec (roadmap V2 gate).

### Authentication (`overall.png` #12) — mostly `BUILT`, one `OUT`
- `BUILT` email + password sign up / log in.
- `OUT` **"Continue with Google / GitHub"** → `v1.md §1`: "no OAuth or social login". Needs a deliberate reversal + likely an ADR (adds an OAuth dependency, callback routes, provider config) + ADR-0008 implications (redirect URIs per environment).

### Onboarding (`overall.png` #13) — `NEW`, mostly frontend
- `NEW` a single `onboarded_at` timestamp on the account (so the flow shows once). Everything else is frontend.

### Global Search (`overall.png`, search bar on every screen) — `NEW`
- `NEW` `GET /api/search?q=` → matches across tasks / notes / goals / events (title + body). Postgres `ILIKE` / `pg_trgm` / `tsvector` — `pg_trgm` or a `tsvector` column is the boring choice; **full-text search via a `tsvector` GIN index stays within "PostgreSQL only" (N1) — no Elasticsearch.** Confirm the approach; a `tsvector` column per searchable table + a `UNION` query is simplest.

### Notifications / Reminders (`overall.png`, bell icon) — `OUT`
- `v1.md` non-goal ("Notifications, reminders, email digests"); roadmap V2 with its own delivery-infra spec. Not now.

### Subscription / "Free Plan" (`overall.png`, sidebar) — `OUT` (not on any roadmap)
- Defer entirely. If ever needed, its own milestone + payment-provider ADR.

### Trash / Archive, Import / Export (`overall.png`) — partly `V1-PLANNED`
- `V1-PLANNED` (M8) full-account export. `NEW`: soft-delete (`deleted_at`) + a Trash view for notes (and maybe tasks/events); per-entity restore. `NEW`: import (v1.md export is one-way — "no import in V1").

## A.2 — v1.md / ADR amendments this expansion requires (the approval checklist)

| # | Change | Touches |
|---|--------|---------|
| 1 | Categories are cross-cutting (tasks, habits, goals, notes, events all carry a category) and carry a meaningful colour + icon | `v1.md §2` scope boundary; `v1.md` domain-concepts intro ("Tasks, habits, and goals carry no category in V1") |
| 2 | Categories become their own module; `time_blocks` keeps only blocks | **new ADR** (module split); `docs/architecture/overview.md` |
| 3 | Tasks gain priority ✅, category ✅ *(MX1)*, starred *(not approved)*, and (optionally) a time-of-day due *(not approved)* | `v1.md §7` scope boundary — **priority (HIGH/MEDIUM/LOW) approved 2026-09-05, product owner, re-asked after the 2026-09-04 answer wasn't captured; written into `v1.md §7`. Starred and time-of-day due asked in the same round, not approved — remain out of scope.** |
| 4 | Habits gain a target descriptor ✅, icon *(not approved)*, category ✅ *(MX1)* | `v1.md §9` scope boundary — **target descriptor approved 2026-09-04, product owner; written into `v1.md §9`** |
| 5 | Goals gain task linkage + derived % progress ✅, milestones *(not approved)*, icon *(not approved)*, category tags *(not approved — MX1 gave one `category_id`, still single)*, derived status *(not approved)* | `v1.md §10` (whole section + scope boundary) — **task linkage + derived progress approved 2026-09-04, product owner; written into `v1.md §10`** |
| 6 | New feature: **Notes** — title + free-text body only ✅. Tags, pin/favourite/archive/trash, categories on notes: *asked, not approved* | `v1.md §15` (new) — **approved 2026-09-05, product owner, narrowed to plain-text title+body (multi-select question; only that option was picked). ⚠ Conflicts with `docs/design/design-system.md` §6.4, which separately ratified Notes as "must not be implemented" — that ratification predates this approval and is stale relative to it (it also still forbids MX1/MX3 items already shipped); product owner explicitly chose to override it here. The frontend agent needs to sync §6.4 — not done by this agent (frontend files off-limits).** |
| 7 | New feature: **Calendar / Events** + week/month/agenda views; likely unify `time_blocks` + events | `v1.md §5` scope boundary; **new ADR** (unified scheduled-item model); roadmap (pull "Calendar" forward from V2) |
| 8 | Timeline blocks gain title / icon / tags / completion; timeline gains range views | `v1.md §3, §4, §5` |
| 9 | New feature: **Focus / Pomodoro sessions** + a daily focus goal | `v1.md` — new section; roadmap (pull forward from V2) |
| 10 | Analytics: range comparison, trend series, heatmap, productivity streak, factual insights | `v1.md §13` (scope boundary says the 5-report list is *exhaustive*); roadmap ("Richer Weekly Analysis" / "Productivity Insights" are V2/V3) |
| 11 | New: **Activity feed** | `v1.md` — new section |
| 12 | New: **account preferences** (theme, accent, font size, compact, celebrations, focus goal, pomodoro config) | `v1.md §1` |
| 13 | New: **profile fields** (display name, roles, bio) | `v1.md §1` scope boundary directly |
| 14 | New: **global search** via Postgres `tsvector` | `v1.md` — new capability; confirm it stays Postgres-only (N1) |
| 15 | New: **soft-delete / trash + restore**; **import** | `v1.md §14` (export is one-way, "no import in V1") |
| 16 | **OUT unless reversed:** OAuth login; assignees/collaboration; notifications/reminders; badges/gamification; subscription; external calendar sync | `v1.md §1` scope boundary; Vision (collaboration = permanent non-goal); `v1.md` non-goals list |

## A.3 — Proposed build order (once approved)

Sequenced by dependency. Each becomes a full milestone with phases (schema → service → API → tests → isolation review) like M2–M5. **Do not start any of these until A.2 items are approved and written into `v1.md`.**

1. **MX1 — Categories module extraction + colour/icon.** ✅ **DONE 2026-09-04.** ADR-0009. `internal/categories` extracted; `colour`/`icon` added; `category_id` on `tasks`/`habits`/`goals`; `GET /api/categories/overview` counts endpoint. Full suite + security review green.
2. **MX2 — Finish V1 first.** ✅ **DONE 2026-09-04.** M6 (Reviews), M7 (Reports), M8 (Export) all backend-complete — see "V1 DONE" above. The expansion now sits on top of a complete V1, as intended, not instead of it.
3. **MX3 — Task / Habit / Goal field additions.** ✅ **DONE 2026-09-04** (narrowed scope — see the MX3 milestone section below). Built: habit target descriptor + `PATCH /api/habits/{id}`; goal task-linkage (`tasks.goal_id`) + derived progress (`done_tasks`/`total_tasks` on `GET /api/goals`). Full suite + security review green. Not yet approved and out of scope: anything on tasks (re-ask), habit icon, goal milestones/icon/multi-category/derived-status.
4. **MX4 — Notes module.** ✅ **DONE 2026-09-05** (narrowed scope — see the MX4 milestone section below). New `internal/notes`; title + free-text body CRUD only. No tags, no pin/favourite/archive/trash, no category. `v1.md §15` written. Full suite + security review green.
5. **MX5 — split into two, on investigation 2026-09-05 (see the MX5 milestone section below):**
   - **MX5a — range-read endpoint.** ✅ **DONE 2026-09-05.** `GET /api/timeline/range` — no new scope, no `v1.md` change (§5's Week/Month amendment already covers "same block data, different rendering"); purely an already-authorized performance fix the frontend's Week/Month views were explicitly waiting on (`docs/left.md`).
   - **MX5b — Calendar/Events unification.** Still **not started, no live requirement**. On investigation, nothing currently requires a distinct "Event" entity — `design-system.md §6.4` still correctly excludes a standalone Calendar screen (unlike Notes, this exclusion was never contradicted by an approval), and Week/Month already ship on the existing block model alone. **Deliberately parked** rather than speculatively built — revisit only if/when an actual requirement for events (distinct from time blocks) appears.
6. **MX6 — Focus / Pomodoro module.** **Investigated 2026-09-05, left as-is (product owner).** The lightweight version — a standalone, non-persistent focus timer — was separately approved 2026-09-05 (`v1.md §4`, design-system G2) and is already fully built frontend-only (`PomodoroCard.tsx`), explicitly with **no backend surface** ("the timer does not read or write any account data"). This full A.3 item — a persistent `internal/focus` module with a `focus_sessions` table, session history, and Pomodoro config in preferences — remains **unapproved and unrequested**; nothing currently needs it. Deliberately left unbuilt rather than speculatively scoped, mirroring MX5b's parking of Calendar/Events. Revisit only if a real requirement for persisted focus history appears.
7. **MX7 — Preferences + profile.** **Investigated 2026-09-05, left parked (product owner).** No live requirement pulls any of it forward: theme preference is already fully built frontend-only (`web/src/theme.ts`, `localStorage`, zero backend), week-start is already fixed globally (Monday/ISO, not a per-user setting), and `Account.tsx` explicitly documents its current scope boundary as intentional ("No profile fields beyond email/password/timezone (§1 boundary)") with nothing built or blocked waiting on accent colour/font size/compact mode/celebrations/display name/bio/role tags. Unlike MX5a, there is no blocked frontend consumer. Deliberately parked, mirroring MX5b/MX6 — revisit only if a real requirement for a specific field appears.
8. **MX8 — Analytics platform.** Range comparison, trends, heatmaps, productivity streak, factual insights, per-tab endpoints. Supersedes/absorbs M7. Amend `v1.md §13`.
9. **MX9 — Activity feed.** `activity_events` written by each module; `GET /api/activity`.
10. **MX10 — Global search.** `tsvector` columns + GIN + a `UNION` search endpoint.
11. **MX11 — Dashboard aggregate endpoint.** Pure composition — build last, when every module it reads exists.
12. **Deferred (own specs / ADRs, not scheduled):** OAuth, notifications/reminders, badges/gamification, external calendar sync, subscription, import. Each is a V1 non-goal or a V2-gated feature.

## Approval status snapshot (updated 2026-09-05)

Per A.0's own rule (unchanged, still binding): nothing may be built until the product
owner approves the specific A.2 scope item and it is written into `v1.md`. Approved so
far, all written into `v1.md`:

- 2026-09-04: habit target descriptor (A.2 row 4); goal task-linkage + derived
  progress (A.2 row 5). → **MX3, done.**
- 2026-09-05: task priority, `HIGH`/`MEDIUM`/`LOW` only (A.2 row 3) — re-asked after
  the 2026-09-04 answer wasn't captured; starred and time-of-day due asked in the same
  round and **not** approved. → **MX3-follow-up, done.**
- 2026-09-05: Notes, narrowed to title + free-text body only (A.2 row 6) — tags,
  pin/favourite/archive/trash, and categories-on-notes were offered in the same
  multi-select question and **not** picked. → **MX4, done.** ⚠ This approval
  overrides `docs/design/design-system.md` §6.4, which separately ratified Notes as
  "must not be implemented" (a ratification that predates this approval and is already
  stale re: MX1/MX3's shipped items — categories-on-tasks/habits/goals and
  goal-linked-tasks are both on that same §6.4 forbidden list). **The frontend agent
  still needs to sync §6.4** — this agent does not edit frontend files.

Still unapproved: habit icon; goal milestones/icon/multi-category/derived-status; task
starred flag and time-of-day due. Everything in A.2 rows 7–16 (MX5–MX11's scope)
remains unapproved and ungated.

Still true regardless of what's approved:

- **The frontend is already being built against the full mockups** — which per A.0's
  own verdict depict the *entire* MX3–MX11 surface. The longer full approval waits, the
  more frontend work exists for scope not yet backed by an approved requirement.
- **A.3's sequencing is a full re-plan, not a queue to pop one item off.** MX5 (the
  scheduled-item / calendar unification) is flagged as "the single biggest modelling
  decision in the whole expansion" — deciding it changes what later milestones should
  build against.

## A.4 — Frontend note

The frontend rebuild (separate agent) has restructured `web/src/` (`features/`,
`components/ui/`, `docs/design/`). **This appendix does not touch the frontend.** The
Goals page and design-system CSS added during M5 may be superseded by that agent's work —
that is expected and fine; coordinate the API contract, not the components.

---

# MX1 — Categories as a shared module

**Approved 2026-09-04.** `v1.md §2` amended; **ADR-0009** written. First expansion
milestone. Extract `categories` from `internal/timeline` into its own module, give it
colour + icon, and let `tasks` / `habits` / `goals` carry a `category_id`.

Sequence after MX1: **M6 → M7 → M8** (finish approved V1), then MX3+.

## Phase 1 — `internal/categories` module + colour/icon  ✅ COMPLETE

- [x] 1.1  Migration `000007_categories_extend` — `ALTER TABLE categories ADD COLUMN colour text, ADD COLUMN icon text` (both nullable; a key, not a hex value). Applied to dev + test DBs.
- [x] 1.2  New `internal/categories` package: `categories.go` (`Service` interface + `Category`, `Input`, `ErrNameTaken`, `ErrNotFound`, `ValidationError`), `service.go`, `http.go`, `queries.sql` → `categoriesdb` (new sqlc block, 2nd in `sqlc.yaml`). Category SQL removed from `internal/timeline/queries.sql`.
- [x] 1.3  `Service`: `Create(accountID, Input{name,colour,icon})`, `Update(accountID, id, Input)` (full set of name+colour+icon), `Archive`, `List`, `AssignableToAccount` (exists + owned + active), `NamesForAccount` (all incl. archived — for read-model labels)
- [x] 1.4  Validation: name 1–60 trimmed + case-insensitive active-unique; colour/icon trimmed, ≤ 40, empty allowed (empty ⇒ NULL)
- [x] 1.5  Tests — `internal/categories/service_test.go` (ported + colour/icon round-trip + `AssignableToAccount` matrix + `NamesForAccount` + isolation), `http_test.go` (endpoint walk-through). 13 tests green.
- ✅ **CP 1** `internal/categories` tests green; `internal/timeline` no longer references the `categories` table (queries + JOIN removed; names resolved via `CategoryStore`)

## Phase 2 — Re-wire `timeline`; move the HTTP routes  ✅ COMPLETE

- [x] 2.1  `internal/timeline` — `CountAssignableCategory` query + inline check removed; `CategoryStore` interface (`AssignableToAccount` + `NamesForAccount`); `assertAssignableCategory` calls it. `NewService(pool, zone, cats)`.
- [x] 2.2  `Create/Rename/Archive/ListActiveCategories` + `Category` type + category sentinels removed from `timeline`; category routes removed from `timeline.Handler.Mount`; `ListBlocksOverlapping` JOIN dropped (read models call `NamesForAccount`)
- [x] 2.3  `internal/categories/http.go` — `GET/POST /api/categories`, `PATCH /api/categories/{id}` (name+colour+icon), `POST /api/categories/{id}/archive`. Body carries `colour`, `icon`. Same `Protector` pattern. *(PATCH still 204 — contract unchanged for the parallel frontend agent.)*
- [x] 2.4  `cmd/server` — `categorySvc := categories.NewService(pool)`; handler mounted; passed into `timeline.NewService`
- [x] 2.5  `internal/timeline` tests: shared helpers moved to `helper_test.go`; category setup via direct SQL (`mkCategory`/`archiveCategory` — CRUD is exercised in the categories pkg); `TestCategoryEndpoints` moved to categories
- ✅ **CP 2** `GET /api/categories` returns colour/icon; time blocks still reject a foreign/archived/unknown category (`http_test` + `block_test`); full `go test ./...`, `go vet`, `golangci-lint`, `sqlc diff` all clean

## Phase 3 — `category_id` on tasks / habits / goals  ✅ COMPLETE

- [x] 3.1  Migration `000008_entity_category` — `ALTER TABLE tasks|habits|goals ADD COLUMN category_id uuid REFERENCES categories(id) ON DELETE RESTRICT` (nullable); indexed `(account_id, category_id)` on each. Applied to dev + test DBs.
- [x] 3.2  `tasks` — `TaskInput`/`Task` gain `CategoryID *uuid.UUID`; a local `CategoryChecker` interface; `CreateTask`/`UpdateTask` validate via `assertAssignableCategory`; `POST`/`PATCH /api/tasks` accept + return `category_id`
- [x] 3.3  `habits` — `HabitInput{Name, CategoryID}` (`CreateHabit` now takes it); `Habit`/`HabitView` gain `CategoryID`; new `SetHabitCategory` + `PUT /api/habits/{id}/category` (nil clears). Create/update both validated. *(No general habit edit yet — MX3, per the note.)*
- [x] 3.4  `goals` — `GoalInput`/`Goal` gain `CategoryID *uuid.UUID`; `CreateGoal`/`UpdateGoal` validate + persist; API accepts + returns `category_id`
- [x] 3.5  Tests — `TestTaskCategory`, `TestHabitCategory`, `TestGoalCategory` (+ HTTP-level category tests in each package): set at create, changed, cleared (`null`), foreign/archived/unknown → 400 `category_id`; isolation covered via the existing per-module isolation tests plus explicit foreign-category cases
- ✅ **CP 3** verified live via curl (fresh server, real DB): category w/ colour+icon created; task/habit/goal each created with it and returned it; unknown category → 400 `VALIDATION_ERROR` on all three; habit's dedicated `PUT .../category` clears it (`category_id: null` on reload). `go build`, `go vet`, `golangci-lint` (0 issues), `sqlc diff` (clean), `go test ./...` all green.

## Phase 4 — Category overview (counts) + MX1 wrap  ✅ COMPLETE

- [x] 4.1  `categories.Counter` interface (`CountByCategory(ctx, accountID) (map[uuid.UUID]int, error)`) declared in `internal/categories`; implemented by `tasks`, `habits` (active only — archived habits don't inflate a category's count), `goals`, `timeline` (blocks)
- [x] 4.2  `GET /api/categories/overview` — `cmd/server/overview.go`, a composition handler over `categories.List` + the four `Counter`s; response `{"categories":[{ id, name, colour, icon, counts:{tasks,habits,goals,blocks} }]}`; mounted behind `read`
- [x] 4.3  Full regression — `go build`, `go vet`, `golangci-lint` (0 issues), `sqlc diff` (clean), `gofmt -l` (clean), `go test ./...` (all packages, incl. new `TestCountByCategory` in all four modules + `cmd/server`'s first test, `TestCategoriesOverview`)
- [x] 4.4  `docs/security-review-mx1.md` — isolation review of the extracted module, the four new FKs, `CountByCategory`, and the composition endpoint. **No findings.**
- [x] 4.5  `docs/architecture/overview.md` "Modules" section — done in Phase 1/2 (module table + `categories` dependency direction); still accurate, no further change needed
- ✅ **CP 4** verified live via curl (fresh server, real DB): category with colour/icon + an empty category; 2 tasks, 1 habit, 1 goal, 1 block attached to one category; `GET /api/categories/overview` returned exactly `{tasks:2, habits:1, goals:1, blocks:1}` for it and all-zero counts for the empty one. Also proven by `TestCategoriesOverview` (`cmd/server`).

### MX1 done when

CP 1–4 met · `internal/categories` owns the table · `timeline` depends on it only via
`CategoryStore` · tasks/habits/goals carry an optional category · isolation tests
green across all five modules · `sqlc diff` clean.

**MX1 is done.** M6 → M7 → M8 are done (2026-09-04). Next: the R1–R4 contract
reconciliation above, then MX3 below.

---

# MX3 — Task / Habit / Goal field additions (narrowed to what's approved)

**Partially approved 2026-09-04**, product owner. Only the two items below are
approved and written into `v1.md`; everything else A.2 rows 3–5 proposed (task
priority/starred/time-of-day-due, habit icon, goal milestones/icon/multi-category/
derived-status) is **not** approved and not in this milestone. Re-ask before adding
anything to `internal/tasks`.

## MX3 decisions

| Topic | Default | Rationale / source |
|---|---|---|
| Habit target descriptor | `target text`, nullable, free text, display-only — never validated against a completion. | `v1.md §9` amendment, 2026-09-04. |
| Habit edit | No general habit edit exists yet (create-only + archive). Add `PATCH /api/habits/{id}` for **name + target** together — the natural single edit surface, and unifies with R4/`docs/left.md`'s "not V1, confirm before building" note, which this approval supersedes. | Avoids a third separate single-field endpoint next to `SetHabitCategory`. |
| Goal-task linkage | Nullable `goal_id` on `tasks` (many tasks per goal, one goal per task) — same shape as `category_id`, same `GoalChecker`-via-consumer-interface pattern as `CategoryChecker` (ADR-0009 precedent, no new ADR needed). | Mirrors the already-established cross-module-FK pattern; simplest shape matching the "12/20 tasks" mockup. |
| Derived progress | `goals.Service` gains `Progress(ctx, accountID, goalID) (done, total int, error)` computed from linked tasks' state; **not** stored, computed on read. Zero linked tasks → `0, 0` (frontend shows "0%" or "no tasks linked", its call). | `v1.md §10`: "the manually set progress state and the derived task-completion percentage are two independent figures." |
| Delete-with-links | Deleting a goal **clears `goal_id` on its linked tasks** (does not delete the tasks, does not block the delete). | Build-time default per `v1.md §10`'s amendment note — **flagged, not yet explicitly product-owner-confirmed**; cheap to change before any frontend depends on it. |

## Phase 1 — Habit target descriptor + edit endpoint ✅ COMPLETE (2026-09-04)

- [x] 1.1  Migration — `ALTER TABLE habits ADD COLUMN target text` (nullable) — `000010_habits_target`, applied to dev + test DBs
- [x] 1.2  `habits`: `Habit`/`HabitInput` gain `Target *string`; new `UpdateHabit(ctx, accountID, habitID uuid.UUID, name string, target *string) (Habit, error)`; `CreateHabit` accepts `target` too, validated via `validateTarget` (100-char bound, trims, empty→nil)
- [x] 1.3  `PATCH /api/habits/{id}` — body `{name, target}`, returns `200` + updated habit; `POST /api/habits` gains optional `target` in the request body
- [x] 1.4  Tests — `TestCreateHabit_WithTarget`, `TestUpdateHabit` (full edit, clearing target, bad-name validation, unknown-habit → 404, isolation, confirms streak/completions untouched), `TestHabitEndpoints_TargetAndEdit` (HTTP-level). 24 tests total in `internal/habits`, all green.
- ✅ **CP 1** verified live via curl against a running server: registered a test account, `POST /api/habits` with `{"name":"Workout","target":"30 minutes"}` → created with target echoed back; marked today's completion (204); `PATCH /api/habits/{id}` with `{"name":"Workout daily","target":"45 minutes"}` → both fields updated in the response; `GET /api/habits` confirmed `current_streak: 1`, `completed_on_date: true`, `last_30_days: 1` all survived the edit untouched. Full repo regression (`go build/vet/test`, 15 packages) green; background `gofmt`/`golangci-lint`/`sqlc diff` job clean (0 issues, sqlc diff OK).

## Phase 2 — Goal-task linkage + derived progress ✅ COMPLETE (2026-09-04)

- [x] 2.1  Migration — `000011_task_goal_link`: `ALTER TABLE tasks ADD COLUMN goal_id uuid REFERENCES goals(id) ON DELETE SET NULL` (nullable), indexed `(account_id, goal_id)`. Applied to dev + test DBs.
- [x] 2.2  `tasks.GoalChecker` (`AssignableToAccount`-shaped, defined in `tasks`, mirroring `CategoryChecker`) validates a caller-assigned `goal_id`; satisfied structurally by a new `goals.Service.AssignableToAccount` (`CountAssignableGoal` — `WHERE account_id = $1 AND id = $2`).
- [x] 2.3  `tasks`: `Task`/`TaskInput` gain `GoalID *uuid.UUID`; `CreateTask`/`UpdateTask` validate it via `assertAssignableGoal`; `POST`/`PATCH /api/tasks` accept/return `goal_id`.
- [x] 2.4  Progress, **revised from the original plan**: rather than a raw SQL join from `goals` into the `tasks` table (which would break this codebase's established one-directional module-boundary convention — see `CountByCategory`/`Counter` precedent), built as a consumer-owned composition: `goals.ProgressReader` interface (`ProgressByGoal(ctx, accountID) (done, total map[uuid.UUID]int, error)`), satisfied structurally by a new `tasks.Service.ProgressByGoal` (own table only, `WHERE account_id = $1 AND goal_id IS NOT NULL GROUP BY goal_id`). `goals.Handler` takes this as a second collaborator (mirroring `tasks.Handler`'s `AccountZone` pattern) and merges `done_tasks`/`total_tasks` onto each goal in the `GET /api/goals` list response at the HTTP layer — one query, no N+1, and the `goals` domain package/table stays untouched by `tasks`' schema.
- [x] 2.5  Tests — `TestTaskGoalLink` (link/unlink/change, foreign/unknown `goal_id` → 400 on create and update, isolation), `TestTaskGoalDeleteClearsLink` (deleting a goal clears `goal_id` on its tasks without deleting them), `TestGoalEndpoints_TaskProgress` (end-to-end: link 3 tasks, mark 2 done, `GET /api/goals` reads `2/3`, delete the goal, tasks survive with `goal_id: null`).
- ✅ **CP 2** verified live via curl: created a goal, linked 3 tasks to it, marked 2 `DONE` — `GET /api/goals` returned `"done_tasks": 2, "total_tasks": 3`; a task create with an unknown `goal_id` → `400 VALIDATION_ERROR {"goal_id":"goal not found"}`; deleted the goal (`204`) — `GET /api/board` confirmed all 3 tasks survive, each with `"goal_id": null`.

## Phase 3 — MX3 wrap ✅ COMPLETE (2026-09-04)

- [x] 3.1  Full regression — `go build/vet/test` (17 packages green), `golangci-lint` (0 issues), `sqlc diff` (clean), `gofmt -l` (clean)
- [x] 3.2  `docs/security-review-mx3.md` — isolation review of `habits.target`, `tasks.goal_id`, `goals.AssignableToAccount`, and the `ProgressByGoal`/`ProgressReader` composition. **No findings.**
- ✅ **CP 3** full suite + review green

### MX3 done when (narrowed scope) ✅ ALL MET (2026-09-04)

CP 1–3 met · a habit can carry and edit a target descriptor · a task can link to a goal
and a goal's derived progress reads correctly · deleting a goal never deletes its tasks ·
isolation tests green. **MX3 is complete and shipped.**

---

# MX3-follow-up — Task priority

**Approved 2026-09-05**, product owner. `v1.md §7` amended. The one item from A.2 row 3
that got approved on re-ask; starred and time-of-day due were asked in the same round
and not picked — still out of scope.

## Decisions

| Topic | Default | Rationale / source |
|---|---|---|
| Scale | `priority` is one of `HIGH` \| `MEDIUM` \| `LOW`, nullable (unset by default) | Matches the frontend mockups' existing "High/Medium/Low" chips exactly (`docs/design/screens/tasks.md`) — no new design decision needed. Product-owner confirmed 2026-09-05. |
| Behaviour | Purely a label — no effect on board ordering, no default sort, no "sort by priority" view | `v1.md §7` scope boundary: "no effect on ordering, board layout, or any other behaviour" — matches the existing "no manual reorder" precedent (Q5). |
| Storage shape | A single nullable text/enum column on `tasks`, same shape as `state` but validated against 3 fixed values instead of 4 | Mirrors `tasks.State`'s existing `CHECK` constraint pattern exactly — no new pattern needed. |

## Phase 1 — `tasks.priority` ✅ COMPLETE (2026-09-05)

- [x] 1.1  Migration `000012_task_priority` — `ALTER TABLE tasks ADD COLUMN priority text CHECK (priority IN ('HIGH','MEDIUM','LOW'))` (nullable, no default). Applied to dev + test DBs.
- [x] 1.2  `tasks`: `Task`/`TaskInput` gain `Priority *Priority` (new `Priority` type + `High`/`Medium`/`Low` consts, mirroring `State`); `validPriority` checked in `validateInput`, same shape as `CreateTask`/`UpdateTask`'s existing category/goal validation
- [x] 1.3  `POST`/`PATCH /api/tasks` accept/return `priority`; `GET /api/board` includes it on every task
- [x] 1.4  Tests — `TestTaskPriority` (create/update with each value, clear on update, invalid → 400 on both create and update), `TestTaskEndpoints_Priority` (HTTP-level round trip)
- ✅ **CP 1** verified live via curl: created a task with `priority: HIGH`, edited to `LOW` (204), cleared it (204, confirmed `null` via `GET /api/board`), invalid value `URGENT` → `400 VALIDATION_ERROR`

## Phase 2 — wrap ✅ COMPLETE (2026-09-05)

- [x] 2.1  Full regression — `go build/vet/test` (18 packages green, alongside MX4), `golangci-lint` (0 issues, one `staticcheck` S1016 finding fixed en route), `sqlc diff` (clean), `gofmt -l` (clean)
- [x] 2.2  Also fixed a completeness gap noticed while touching this code: `internal/export`'s own `Task`/`Habit` bodies were missing `goal_id`/`priority` and `target` respectively (both added in earlier milestones but never wired into the M8 export shape) — fixed and covered by the existing export round-trip test plus a live curl check. `docs/export-format.md` amended. `export`'s `Goal` also lacked `done_tasks`/`total_tasks` at the time — flagged as a known follow-up then, **and fixed 2026-09-05**: `export.TasksReader` gained `ProgressByGoal`, `export.Export` gained `GoalDoneTasks`/`GoalTotalTasks`, merged into `goalBody` the same way `goals.Handler.list` already does. New `TestExport_GoalProgress`; addendum in `docs/security-review-mx3.md`; live-verified byte-for-byte identical between `GET /api/goals` and `GET /api/export`.
- ✅ **CP 2** full suite green

### Done when

CP 1–2 met · a task can carry, edit, and clear a `HIGH`/`MEDIUM`/`LOW` priority ·
invalid values rejected · no effect on board ordering. **MX3-follow-up is complete
and shipped.**

---

# MX4 — Notes module

**Approved 2026-09-05**, product owner, narrowed scope (multi-select question — only
"plain-text body + title" was picked; tags, pin/favourite/archive/trash, and
categories-on-notes were offered and not picked). `v1.md §15` written (new section).

⚠ **Conflicts with `docs/design/design-system.md` §6.4**, which separately ratified
Notes as "must not be implemented." That ratification predates this approval and is
already stale with respect to MX1/MX3's shipped scope (categories-on-tasks/habits/goals
and goal-linked-tasks are both on that same §6.4 list, and are already built). The
product owner explicitly chose to override §6.4 for Notes specifically. **This agent
does not edit frontend files** — the frontend agent needs to reconcile §6.4 against the
real `v1.md` state (removing Notes, and the already-shipped MX1/MX3 items, from the
forbidden list). Flag this to the frontend agent/owner; don't silently proceed as if
§6.4 didn't exist.

## Decisions

| Topic | Default | Rationale / source |
|---|---|---|
| Fields | `title` (required, bounded like other titles), `body` (free text, generous bound e.g. 20,000 chars — no rich text, no structure) | `v1.md §15`: "plain text only." |
| No linkage | No `category_id`, no tags, no relation to any other module | Explicitly not picked in the scoping question; also sidesteps the reference doc's own unresolved "tags vs. categories" ambiguity. |
| No states | No pin/favourite/archive; delete is a real hard delete, not a trash/soft-delete | Explicitly not picked; matches `tasks.DeleteTask`'s existing hard-delete shape — no new soft-delete pattern needed anywhere in the codebase yet. |
| Export | Notes join the M8 export bundle (`v1.md §14` amended) | "the user can export all of their data" — notes are now part of "all". |
| Module shape | New `internal/notes`, one table `notes`, standard CRUD `Service` — the simplest module in the codebase (even simpler than `goals`: no target date, no progress, no category) | No cross-module dependency at all — no `CategoryChecker`-equivalent needed. |

## Phase 1 — `internal/notes` module ✅ COMPLETE (2026-09-05)

- [x] 1.1  Migration `000013_notes` — `notes` table: `id uuid PK`, `account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE`, `title text NOT NULL`, `body text NOT NULL DEFAULT ''`, `created_at`/`updated_at timestamptz NOT NULL DEFAULT now()`; index `(account_id, created_at DESC)`. Applied to dev + test DBs.
- [x] 1.2  `internal/notes/notes.go` — `Note`, `NoteInput{Title, Body string}`, `ErrNoteNotFound`, `ValidationError`, `Service` (`CreateNote`, `UpdateNote`, `DeleteNote`, `ListNotes` — `ListNotes` doubles as the export-reader method, mirroring `goals.Service.ListGoals`, no redundant `ListAll`)
- [x] 1.3  `internal/notes/queries.sql` + sqlc generate — `CreateNote`, `UpdateNoteFields`, `DeleteNote`, `ListNotes`. New `sqlc.yaml` block for `notesdb`.
- [x] 1.4  `internal/notes/service.go` — title required + bounded (200 chars); body bounded (20,000 chars); no cross-module checker needed (no linkage at all)
- [x] 1.5  `internal/notes/http.go` — `GET /api/notes`, `POST /api/notes`, `PATCH /api/notes/{id}`, `DELETE /api/notes/{id}`
- [x] 1.6  Wired in `cmd/server/main.go`: `notesSvc := notes.NewService(pool)`; `notes.NewHandler(notesSvc).Mount(mux, write, read)`
- [x] 1.7  Tests — `TestNoteLifecycle`, `TestNoteValidation`, `TestNoteIsolation`, `TestNoteEndpoints` (HTTP round trip) — all in `internal/notes/notes_test.go`
- ✅ **CP 1** verified live via curl: created a note, edited its title and body (204), listed it back with the edit reflected, deleted it (204), confirmed the list is empty afterward (hard delete, no trash)

## Phase 2 — Export + wrap ✅ COMPLETE (2026-09-05)

- [x] 2.1  `internal/export`: added `NotesReader` interface (`ListNotes`, reusing the shape directly — mirrors `GoalsReader`), wired `notes.Service` into `export.NewService` in `cmd/server/main.go`, added `notes` to the exported bundle. `docs/export-format.md` amended with a `Note` section.
- [x] 2.2  Full regression — `go build/vet/test` (18 packages green), `golangci-lint` (0 issues), `sqlc diff` (clean), `gofmt -l` (clean)
- [x] 2.3  `docs/security-review-mx4.md` — isolation review of the new module. **No findings** — single-table, account-scoped throughout, no cross-module reads except the export composition (same pattern as every other `*Reader` in `export`).
- ✅ **CP 2** verified live via curl: exported a fresh account's data after creating a note, confirmed it appears in the `notes` array of the export bundle; full suite + review green

### Done when

CP 1–2 met · a note can be created, edited, listed, and deleted · isolation tests green
· notes appear in a full account export · `docs/design/design-system.md` §6.4
sync flagged to the frontend agent (not this agent's file to edit). **MX4 is complete
and shipped.**

---

# MX-TL — Task / Time Block linkage

**Approved 2026-09-05**, product owner ("let's start implementation according to our
plan," approving the full recommended design from
`docs/architecture/task-timeblock-model-analysis.md`). `v1.md §3, §4, §5, §6, §7, §13`
and the domain-concepts intro amended in place. Full analysis, edge cases, and the
requirements-gate reasoning live in that doc — not repeated here.

## Decisions (from the analysis doc, all approved as recommended)

| Topic | Default | Rationale / source |
|---|---|---|
| Core linkage | `time_blocks.task_id`, nullable FK to `tasks(id)` — many blocks per task (planned and/or actual), one task per block | Mirrors the `goals`↔`tasks` shape exactly (MX3). |
| Category inheritance | `CHECK (task_id IS NULL OR category_id IS NULL)` at the DB layer — a task-linked block never stores its own category; resolved via a join to the task at every read path | Avoids denormalization drift; falls out for free when the task's category changes. |
| Delete default | `ON DELETE SET NULL` — deleting a task clears `task_id` on its blocks without deleting them | Mirrors MX3's `goal_id` delete precedent exactly. |
| API surface | Minimum only: `task_id` on block create/update/read; no `GET /api/tasks/{id}/blocks`, no state sync either direction | Analysis §5/§17; avoids unsupported-feature creep. |
| Analytics | `Comparison`/`ComparisonRange`/`DailyActualTotals`/`CountByCategory` must resolve inherited category before bucketing — required in this same change, not follow-up | Analysis §13 — without this, task-linked blocks silently misreport as "Uncategorized". |

## Phase 1 — Schema + `tasks.AssignableToAccount` ✅ COMPLETE (2026-09-05)

- [x] 1.1  Migration `000014_timeblock_task_link` — `ALTER TABLE time_blocks ADD COLUMN task_id uuid REFERENCES tasks(id) ON DELETE SET NULL`; `ADD CONSTRAINT time_blocks_category_xor_task CHECK (task_id IS NULL OR category_id IS NULL)`; index `(account_id, task_id)`. Applied to dev + test DBs.
- [x] 1.2  `tasks`: new `AssignableToAccount(ctx, accountID, taskID uuid.UUID) (bool, error)` on `Service` — satisfies `timeline.TaskChecker`. Also added `CategoriesForTasks` (bulk task→category lookup) in the same pass, needed by Phase 2.
- ✅ **CP 1** migration applied cleanly to dev + test DBs; `TestAssignableToAccount`/`TestCategoriesForTasks` cover own account / other account / unknown id.

## Phase 2 — `timeline`: linkage, inheritance, and analytics correctness ✅ COMPLETE (2026-09-05)

- [x] 2.1  `timeline.go`: new `TaskChecker` interface (`AssignableToAccount` + `CategoriesForTasks`) wired to `tasks.Service` in `cmd/server` (reordered: `goalsSvc`/`tasksSvc` now constructed before `timelineSvc`).
- [x] 2.2  `Block`/`BlockInput` gain `TaskID *uuid.UUID`; `AddBlock`/`EditBlock` validate it via `assertAssignableTask` and reject `CategoryID != nil && TaskID != nil` via `validateCategoryTaskExclusive` (`*ValidationError`, field `category_id`) — checked before either cross-module call.
- [x] 2.3  `blocksOverlapping` resolves each task-linked block's inherited category via `CategoriesForTasks` (only queried when at least one row in the window needs it) before returning — `Timeline`, `Comparison`, `ComparisonRange`, `DailyActualTotals` all get the fix for free.
- [x] 2.4  `CountByCategory` also resolves inherited categories via a new `CountBlocksByTask` query + `CategoriesForTasks`, so a task-linked block still counts toward its (inherited) category's total.
- [x] 2.5  Tests — `TestAddBlock_TaskLink`, `TestAddBlock_TaskAndCategoryMutuallyExclusive`, `TestEditBlock_TaskLink`, `TestDeleteTask_ClearsBlockLink`, `TestComparison_TaskLinkedBlockInheritsCategory`, `TestTimeline_TaskLinkedBlockShowsInheritedCategoryName`, `TestCountByCategory_TaskLinkedBlockCountsTowardInheritedCategory` — 40 tests total in `internal/timeline`, all green.
- ✅ **CP 2** verified live via curl: created a category-carrying task, added an actual block linked to it (no own category), `GET /api/timeline` and `GET /api/comparison` both correctly showed the task's category (`Deep Work`, 5400 actual seconds); deleted the task (204), block survived with `task_id: null` and `category_id: null`, reverted cleanly to Uncategorized.

## Phase 3 — HTTP + export ✅ COMPLETE (2026-09-05)

- [x] 3.1  `blockRequest`/`blockBody` gain `task_id`; `POST`/`PUT /api/blocks` accept/return it.
- [x] 3.2  `internal/export`: `Block` export shape gains `task_id` (raw); `docs/export-format.md` amended.
- ✅ **CP 3** verified live via curl (same session as CP 2) — a block create with both `task_id` and `category_id` correctly rejected with `400`; `GET /api/categories/overview` correctly counted the task-linked block toward its inherited category.

## Phase 4 — wrap ✅ COMPLETE (2026-09-05)

- [x] 4.1  Full regression — `go build/vet/test` (18 packages green), `golangci-lint` (0 issues), `sqlc diff` (clean), `gofmt -l` (clean).
- [x] 4.2  `docs/security-review-mx-tl.md` — isolation review of `TaskChecker` and the category-inheritance read paths. **No findings.**
- ✅ **CP 4** full suite + review green.

### Done when ✅ ALL MET (2026-09-05)

CP 1–4 met · a block can link to a task and inherits its category · a task-linked
block cannot carry its own category · every category-bucketing read path (timeline,
comparison, category counts) resolves the inheritance correctly · deleting a task
clears the link on its blocks without deleting them · isolation tests green.
**MX-TL is complete and shipped.**

## Addendum 2026-09-05 — `GET /api/tasks/{id}/blocks` (reverse lookup) ✅ COMPLETE

Found via a fresh read of `docs/left.md`: the frontend built the task→block display
direction (a block shows its linked task, with deep-links either way) and then hit a
real need for the reverse — viewing a task, see its scheduled blocks. This directly
reopened the `v1.md §7` line written during the original MX-TL pass ("no
`GET /api/tasks/{id}/blocks`") — amended in place rather than silently built around,
with the new bullet explaining the reversal and why.

- [x] `timeline.ErrTaskNotFound` sentinel; `Service.BlocksForTask(ctx, accountID, taskID) ([]Block, error)` — validates ownership via the existing `TaskChecker.AssignableToAccount`, then `ListBlocksByTask` (new query, reuses the `time_blocks_account_task_idx` index MX-TL already added), then resolves the task's inherited category once for the whole list via the existing `CategoriesForTasks`
- [x] `GET /api/tasks/{id}/blocks` on `timeline.Handler` (the module that owns `time_blocks`, even though the URL prefix reads like a `tasks` route) — returns `{blocks: [...]}` using the existing `blockBody` shape, unbounded across any date (unlike `Timeline`/`TimelineRange`, which are date-windowed)
- [x] Tests — `TestBlocksForTask` (multiple blocks across different dates, correct ordering, inherited category), `TestBlocksForTask_EmptyAndNotFound` (empty list vs. 404 for another account's/unknown task), `TestBlocksForTaskEndpoint` (HTTP-level) — 51 tests total in `internal/timeline`, all green
- ✅ **CP** verified live via curl: two blocks (one planned, one actual, different dates, weeks apart) linked to a category-carrying task; `GET /api/tasks/{id}/blocks` returned both, correctly ordered, both showing the task's inherited category; unknown task id → `404`
- Full regression green (`go build/vet/test`, `golangci-lint` 0 issues, `sqlc diff` clean — no schema change, the index already existed), addendum in `docs/security-review-mx-tl.md` — no findings

---

# MX5 — Timeline range-read endpoint (Calendar/Events unification parked)

**Investigated 2026-09-05** at the product owner's request to "tackle MX5." The
milestone as originally scoped in A.3 bundled two very differently-sized things; on
investigation, only one had a live requirement behind it.

## What investigation found

- The frontend agent had **already shipped** Week and Month timeline views
  (`web/src/features/timeline/WeekView.tsx`/`MonthView.tsx`, marked complete
  2026-09-05) — built entirely on the existing block model, no new entity, per the
  already-approved `v1.md §5` Week/Month amendment ("same block data, no new
  fields"). They work today by firing up to 42 individual `GET /api/timeline?date=`
  calls per month view — correct, but explicitly flagged in `docs/left.md` as
  **"(b) optimisation, not blocking,"** with the exact endpoint shape the frontend
  expects if built.
- **No requirement anywhere** — not `v1.md`, not `docs/left.md`, not a product-owner
  ask — calls for a distinct "Calendar" or "Event" entity. `design-system.md §6.4`'s
  exclusion of a standalone Calendar screen was never contradicted by an approval
  (unlike Notes), so there is nothing to reconcile and no ADR question to force by
  writing speculative scope.
- Conclusion, put to the product owner directly: build the small, already-authorized
  range-read endpoint now; leave Calendar/Events unification parked until an actual
  requirement for it exists. **Approved as recommended.**

## Phase 1 — `GET /api/timeline/range` ✅ COMPLETE (2026-09-05)

- [x] 1.1  `timeline.go`: new `RangeTimeline{From, To, Days []DayTimeline}`; `Service` gains `TimelineRange(ctx, accountID, from, to timezone.Date) (RangeTimeline, error)` — `*ValidationError` if `to` is before `from` or the range exceeds 62 days (enough for a month grid's 42 visible cells)
- [x] 1.2  `readmodel.go`: `TimelineRange` reuses `blocksOverlapping` **once** for the whole window (not once per day) plus one `NamesForAccount` call, then buckets per day — mirrors `categoryTotals`'s existing single-query-then-bucket shape. New private `daysInRange` helper (mirrors `internal/reports`'s identical function).
- [x] 1.3  `http.go`: `GET /api/timeline/range?from=&to=` returns `{from, to, days: [{date, planned, actual}]}` — exactly the shape `docs/left.md` specified, each day using the same `PositionedBlock` body the single-date endpoint already returns
- [x] 1.4  Tests — `TestTimelineRange_MatchesPerDayTimeline` (cross-checked against `Timeline` for identical output), `TestTimelineRange_MidnightBlockAppearsOnBothDays`, `TestTimelineRange_ToBeforeFrom`, `TestTimelineRange_ExceedsMaxDays`, `TestTimelineRange_Isolation`, `TestTimelineRangeEndpoint` (HTTP-level) — 6 new tests, all green
- ✅ **CP 1** verified live via curl: created blocks across a week, `GET /api/timeline/range?from=2026-08-31&to=2026-09-06` returned exactly 7 day entries with blocks correctly placed, matching `docs/left.md`'s expected shape field-for-field; a >62-day range and an inverted range both correctly rejected with `400`

## Phase 2 — wrap ✅ COMPLETE (2026-09-05)

- [x] 2.1  Full regression — `go build/vet/test` (18 packages green), `golangci-lint` (0 issues), `sqlc diff` (clean — no schema change, pure read composition), `gofmt -l` (clean)
- [x] 2.2  `docs/security-review-mx5.md` — isolation review. **No findings** — read-only, no new query shape beyond what `Comparison`/`ComparisonRange` already established, scoped by the same single `accountID` throughout.
- ✅ **CP 2** full suite + review green

### Done when ✅ ALL MET (2026-09-05)

CP 1–2 met · `GET /api/timeline/range` returns exactly what N individual `Timeline`
calls would · bounded to 62 days · isolation tests green · **frontend can now swap
`WeekView`/`MonthView`'s `Promise.all` fetch for one call, per `docs/left.md`'s own
documented swap point** (not done by this agent — frontend files off-limits).
Calendar/Events unification (MX5b) remains explicitly parked, no requirement to
build against.

---

# APPENDIX B — Timeline reference-gap plan (browser-verified 2026-09-04)

> Verified live: backend on `:8080` (`/readyz` ok) + Vite dev on `:5173`
> (proxy `/api → :8080`). Seeded account `seed@example.com` (Asia/Kolkata):
> 3 categories, planned `09:00–11:00` + actual `09:30–10:30` (Deep Work) for
> 2026-09-04. Logged in through the real Login page; `/timeline` (Day) and
> `/timeline?view=agenda` both render real data; comparison table correct
> (planned 2h / actual 1h / diff −1h). Screenshots: `/tmp/opencode/tl-day-desktop.png`,
> `tl-agenda-desktop.png`, `tl-day-mobile.png`. The two pre-login `401 /api/account`
> console entries are the expected auth-guard probe, not a defect.

## B.1 — Deliberate deviations (keep — required by v1.md / design-system §6.4)

| Reference element | Our rendering | Why it stays |
|---|---|---|
| Merged single-column titled block list | Two labelled lanes Planned \| Actual, dashed vs solid, category colour | `v1.md §5` requires the planned/actual distinction; the PNG shows none |
| Block titles, icons, tag chips | Time range + category name only | A block has only start/end/category (`v1.md §3/§4`) |
| Per-block checkbox | None | Blocks have no done state (spec `timeline.md`) |
| Avatars on "Team Sync" | None | P4 — no collaboration, permanent non-goal |
| View switcher Day/Week/Month/Agenda | Day/Agenda only | `v1.md §5`: "one date at a time; no week or month timeline" |
| Greeting header ("Good morning, Satyajit") | Factual `PageHeader` (eyebrow + date + plain subtitle) | VP3/D6 — no motivational copy |
| Search bar, notification bell, SPACES list, Focus/Pomodoro, quote cards, KPI row, donut, Insights, Top Priorities | Not built | All in `design-system.md §6.4` exclusion list |
| Agenda sort dropdown, list/grid toggle | Not built | Reference-only affordances, no V1 need |

## B.2 — Gaps to build (all V1-legit, no new backend)

- [x] **B2.1 `SplitButton` primitive** ✅ DONE 2026-09-04 — `components/ui/SplitButton.tsx`
      (+ test, barrel, `ui-split` styles), `BlockDialogTarget` kind preset, wired into
      `TimelineScreen` toolbar ("Add block ▾" → Add planned/actual). Browser-verified
      (menu opens, Actual preselected); typecheck + 28 timeline tests green; 0 console errors.
- [x] **B2.2 Agenda foot affordance** ✅ DONE 2026-09-04 — `AgendaList onAdd?` prop +
      dashed `.agenda__add` row ("＋ Add an agenda item") opening the same `BlockDialog`;
      wired in `TimelineScreen`. (+2 tests; browser check deferred to §B.5.)
- [x] **B2.3 Rail "Today's Tasks"** ✅ DONE 2026-09-04 — `features/timeline/TodayTasks.tsx`
      (read-only, tasks with `due_date === date` via existing `GET /api/board`, done
      styling + `n/m` count + link to `/tasks`; hides on empty/error so it never breaks
      the screen). Rail now mini-calendar + tasks. (+3 tests; browser check deferred to §B.5.)
- [ ] **B2.4 Now-line label re-verify** — the audit fixed the lane-selector bug; confirm
      the `10:24`-style time pill is visible on "today" at desktop + mobile during the
      B2 build's Playwright pass.
- [ ] **B2.5 Mobile Day check** — lanes currently squeeze inside `.tl2__scroll`
      (page never h-scrolls — good). Confirm Agenda remains the practical mobile view;
      only stack lanes vertically if the Playwright pass shows real collision.

## B.3 — Backend needs for Timeline: none

Day + Agenda run entirely on existing endpoints (`GET /api/timeline?date=`,
`GET /api/comparison?date=`, `GET /api/categories`, block CRUD). **No new backend
for this appendix.** Two conditional notes for the backend agent (do NOT build now):

- *If* Week/Month views are ever approved (requires a `v1.md §5` amendment first):
  add `GET /api/timeline?from=&to=` range read (comparison-range already exists from
  M6 Phase 1). Until then, unmarked and out of scope.
- MiniCalendar day-activity dots: decorative only — explicitly not worth an endpoint.

## B.4 — Build order (frontend agent)

1. B2.1 `SplitButton` (+ tests) → 2. wire into Timeline toolbar → 3. B2.2 agenda foot
   row → 4. B2.3 Today's Tasks rail → 5. Playwright pass (1440/390 + dark, B2.4/B2.5)
   → 6. `pnpm typecheck && pnpm test && pnpm build` → 7. visual QA vs
   `timeline.png` / `timeline-agenda.png` (Day/Agenda only) → 8. tick boxes here.

---

# APPENDIX C — Per-page reference-gap round (6 parallel agents, 2026-09-04)

> Method: seeded demo data (3 habits/1 marked today, 3 goals across all 4 states,
> tasks across BACKLOG/IN_PROGRESS), captured 11 live screenshots
> (`/tmp/opencode/pg-{tasks,board,habits,goals,categories,reports,reviews-daily,account,export,login,register}.png`
> at 1440×900, logged in as `seed@example.com` except auth). Six agents audited
> reference-vs-actual in parallel (Timeline done separately in Appendix B), wrote a
> per-page gap plan, and implemented it. Constraints: feature-dir only, no shared
> primitives/`api.ts` edits, tokens only, tests written but not run, no browser.
> Central `pnpm typecheck` after merge: **clean** (three transient cross-agent
> complaints resolved each other: reports JSX, Login `authBits`, habits `Last30`).

| Page | Plan file | Built | Deliberate keeps |
|---|---|---|---|
| Tasks + Board | `screens/tasks-board-gaps.md` | Checkbox uncheck restores previous state (was always TODO); 2 tests | No Starred/priority/tags/assignees/categories; plain counts; kebab covers keyboard-DnD |
| Habits | `screens/habits-gaps.md` | Rendered `last_30_days` (was API-only) in all 4 views via feature-local `Last30` bit; tests | No longest-streak/consistency/categories/sub-labels; Archive-only kebab; month stays mock |
| Goals | `screens/goals-gaps.md` | Token hygiene (`5px/4px/8px` → tokens); filter-exclusion test. No functional gaps | No %/bars/task-linkage/milestones/categories/On-Track wording; count-list rail |
| Categories | `screens/categories-gaps.md` | Token fix; first-row double-border fix; dialog trims name + disables empty save; tests | No counts/donut/icons/import/recent/archived-tab (C1); D2 colour-as-hash |
| Reports | `screens/reports-gaps.md` | Totals-row pos/neg; captions; bar aria-labels; empty states; removed `To max=today` (Q9); from/to inversion normalised; 1→2-col grid fix for table clipping | 5 R1 viz; no trends/deltas/insights/heatmap/focus/export; mocks intact |
| Reviews/Account/Auth | `screens/misc-gaps.md` | Account + Login/Register rewritten to primitives (PageHeader, Q4 `TimezoneSelect`, Q6 confirm/hints, brand tile, logout); review `.review-form`; 12/12 gaps; tests | Daily stays mock; weekly + export stay Placeholder; no profile fields |

## C.1 — Backend gaps surfaced (for the backend agent) — reconciled 2026-09-04

- [x] Reviews endpoints exist and **are now mounted** in `cmd/server` (M6 Phase 3, done 2026-09-04) — `GET/PUT /api/reviews/daily|weekly` live.
- [x] Weekly reference reads exist as of M6 Phase 1 (`GET /api/comparison?from=&to=`, `GET /api/habits/range?from=&to=`, `GET /api/tasks/throughput?from=&to=`) — **wiring gap, not a backend gap**; `WeeklyReviewScreen.tsx` still uses a mock per C.3, nothing left for the backend to build here.
- [ ] `GET /api/reports?from&to` — built, but as **5 separate routes with different field names** than `docs/left.md` Phase 9 specifies. **Not yet reconciled — see "NEXT — Contract reconciliation" → Phase R1 above.**
- [ ] `GET /api/habits/history` + optional `/week` — **not built** (M6/M7 only needed counts, not per-day date lists). **See Phase R2 above.**
- [ ] Category `PATCH {name}`-only wipes ADR-0009 colour/icon — confirmed real, **not yet fixed. See Phase R3 above.**
- [x] `Task` API now has `category_id` (MX1 Phase 3, done 2026-09-04) — resolved.
- [x] Export endpoint built (M8, done 2026-09-04) — `GET /api/export`, single JSON document (Q3 resolved). Categories unarchive: still not built — no approval to build it (`docs/left.md` Phase 8 says "not V1, confirm before building"; unchanged).

## C.3 — Weekly Review + Export built (2026-09-04, UI-first)

- [x] `features/reviews/WeeklyReviewScreen.tsx` — ISO week stepper (`?week=`,
      Mon-first D8), mocked weekly reference (category time, habit counts,
      tasks→DONE) with sample-data note, 4 fixed Q2 prompts, mock upsert store;
      `reviewData.ts` gains `WEEKLY_REVIEW_PROMPTS`/`fetchWeeklyReview`/
      `saveWeeklyReview`/`mockWeekReference`. Route wired, `+` test file.
- [x] `features/export/ExportScreen.tsx` — one-click provisional JSON snapshot
      (account, categories, tasks, habits, goals; reviews noted pending),
      download link, error alert. Route wired, `+` test file. Q3 still open.
- [x] `App.tsx` — no more `Placeholder` routes; every D10 route renders a screen.
- [x] `pnpm typecheck` clean. Full `pnpm test` + `pnpm build` + browser pass
      deferred to the testing phase (user call: build all UI first).

## C.2 — Deferred verification (the testing phase)

- [ ] Full `pnpm test` + `pnpm build` (agents wrote tests, none ran the suite)
- [ ] Playwright pass per page (1440/390 + dark, zero-console-errors) incl. B2.4/B2.5
- [ ] Promote feature-local bits (`Last30`, `TimezoneSelect`) to `components/ui` if a second screen needs them
- [ ] Tick B2.4/B2.5 + close Appendices B/C after the pass
