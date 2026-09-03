# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project status

Productivity OS is a personal productivity web application, in early development under
Spec-Driven Development. Approved and committed so far: the product baseline
(`docs/product/`, `docs/requirements/v1.md`, `docs/roadmap.md`), the SDD / engineering
documentation structure (`docs/specs/`, `docs/decisions/`, `docs/architecture/`), and
the foundational architecture decisions (ADR-0001–0007).

**No application code exists yet.** Milestone M0 (foundation) is in progress;
implementation begins at M1.

Product requirements live in `docs/requirements/v1.md`. They are not extended or
reinterpreted in code — proposed changes go back to the product owner.

Remote: `git@github.com:satya-18-w/Productivity_OS.git` (branch `main`).

## Development workflow (mandatory)

Spec-Driven Development. Every milestone moves through eight stages in order, and **each
stage produces a written artifact that must be agreed before the next stage begins**:

```
IDEA → SPECIFICATION → PLAN → IMPLEMENTATION → TEST → REVIEW → ACCEPTANCE → COMMIT
```

| Stage | Artifact | Gate |
|-------|----------|------|
| IDEA | `docs/product/*` | Authored by the product owner; Claude challenges assumptions |
| SPECIFICATION | `docs/specs/v1/M<n>-<slug>/spec.md` | Unambiguous, testable; every consumed open question resolved |
| PLAN | `…/plan.md` | Approved by the product owner before any edit |
| IMPLEMENTATION | code | Matches the approved plan; small, focused commits |
| TEST | tests + real run output | Tests pass; failures and skips reported honestly |
| REVIEW | review notes | `/code-review`, and `/security-review` where relevant |
| ACCEPTANCE | `…/acceptance.md` | Product owner verifies each criterion |
| COMMIT | conventional commit / PR | Only on the product owner's explicit say-so |

Do not skip stages. Do not start implementation without an approved `plan.md`. Do not
commit without explicit acceptance. Milestone specifications are written immediately
before the milestone, not in advance (`docs/specs/README.md`).

## Architectural constraints (hard)

- **Modular monolith.** One deployable Go binary of internally-separated modules. No
  microservices, message broker, cache, or search engine unless a later demonstrated
  need is recorded in an ADR.
- **Module boundaries are enforced.** A module's exported interface is its only public
  surface; no module imports another's internal packages (ADR-0002).
- **Go backend. PostgreSQL** as the only datastore.
- **React + TypeScript + Vite** frontend, served single-origin by the Go binary
  (ADR-0006).
- **₹0/month running cost target.** Every infrastructure and hosting choice must be
  free-tier viable. Local Postgres runs in Docker; the production hosting and database
  provider are not yet selected (ADR-0008, pending a spike).
- **Small user base.** Optimize for simplicity, correctness, and low operational
  burden, not for scale.
- **Deterministic V1.** No AI, prediction, or heuristics.
- **Simple infrastructure.** Prefer boring, well-understood tooling; justify every
  added dependency.

## Documentation map

| Path | Holds |
|------|-------|
| `docs/product/` | Vision and principles (stable) |
| `docs/requirements/v1.md` | V1 scope — testable, implementation-independent |
| `docs/roadmap.md` | Milestone order M0–M8; V2 / V3 themes |
| `docs/architecture/overview.md` | Living architecture overview |
| `docs/architecture/conventions.md` | Idiomatic implementation guidance |
| `docs/decisions/` | Architecture Decision Records (`NNNN-*.md`) + `README.md` index |
| `docs/specs/` | Per-milestone SPEC / PLAN / ACCEPTANCE + `README.md` |

Approved ADRs: 0001 deployment shape · 0002 backend HTTP / API / modules · 0003
PostgreSQL persistence · 0004 authentication & account isolation · 0005 time & timezone
· 0006 frontend architecture · 0007 testing & local development. ADR-0008 (production
hosting) is deferred until a free-tier spike.

## Toolchain present on the dev machine

Go 1.22.2 (upgrade to latest stable — ADR-0002), Node 22, npm / pnpm, Docker + Compose,
`make`, `golang-migrate` (`migrate`), `jq`, `git`. Not installed: `psql`, `gh`, `sqlc`.

## Commands

No build, test, lint, or run commands exist yet. They will be added here and in
`docs/architecture/conventions.md` when the M1 skeleton and its `Makefile` land.
