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
| Password policy | 12–128 characters; no composition rules; no denylist | Q6 (default) |
| Session lifetime | 30 days absolute; no idle timeout; no "remember me" | ADR-0004 deferred |
| Session token | 32 bytes from `crypto/rand`, base64url, stored as-is | ADR-0004 |
| CSRF | `SameSite=Lax` cookie + double-submit token header on state-changing requests | ADR-0004 deferred |
| Argon2id params | 19 MiB memory, time=2, parallelism=1, 16-byte salt, 32-byte key | ADR-0004 deferred |
| Login rate limiting | In-memory limiter keyed on email + client IP (5 failures / 15 min) | ADR-0004 deferred |
| Cookie | `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`, name `session` | ADR-0004 |
| Go version | `go 1.26.0` in `go.mod` (`x/crypto` requires it); toolchain auto-manages | ADR-0002 |
| UUID type | `github.com/google/uuid` for account/session ids (added — was not in the original dep list) | — |
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
- [x] 4.4  `Register` — validate email / 12–128 password / IANA timezone (`internal/platform/timezone`); case-insensitive uniqueness via `citext` + `23505`; account + session in one tx (`inTx` DBTX seam)
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

## Phase 8 — Frontend shell

- [ ] 8.1  `web/` — Vite + React + TS scaffold; typed API client wrapper; CSRF token handling
- [ ] 8.2  Register screen (email, password, timezone pre-filled from `Intl.DateTimeFormat().resolvedOptions().timeZone`)
- [ ] 8.3  Login screen (email, password)
- [ ] 8.4  Authenticated shell — show email + timezone, change-timezone control, change-password form, logout button; no product features
- [ ] 8.5  Route guards — logged-out → login on protected routes; logged-in → shell on login/register; auth state resolved by calling `GET /api/account` on load
- [ ] 8.6  `go:embed web/dist`; Go serves assets + SPA fallback for non-`/api`, non-health routes
- [ ] 8.7  Responsive check at 375px and 1280px — no horizontal overflow, every control operable
- ✅ **CP 8a** `make web-build && make run` — register through the browser, land on the shell, see email/timezone, log out
- ✅ **CP 8b** change timezone and password through the UI; after password change, back to login
- ✅ **CP 8c** renders clean at 375px and 1280px

## Phase 9 — CI

- [ ] 9.1  `.github/workflows/ci.yml` — Postgres service; steps: `go build`, `go vet`, `golangci-lint`, `go test ./...`, `sqlc diff`, `web` typecheck + build
- [ ] 9.2  Runs on push + PR
- ✅ **CP 9** the pipeline is green on a pushed branch

### M1 done when

CP 1a–9 all met · isolation suite green · security review clean · CI green · a person can
do the whole §1 flow through the browser.

---

# MILESTONE 2 — Categories and the timeline core

Covers `v1.md §2–§6` and completes **N4**. **Resolve first:** Q7 (overlap totals),
Q8 (Uncategorized bucket), Q9 (future-dated actual blocks) — record in `v1.md`.

- [ ] `platform/timezone` — date/week bucketing, range→instant-range, DST-correct, midnight-spanning; unit tests across zones/DST/midnight
- [ ] `internal/category` — create, rename, archive, list active; migration; isolation tests
- [ ] `internal/timeblock` — planned + actual blocks; `start < end`; optional category; cross-midnight; `planned|actual` immutable; migration; isolation tests
- [ ] Timeline read model — a date's planned + actual blocks together
- [ ] Planned-vs-actual per-category totals + differences for a date (applies Q7/Q8)
- [ ] API + frontend: category management; timeline view (hour grid, planned/actual distinguishable, correct cross-midnight placement); block add/edit/delete; per-date comparison
- ✅ **CP** timezone helpers pass the DST/midnight suite
- ✅ **CP** add planned + actual blocks for a date (incl. one crossing midnight), see them on the timeline, see per-category totals and variance
- ✅ **CP** isolation tests green; CI green

---

# MILESTONE 3 — Tasks and the Kanban board

Covers `v1.md §7–§8`. **Resolve first:** Q5 (ordering within a column), Q10 (how
"entered DONE in a range" is defined; confirm transition time is recorded).

- [ ] `internal/task` — create/edit/delete; state `BACKLOG|TODO|IN_PROGRESS|DONE` any direction; record DONE-entry events (for M6/M7); migration; isolation tests
- [ ] Board read model — tasks grouped by state, fixed column order
- [ ] API + frontend: task list; Kanban board with move-between-columns; 375/1280
- ✅ **CP** create tasks, move across all four columns, state persists
- ✅ **CP** DONE-entry events recorded; isolation + CI green

---

# MILESTONE 4 — Habits and streaks

Covers `v1.md §9`. **Resolve first:** Q9 (future-dated completions), Q11
(archive→unarchive preserves history for streaks).

- [ ] `internal/habit` — create; archive/unarchive; mark/unmark completion per date; completion history; streak (consecutive dates ending today/yesterday in account tz, one miss resets, no grace); migration; isolation tests
- [ ] Streak + completion-% logic — unit tests across tz/DST/midnight and reset behaviour
- [ ] API + frontend: habit list with current streak; per-date toggles; history; archive controls
- ✅ **CP** mark a habit several consecutive days → streak counts; skip a day → resets
- ✅ **CP** isolation + CI green

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
