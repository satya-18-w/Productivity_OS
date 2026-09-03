# 0002 — Backend HTTP/API & Module Architecture

**Status:** Accepted — 2026-09-03
**Applies to:** V1 (and beyond unless superseded)

## Context

Before M1 implementation, the backend needs a fixed internal structure, an HTTP layer,
and an API contract. Constraints: E2 (enforced module boundaries), E3 (boring,
well-understood tech; justify dependencies), E4 (spec before code); N1 (single binary).
This sits inside the monolith shape of ADR-0001. Account scoping is covered by ADR-0004.

## Decision

**Toolchain and module.** The backend is a single Go module. The Go version is pinned
explicitly through the `go` directive in `go.mod` (and the `toolchain` directive if
used) and matched in CI. The concrete version is chosen when the project is initialized
(M1) — it must be recent enough for the standard-library router (Go 1.22+) — and is
moved forward deliberately, never mid-milestone.

**HTTP layer.** The standard library `net/http` with `http.ServeMux` (method + path
pattern routing). Middleware — request ID, structured logging (`log/slog`), panic
recovery, authentication — is written by hand as `http.Handler` decorators. **No web
framework.**

**Package / module structure.**
- `cmd/server/` — `main`, config loading, and composition (wiring modules together).
- `internal/<domain>/` — one package per domain module (e.g. `account`, `timeblock`,
  `task`, …); the list emerges per milestone.
- `internal/platform/` — infrastructure with no business logic: database pool, config,
  clock, timezone helpers, shared HTTP helpers.
- Each domain module exposes **exactly one Go interface at its package root** as its
  entire public surface. Handlers, services, and data access are unexported or in
  child packages. A module may depend on another only through that exported interface,
  wired in `cmd/server`. This is the enforcement of E2; a violation is a defect.

**API contract.** JSON over HTTP, resource-oriented. All endpoints under a single
`/api` prefix, with **no version segment**. Change rule: additive changes only; a
breaking change would introduce `/api/v2`.

**Validation.** Performed at the application boundary (request decoding / service
entry) and returning structured, field-level errors. Implemented with explicit
standard-library code — no validation library. Implementation idiom lives in
`docs/architecture/conventions.md`.

**Error response contract.** Every error response uses one envelope:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "fields": { "title": "title is required" }
  }
}
```

- `code` — a stable machine-readable `UPPER_SNAKE_CASE` slug.
- `message` — a short human-readable summary.
- `fields` — optional; present for validation failures, mapping input field → message.
- The same envelope applies to 400, 401, 403, 404, 409, and 500, each with the
  appropriate HTTP status code.
- **500 responses never expose internal implementation details, database errors, stack
  traces, or secrets.** They return a generic `code` and `message`; detail is logged
  server-side only.
- A central error-writing helper and the panic-recovery middleware are the only places
  that emit this envelope.

The API error contract above is distinct from the validation implementation.

## Alternatives considered

- **A web framework (gin, echo, Fiber).** Faster initial setup, but obscures the
  request lifecycle (E3). The Go 1.22+ stdlib router plus hand-written middleware
  covers the need with zero dependencies.
- **A third-party router only (chi).** Minimal and `net/http`-compatible. Not adopted
  now, to keep the dependency count at zero; may be adopted later if hand-written
  routing/middleware shows real friction — a contained, reversible change.
- **Layer-first packages** (`internal/handlers`, `internal/services`,
  `internal/store`). Rejected: organizes by technical layer, not domain; boundaries
  blur; contradicts E2.
- **Path-versioned API (`/api/v1`) from the start.** Rejected as unnecessary for a
  single first-party client; `/api/v2` can be added later as a pure prefix.
- **RFC 9457 `application/problem+json`.** A real standard, but more verbose and needs
  its vocabulary mapped onto the project's needs for a single consumer. The small
  custom envelope is simpler to produce and consume.
- **A validation library (`go-playground/validator`).** Reflection- and tag-driven;
  message mapping and tag/logic drift. Deferred — reconsider only on demonstrated
  friction.

## Consequences

- No web-framework dependency; the request path is fully visible and debuggable.
- The one-interface-per-module rule gives clear test seams and keeps cross-module
  coupling explicit, at the cost of discipline (cross-module types stay unexported).
- The API contract is small and stable; the frontend handles one predictable error
  shape.
- Hand-written validation means more code per input type, in exchange for total
  clarity and no magic.
- The 500-response rule requires every handler path to route internal errors through
  the central writer; enforced by review and the recovery middleware.

## Revisit conditions

- Hand-written routing or middleware becomes a maintenance burden → consider a minimal
  router (e.g. chi).
- Explicit validation becomes repetitive across many input types → reconsider a
  validation library.
- A second, third-party API consumer appears → reconsider path versioning.
- Internal service/store layering (owned by `conventions.md`) proves insufficient for
  a complex module → that module may add structure without changing this ADR.

## Related documents

- `docs/product/principles.md` — E2, E3, E4
- `docs/requirements/v1.md` — N1, N3
- `docs/architecture/conventions.md` — error idiom, naming, layering detail
- ADR-0001, ADR-0003, ADR-0004, ADR-0006
