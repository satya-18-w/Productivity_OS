# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working in this repository.
It holds durable operating rules for the agent. It is not documentation — it points to the
docs and does not restate them.

## What this project is

Productivity OS is a personal productivity web application: an authenticated single user
plans their day as time blocks, records what actually happened, and reviews where their
time went. It is multi-user with strict per-account isolation, but has no collaboration or
social features.

**Hard constraints — do not violate. Changing any requires a product-owner decision, and
an ADR where it is architectural:**

- Modular monolith: one Go backend process; PostgreSQL as the only datastore.
- React + TypeScript + Vite frontend, served single-origin by the Go process.
- No microservices, message broker, cache, queue, or search engine unless a demonstrated
  need is approved in a new ADR.
- No unnecessary distributed infrastructure.
- V1 is deterministic — no AI, prediction, or heuristics.
- ₹0/month infrastructure — every hosting or service choice must be free-tier viable.
- Claude does not invent, extend, or reinterpret product requirements.

## Source of truth

| Document | Authoritative for |
|----------|-------------------|
| `docs/product/vision.md`, `docs/product/principles.md` | Product direction and the decision filters |
| `docs/requirements/v1.md` | Approved V1 product requirements — scope, behaviour, non-goals, open questions |
| `docs/roadmap.md` | Milestone order and project status |
| `docs/decisions/` (ADRs) | Architectural decisions and their rationale |
| `docs/architecture/overview.md` | The current architectural shape |
| `docs/architecture/conventions.md` | Implementation idioms — style, naming, layering detail, commands |
| active milestone spec — `docs/specs/v1/M<n>-<slug>/` | The exact implementation scope for the milestone in progress |
| `CLAUDE.md` (this file) | Durable agent operating rules |

## Precedence when instructions or documents conflict

Highest wins:

1. An explicit current instruction from the product owner in this conversation.
2. Approved product requirements and product decisions.
3. Approved ADRs.
4. The active milestone specification.
5. This file's engineering workflow rules.
6. Patterns already established in the implementation.
7. Claude's own assumptions — last resort; prefer asking.

**Conflict behaviour (strict):**

- A current product-owner instruction may *initiate* a change to a product or
  architecture decision. It does **not** silently invalidate approved documentation.
- If a current instruction conflicts with an approved requirement, ADR, or the active
  spec: **stop the conflicting work**, explain the conflict, identify which authoritative
  document must change, and require that document to be updated and approved **before**
  the conflicting implementation is committed.
- If the product owner is changing a decision: update the authoritative document first,
  then implement against the updated decision.
- The active milestone spec never overrides a requirement or an ADR. If a milestone
  appears to need more than the requirements allow, that is a product-owner decision and
  the spec is blocked. If it appears to need a departure from an ADR, that is a new or
  amended ADR and the spec is blocked.
- If a requirement and an ADR themselves appear to conflict, stop and escalate — do not
  choose a winner.

## Engineering workflow

Requirements and product decisions are upstream and fixed for V1; IDEA is a product-owner
precursor, outside this workflow. Every milestone moves through these stages in order, and
each produces a written artifact that must be agreed before the next begins:

```
SPEC → PLAN → HUMAN APPROVAL → IMPLEMENT → TEST → REVIEW → ACCEPTANCE → COMMIT
```

- **SPEC / PLAN** — written under `docs/specs/v1/M<n>-<slug>/` from the templates. A spec
  may only *cover* requirements, never widen or reinterpret them. Every open question it
  consumes is first resolved in `docs/requirements/v1.md`.
- **HUMAN APPROVAL** — the product owner approves `spec.md` and `plan.md` (the spec
  reaches "Approved" per `docs/specs/README.md`). No implementation before this.
- **IMPLEMENT** — match the approved plan; small, focused commits; tests written as part
  of the work.
- **TEST** — the relevant tests pass; failures and skips are reported honestly.
- **REVIEW** — `/code-review`. `/security-review` is **mandatory** for any milestone
  touching authentication, sessions, account isolation, authorization, or API error
  handling / internal-error exposure.
- **ACCEPTANCE** — the product owner verifies each criterion in `acceptance.md`.
- **COMMIT** — only when the workflow reaches this stage or the product owner asks.

Do not skip stages. Milestone specs are written immediately before the milestone, not in
advance. Once the CI pipeline of ADR-0007 exists, its required checks must be green before
a milestone reaches ACCEPTANCE or COMMIT.

## Before modifying code

Do all of:

1. Inspect repository state (`git status`, current branch, recent commits).
2. Read the authoritative documents relevant to the work — the active spec, its plan, and
   the ADRs and requirement sections they cite.
3. Identify the active milestone specification and confirm it is Approved.
4. Inspect the existing code the change will affect.
5. Check for working-tree changes unrelated to this task.
6. Plan non-trivial work before implementing it.

If there is no Approved spec and plan for the work, stop and raise it.

## Claude must not, without explicit product-owner approval

- Invent product requirements, or add features not in the active spec.
- Silently resolve an open product question (Q1–Q11 in `v1.md`) — these are resolved by
  the product owner and recorded in `v1.md`.
- Refactor code unrelated to the current task.
- Introduce infrastructure the hard constraints exclude (broker, cache, queue, search,
  Row-Level Security, a second datastore, a second deployable), or any infrastructure
  without a demonstrated need.
- Adopt a dependency not already present in the approved plan. Standard-library
  functionality needs no approval; convenience is not a reason.
- Make an architecturally significant change without an ADR and approval.
- Widen or break an API contract.
- Create slash commands, subagents, or MCP servers.
- Commit, push, open a PR, deploy, or select a hosting or database provider.

## Module boundaries

- Module boundaries are mandatory: domain modules must not import another domain module's
  internal implementation; cross-module interaction goes through the owning module's
  public interface. (ADR-0002)

## Database and schema changes

- Schema changes use **forward-only** migrations in the single migrations directory; never
  edit a migration that has shipped — fix forward with a new one.
- A schema change must be represented in the active milestone specification.
- It must include appropriate tests, and must consider indexes, constraints, transaction
  boundaries, and account isolation.
- Regenerate `sqlc` and keep `sqlc diff` clean. (ADR-0003)

## Account isolation

For authenticated data:

- `account_id` comes only from the authenticated session context — never from a request
  body, path, query, or header.
- Every persistence operation over account-owned data is scoped by `account_id`.
- Every module that owns account data has cross-account isolation tests. (ADR-0004)

## Time and timezone

All date, week, and range bucketing goes through the platform timezone helpers; never
scatter `AT TIME ZONE` through SQL or rely on the server's local time. (ADR-0005)

## Testing

- Tests are part of implementation, not an optional follow-up. A milestone is not complete
  because the code compiles.
- Add the unit, integration, and — where the acceptance criteria call for it — end-to-end
  tests the milestone requires.
- Integration tests run against a real PostgreSQL; isolation tests are mandatory for
  modules owning account data; time-dependent logic is tested across timezone, DST, and
  midnight-spanning cases. (ADR-0007)

## Error handling

- API responses follow the approved error contract — the custom envelope, correct status
  codes.
- Internal implementation and database errors, stack traces, and secrets are never
  exposed through 500 responses. (ADR-0002)

## ADRs

- Create or update an ADR when making a meaningful architectural decision, or changing an
  existing one.
- Do not create an ADR for ordinary implementation details, routine configuration, or
  every dependency. Product decisions are never ADRs.
- ADR conventions: `docs/decisions/README.md`.

## MCPs and subagents

- Use an MCP only when the capability it provides materially improves the task.
- Use a subagent only when the work decomposes into genuinely independent concerns.
- Do not spawn subagents or install MCPs by default, and do not create custom ones
  without approval.

## Git and working tree

- Never overwrite or bundle in working-tree changes unrelated to the current task; point
  them out instead, and stage only what belongs to this milestone with a path-scoped
  `git add`.
- Before declaring any stage complete: inspect `git diff` and `git status`, run the
  relevant verification, report the files changed, and flag anything unexpected.
- Do not commit unless the workflow reaches the COMMIT stage or the product owner asks.

## Definition of done

A milestone is complete only when:

- the implementation matches the Approved spec;
- the relevant tests pass;
- review is complete, including `/security-review` where required;
- every acceptance criterion is satisfied;
- the relevant documentation is updated;
- no unexplained or unrelated changes remain in the tree.

## Not in this file

Architecture rationale → the ADRs. Full requirements and scope-boundary text →
`docs/requirements/v1.md`. Coding style, naming, error / log idioms, test layout, tooling
config, and build / test commands → `docs/architecture/conventions.md`. Milestone order
and status → `docs/roadmap.md`. Toolchain versions → `go.mod` and CI (the Go version is
pinned at M1 — ADR-0002).
