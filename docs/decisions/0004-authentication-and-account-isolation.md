# 0004 — Authentication & Account Isolation

**Status:** Accepted — 2026-09-03
**Applies to:** V1 (and beyond unless superseded)

## Context

V1 is multi-user with strict per-account data isolation (`docs/requirements/v1.md` §1,
N3), but conceptually personal (`docs/product/principles.md` P4 — authentication is
plumbing, never a product surface). Before the M1 authentication spec, the project
needs a fixed session architecture, a password-hashing algorithm, and an isolation
mechanism. Single-origin deployment (ADR-0001) simplifies the cookie posture. Several
security parameters are deliberately left to the M1 security design.

## Decision

**Sessions.**
- Server-side sessions. On login the server creates a session row in PostgreSQL
  (`id`, `account_id`, `created_at`, `expires_at`, `last_seen`) and returns an
  **opaque, high-entropy** token in a cookie marked `HttpOnly` and `Secure`.
- Authentication middleware resolves the cookie to a session row and puts the
  `account_id` into the request context. Logout deletes the session row. Changing the
  password invalidates all of that account's sessions.
- **No JWT.** No client-readable token; no stateless token.

**Password hashing.**
- **Argon2id** is the password-hashing algorithm. This ADR selects the algorithm only.

**Account isolation.**
- The authenticated `account_id` is taken **only** from the session context — never
  from a request body, path parameter, query parameter, or header.
- Every data-access method is account-scoped: `account_id` is a required parameter and
  appears in the `WHERE` clause of every query over account-owned data.
- Cross-account isolation is verified by **mandatory integration tests** asserting that
  one account cannot read or mutate another account's data, for every module that owns
  account data. This verifies N3; it is not new scope.
- **PostgreSQL Row-Level Security is not used in V1.** Isolation is enforced at the
  application layer plus the mandatory tests. RLS may be reconsidered only through a
  later ADR.

**Deferred — decided in the M1 authentication specification, not here:**
- Argon2id parameters: memory, iterations, parallelism, salt handling, encoding —
  selected and security-reviewed at M1.
- Session lifetime, idle timeout, and whether a "remember me" option exists.
- CSRF protection mechanism — M1 must evaluate the threat model for cookie-based auth
  and choose the approach explicitly.
- Login rate limiting — carried as an M1 security consideration requiring explicit
  product-owner approval; **not** a V1 requirement.
- Password policy (minimum length / rules) — product open question **Q6**, to be
  resolved before the M1 spec is Approved.

## Alternatives considered

- **JWT in a cookie.** Stateless, no session table. Rejected: revocation (logout,
  password change, "log out everywhere") needs a server-side denylist, which
  reconstructs the session table with extra moving parts. With one backend, a session
  lookup per request is free.
- **JWT in `localStorage` + `Authorization` header.** Rejected: readable by any
  injected script (XSS-exposed); an `HttpOnly` cookie is strictly safer.
- **A hosted auth provider (Auth0, Ory, SuperTokens).** Rejected: an external
  dependency with possible cost against E1, and heavy for email + password on a
  personal tool.
- **bcrypt for password hashing.** Fewer parameters to misconfigure and perfectly
  acceptable. Argon2id was chosen as the current best-practice memory-hard function;
  the parameter risk is handled by the M1 security review.
- **RLS as the primary isolation mechanism.** Strong defense in depth, but adds a
  per-connection session-variable dance with `pgxpool` and complicates debugging.
  Application-layer scoping plus tests is simpler and sufficient for V1; RLS remains
  available as a later hardening.

## Consequences

- One indexed session lookup per authenticated request — negligible at the N2 scale —
  in exchange for trivial, immediate revocation and a model that is easy to reason
  about.
- The session table is account-owned data like any other and follows the same
  isolation rules.
- Argon2id verification is CPU- and memory-intensive by design; the M1 parameter
  choice must balance resistance against the resource limits of the (not yet chosen)
  production host.
- "account_id only from the session" is a hard review checkpoint for every new
  endpoint and every new query.
- Deferring CSRF, rate limiting, and password policy keeps this ADR to architecture;
  those decisions have their own gates and owners.

## Revisit conditions

- The M1 security review finds Argon2id impractical on the chosen host's resources →
  revisit the algorithm (e.g. bcrypt) via an amendment or new ADR.
- A need for token-based auth for a non-browser client appears → reconsider (V2+).
- Isolation defects slip past application-layer scoping in review → reconsider adding
  RLS as defense in depth (new ADR).

## Related documents

- `docs/product/principles.md` — P4
- `docs/requirements/v1.md` — §1, N3; open question Q6
- `docs/architecture/conventions.md` — request-context helpers, account-scoping idiom
- ADR-0001, ADR-0002, ADR-0003, ADR-0005
