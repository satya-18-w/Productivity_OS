# 0001 — Modular Monolith & Deployment Shape

**Status:** Accepted — 2026-09-03
**Applies to:** V1 (and beyond unless superseded)

## Context

The product constraints (`docs/product/principles.md` E1, E2; `docs/requirements/v1.md`
N1, N2) fix a small user base, a ₹0/month operating target, one Go binary, PostgreSQL
as the only datastore, and no distributed infrastructure. This ADR records the *shape*
of what is built and deployed — as an architectural decision, not only a principle —
and the relationship between the HTTP API and the web client.

It does **not** choose a hosting provider, a database provider, a container strategy,
or the mechanism by which frontend assets travel with the binary.

## Decision

- Productivity OS is a **modular monolith**: one Go process, internally divided into
  domain modules that interact only through published Go interfaces (see ADR-0002).
  There is no network hop between modules.
- The system is **one deployable unit**: the Go application. In production it serves
  **both** the JSON HTTP API and the built frontend static assets from a **single
  origin**.
- No message broker, cache, search engine, queue, or second datastore is part of the
  architecture. Adding any of these requires a new ADR recording a demonstrated need.
- PostgreSQL is the only datastore.
- **Deferred:** the production hosting provider and database provider (future
  ADR-0008, after a free-tier spike); the production container/image strategy; and the
  asset-packaging mechanism (`go:embed` vs. files shipped alongside the binary — an
  implementation decision, see ADR-0006).

## Alternatives considered

- **Separate static hosting for the frontend + an API service.** Two origins, meaning
  CORS configuration, `SameSite=None` cookies, and a second free service to keep
  running. Rejected: cross-origin auth complexity for no benefit at this scale, and it
  works against E1.
- **Microservices / service-per-domain.** Independent deployables with calls between
  domains. Rejected by E2: distributed failure modes with no matching benefit for a
  tens-of-users tool maintained by one person.
- **Serverless functions per endpoint.** Rejected: cold starts, state-management
  complexity, provider lock-in, and harder local development; one long-lived process
  is simpler and cheaper here.

## Consequences

- One artifact to build, test, deploy, and reason about; local and production
  topologies stay close.
- Same origin means the browser sends the session cookie automatically and there is no
  CORS layer (supports ADR-0004 and ADR-0006).
- The frontend and backend share a release cadence — a frontend-only change redeploys
  the whole unit. Acceptable for a solo project.
- Module boundaries are a compiler and code-review concern, not a network boundary:
  cheaper to enforce, easier to erode. ADR-0002 defines the enforcement.
- A genuine need for asynchronous processing or horizontal scaling becomes visible and
  forces an explicit ADR rather than creeping in.

## Revisit conditions

- A measured need for background / asynchronous work not served by a request-scoped
  path or a simple in-process scheduler.
- Sustained load a single modest instance cannot serve, well beyond the N2 envelope.
- A hosting outcome (ADR-0008) that makes single-origin asset serving impractical.

## Related documents

- `docs/product/principles.md` — E1, E2, E3
- `docs/requirements/v1.md` — N1, N2, N5
- `docs/architecture/overview.md`
- ADR-0002, ADR-0004, ADR-0006; ADR-0008 (deferred — production hosting)
