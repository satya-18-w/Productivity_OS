# Security review — M1 (skeleton and authentication)

Date: 2026-09-04 · Reviewer: Claude (manual) · Scope: `internal/account`,
`internal/platform/{httpx,password,ratelimit,postgres}`, `cmd/server`, migration `000001`.

> The bundled `/security-review` skill could not run — it diffs against `origin/HEAD`
> and this repository has no remote. This is a manual review of the same surface against
> `CLAUDE.md` (account isolation, authorization, auth/session behaviour, API error
> exposure, deferred security decisions, ADR compliance).

## Verdict

**No blockers for M1** in the target context (personal tool, tens of accounts, ₹0
free-tier). Three low/low-medium findings were fixed in this pass; the rest are accepted
for V1 with rationale or deferred to a named decision.

## Checked and OK

| Area | Finding |
|---|---|
| Account isolation (N3, ADR-0004) | The acting account comes only from `IdentityFrom(ctx)`, set only by `RequireAuth` from the session cookie. Every account-scoped query is keyed by that id or by the caller's own opaque session token. No endpoint accepts an account selector; `DecodeJSON` rejects unknown body fields, so a smuggled `account_id` is a 400. Query params and `X-Account-Id`-style headers are ignored. Covered by `isolation_test.go` (7 tests). |
| Authorization | Only `POST /api/accounts` and `POST /api/sessions` are public. Everything else is behind `RequireAuth`; state-changing routes also behind `RequireCSRF`. |
| Password storage (ADR-0004) | Argon2id (19 MiB, t=2, p=1, 16-byte salt, 32-byte key), PHC string format, `subtle.ConstantTimeCompare` on verify. Plaintext is never persisted, logged, or returned (asserted in tests). |
| Login user-enumeration | Unknown email and wrong password return byte-identical `401 INVALID_CREDENTIALS`; a decoy Argon2id verify runs on the unknown-email path. Registration reveals a taken email (`409`) — accepted per `v1.md §1` ("no email verification"). |
| API error exposure (ADR-0002) | Every status ≥ 500 is logged server-side with the request id and returned as a generic `INTERNAL` with no SQL, stack, secret, or path. Panics recovered to the same envelope. `TestHardening_InternalErrorDoesNotLeak` forces a DB failure mid-request and asserts no leak. |
| SQL injection | All persistence via `sqlc`-generated parameterized queries; no string-built SQL. Injection-shaped email/password values are stored and compared as literals (`TestIsolation_SQLInjectionShapedValuesAreInert`). |
| Session expiry | `ResolveSession` rejects a session at/after `expires_at`; logout deletes the row; password change deletes every row for the account in one transaction. |
| CSRF | `SameSite=Lax` session cookie + double-submit `csrf_token` cookie / `X-CSRF-Token` header, constant-time compared, required on all unsafe methods. |
| Body limits | 1 MiB request cap; malformed / oversized / trailing-data bodies → 400, never 500. |
| Cookies | Session cookie `HttpOnly`, `SameSite=Lax`, `Secure` when `ENV=production`, `Path=/`. |

## Fixed in this pass

| # | Finding | Fix |
|---|---|---|
| F1 (low) | Session tokens were stored verbatim — a database dump would yield usable tokens. | Store `sha256(token)`; the raw token lives only in the client cookie. `TestRegister_CreatesAccountAndSession` asserts the raw token is not in the table. |
| F2 (low-med) | `POST /api/accounts` had no rate limit; both auth endpoints run a 19 MiB Argon2id hash per request → unauthenticated memory/CPU DoS. | Added a per-IP budget (30 requests / 5 min) on both auth endpoints (`throttleByIP`). |
| F3 (low) | The in-memory rate-limiter map never shrank. | Opportunistic sweep of expired windows every 256 operations. |
| F6 (low) | A transient `TouchSession` failure returned 500 and logged the user out. | `last_seen_at` update is now best-effort (logged, not fatal). |

## Accepted for V1 / deferred

| # | Item | Disposition |
|---|---|---|
| F4 | `POST /api/accounts` and `POST /api/sessions` are CSRF-exempt (no token exists pre-auth). | Accepted. Login/registration CSRF is low severity; `SameSite=Lax` still applies. |
| F5 | The CSRF token is a stateless double-submit value, not HMAC-bound to the session. | Matches the planning decision. HMAC-binding is optional future hardening. |
| F7 | `clientIP` reads `RemoteAddr`; behind a proxy every user could share one bucket, and naively trusting `X-Forwarded-For` would be spoofable. | Deferred to **ADR-0008** (deployment/proxy shape). Marked with a `NOTE` in `http.go`. |
| — | No breached-password denylist. | Explicitly out of V1 per Q6. |
| — | Password policy: min 6, composition rules (lower/upper/special). | Product-owner decision (Q6, 2026-09-04). Note: a short minimum plus composition rules is weaker against modern guessing than a longer minimum with no composition — accepted for a personal tool. Revisit if the threat model changes. |
| — | No `__Host-` cookie prefix. | Optional hardening; revisit with ADR-0008 (needs HTTPS + no `Domain`). |

## Re-verification

`go build` · `go vet` · `golangci-lint run` (0 issues) · `go test ./...` — all green after
the fixes.
