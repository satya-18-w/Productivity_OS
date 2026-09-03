# 0006 — Frontend Architecture

**Status:** Accepted — 2026-09-03
**Applies to:** V1 (and beyond unless superseded)

## Context

The requirements call for a modern, responsive web client (N5); no native app, offline
mode, or PWA in V1. The application is authenticated and has no SEO or public-content
surface. It must be deployable at ₹0 (E1) and fit the single-origin shape of ADR-0001.
It needs enough interactivity for the timeline view (§5), the Kanban board (§8), and
report charts (§13).

## Decision

- **React + TypeScript + Vite.** A single-page application with client-side routing,
  built to static assets.
- **No server-side rendering and no Next.js.** The app is behind authentication with no
  SEO requirement, so SSR adds infrastructure (a Node runtime in production) for no
  benefit and works against E1.
- **Single origin.** The built static assets are served by the Go application
  (ADR-0001). There is no separate static host and no CDN in V1. The browser therefore
  talks to a same-origin API and the session cookie is sent automatically — no CORS
  layer.
- The frontend consumes the API contract and error envelope defined in ADR-0002.

**Deferred — not decided here:**
- The **asset-packaging mechanism** — `go:embed` into the binary vs. files shipped
  alongside the binary — is a build / implementation decision, made when the M1
  skeleton is built.
- The **data-fetching approach** (a query / cache library vs. hand-written hooks) —
  decided at M1, when there are real data-bound views.
- **Styling and component libraries** — documented separately (frontend section of
  `docs/architecture/conventions.md`) when frontend work starts. No libraries beyond
  this approved stack are added without demonstrated need.

## Alternatives considered

- **Next.js / Remix (React meta-frameworks).** Rejected: SSR / routing infrastructure
  and a production Node server; harder and potentially costlier to host for free;
  unnecessary for an authenticated app with no SEO need.
- **Svelte / Vue.** Both viable and lighter in places. React was chosen for the largest
  ecosystem (drag-and-drop, charts), the most transferable skill, and TypeScript
  maturity.
- **htmx + Go HTML templates.** Minimal frontend infrastructure and a strong "boring"
  fit, but the timeline positioning, Kanban drag-between-columns, and report charts are
  interactive enough that meaningful client-side JavaScript is unavoidable, and htmx's
  model would be fought rather than used.
- **Separate static hosting for the SPA.** Rejected here as in ADR-0001: two origins,
  CORS, `SameSite=None` cookies, a second free service.

## Consequences

- A frontend build step (Vite) and a JavaScript bundle shipped to the browser.
- One origin: no CORS configuration, straightforward cookie auth, one deployable
  (ADR-0001). Frontend-only changes redeploy the whole unit.
- TypeScript gives compile-time checking of the API client and component props; API
  types are kept in sync with the backend manually (small surface) unless a generator
  is adopted later.
- The interactive views will pull in focused libraries later; each is a milestone-level
  decision recorded in conventions, not an ADR.

## Revisit conditions

- A future need for public, indexable content or link previews → reconsider SSR / a
  meta-framework for that surface (V2+).
- Manual API / type drift between frontend and backend becomes error-prone → consider
  generating the client from a schema.
- Single-origin asset serving proves impractical on the chosen host (ADR-0008) →
  revisit the frontend / deployment relationship.

## Related documents

- `docs/requirements/v1.md` — N5; §5 (timeline), §8 (board), §13 (reports)
- `docs/architecture/conventions.md` — frontend implementation guidance (added at M1)
- ADR-0001, ADR-0002, ADR-0004
