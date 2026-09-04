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

Covers `v1.md §10`. No blocking open questions.

- [ ] `internal/goal` — create/edit/delete; four-state progress (not started / in progress / achieved / abandoned), set manually; list; migration; isolation tests
- [ ] API + frontend: goal list, create/edit, progress control
- ✅ **CP** create a goal, set each progress state, edit, delete
- ✅ **CP** isolation + CI green

---

# MILESTONE 6 — Daily and weekly reviews

Covers `v1.md §11–§12`. Depends on M2, M3, M4. **Resolve first:** Q1 (daily prompts),
Q2 (weekly prompts), Q9, Q10.

- [ ] `internal/review` — daily review per date + weekly review per ISO week; fixed prompt sets; free-text answers; create/edit/view; migration; isolation tests
- [ ] Reference panel — reuse M2 totals, M4 completions, M3 DONE-count read-only, through those modules' public interfaces
- [ ] API + frontend: daily + weekly review forms and past views with reference panels
- ✅ **CP** complete a daily review with that day's totals shown; edit it; view a past one
- ✅ **CP** complete a weekly review for an ISO week with the week's figures; isolation + CI green

---

# MILESTONE 7 — Reports

Covers `v1.md §13`. Depends on M2, M3, M4. **Resolve first:** confirm Q7/Q8/Q10.

- [ ] `internal/reports` — read-only aggregation, never writes: Time by category · Planned vs actual by category · Habit completion · Task throughput · Daily actual totals; over a chosen range; DST-crossing ranges correct
- [ ] API (range params) + frontend: five report views; charts optional, figures fixed
- ✅ **CP** each of the five reports over a one-month range returns correct deterministic figures, including a range crossing a DST change
- ✅ **CP** CI green

---

# MILESTONE 8 — Data export

Covers `v1.md §14`, principle P5. Depends on every entity. **Resolve first:** Q3
(single JSON document vs archive of CSVs).

- [ ] `internal/export` — gather all account-owned data (categories, planned + actual blocks, tasks, habits + completions, goals, daily + weekly reviews); serialize to the chosen open format; documented schema; single user-initiated download; account-scoped
- [ ] API + frontend: export button + download
- ✅ **CP** export produces a file containing every entity the account created (round-trip completeness test)
- ✅ **CP** export contains only the caller's data (isolation test); CI green

---

# V1 DONE

- [ ] Every capability §1–§14 demonstrably works
- [ ] Q1–Q11 all resolved and reflected in `v1.md`
- [ ] N1–N7 verified: one binary; runs on free tiers; isolation tested; timezone / DST / midnight covered; responsive at 375 + 1280; a database restore performed successfully once
- [ ] Deployment spike done → **ADR-0008** (hosting provider, database provider, backup mechanism)
- [ ] No V1 non-goal present in the product
