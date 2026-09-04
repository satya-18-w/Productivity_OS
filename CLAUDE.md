# CLAUDE.md

Durable operating rules for the agent working in this repository. It points at the
docs; it does not restate them.

## What this project is

Productivity OS is a personal productivity web application: an authenticated user plans
their day as time blocks, records what actually happened, and reviews where their time
went. It is multi-user with strict per-account isolation, but has no collaboration or
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

## How we work

We build directly against `planning.md` — a phased implementation plan with checkpoints.
There is no separate spec/plan/approval gate.

- `planning.md` lists ordered phases; each phase has sub-tasks and one or more
  **checkpoints** — an observable result that proves the phase works.
- Implement the current phase, tick its tasks, and stop at each checkpoint so the
  product owner can see it working.
- Tests are written as part of the phase, not afterwards. A checkpoint that needs a
  passing test is not met until the test passes.
- Keep commits small and phase-scoped. Commit when a checkpoint passes or the product
  owner asks — never otherwise.
- `planning.md` is kept current: check off completed work, and refine later phases when
  we reach them.

## Source of truth

| Document | Authoritative for |
|----------|-------------------|
| `docs/product/vision.md`, `docs/product/principles.md` | Product direction and the decision filters |
| `docs/requirements/v1.md` | Approved V1 product requirements — scope, behaviour, non-goals, open questions |
| `docs/roadmap.md` | Milestone order (M1–M8) |
| `docs/decisions/` (ADRs) | Architectural decisions and their rationale |
| `docs/architecture/overview.md` | The current architectural shape |
| `docs/architecture/conventions.md` | Implementation idioms — style, naming, layering detail, commands |
| `planning.md` | The active implementation plan and its checkpoints |
| `CLAUDE.md` (this file) | Durable agent operating rules |

## Precedence when instructions or documents conflict

Highest wins:

1. An explicit current instruction from the product owner in this conversation.
2. Approved product requirements and product decisions (`docs/requirements/v1.md`).
3. Approved ADRs (`docs/decisions/`).
4. `planning.md`.
5. Patterns already established in the implementation.
6. Claude's own assumptions — last resort; prefer asking.

**Conflict behaviour:**

- A current product-owner instruction may change a product or architecture decision, but
  it does not silently invalidate the approved documentation. If the owner is changing a
  decision, update the authoritative document (`v1.md` or the relevant ADR) first, then
  implement against the updated decision.
- If a requirement and an ADR appear to conflict, stop and escalate — do not pick a
  winner.
- Do not resolve an open product question (Q1–Q11 in `v1.md`) by assumption. Sensible
  build-time defaults may be used to keep moving **only** when recorded in `planning.md`
  and flagged for the owner.

## Before modifying code

1. Inspect repository state (`git status`, branch, recent commits).
2. Read the docs relevant to the work — the `planning.md` phase, the ADRs and
   requirement sections it touches.
3. Inspect the existing code the change affects.
4. Check for working-tree changes unrelated to this task.
5. Plan non-trivial work before implementing it.

## Claude must not, without explicit product-owner approval

- Invent product requirements, or build features not in `docs/requirements/v1.md`.
- Silently resolve an open product question (Q1–Q11 in `v1.md`).
- Refactor code unrelated to the current phase.
- Introduce infrastructure the hard constraints exclude (broker, cache, queue, search,
  Row-Level Security, a second datastore, a second deployable), or any infrastructure
  without a demonstrated need.
- Adopt a dependency not already recorded in `planning.md`. Standard-library
  functionality needs no approval; convenience is not a reason.
- Make an architecturally significant change without an ADR and approval.
- Widen or break an API contract.
- Create slash commands, subagents, or MCP servers.
- Commit, push, open a PR, deploy, or select a hosting or database provider.

## Module boundaries

Domain modules must not import another domain module's internal implementation;
cross-module interaction goes through the owning module's public interface. (ADR-0002)

## Database and schema changes

- Schema changes use **forward-only** migrations in the single migrations directory;
  never edit a migration that has shipped — fix forward with a new one.
- A schema change is represented in `planning.md`.
- It includes appropriate tests, and considers indexes, constraints, transaction
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

- Tests are part of implementation, not an optional follow-up. A phase is not complete
  because the code compiles.
- Add the unit, integration, and — where a checkpoint calls for it — end-to-end tests
  the work requires.
- Integration tests run against a real PostgreSQL; isolation tests are mandatory for
  modules owning account data; time-dependent logic is tested across timezone, DST, and
  midnight-spanning cases. (ADR-0007)

## Error handling

- API responses follow the approved error contract — the custom envelope, correct status
  codes.
- Internal implementation and database errors, stack traces, and secrets are never
  exposed through 500 responses. (ADR-0002)

## ADRs

- Create or update an ADR when making a meaningful architectural decision, or changing
  one.
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
  them out instead, and stage only what belongs to this phase with a path-scoped
  `git add`.
- Before declaring a checkpoint met: inspect `git diff` and `git status`, run the
  relevant verification, report the files changed, and flag anything unexpected.
- Do not commit unless a checkpoint passes or the product owner asks.

## Definition of done (per phase)

- The phase's checkpoints are demonstrably met.
- The relevant tests pass; failures and skips are reported honestly.
- The relevant documentation (`planning.md`, and `conventions.md` / an ADR where the
  work established something durable) is updated.
- No unexplained or unrelated changes remain in the tree.

## Not in this file

Architecture rationale → the ADRs. Full requirements and scope-boundary text →
`docs/requirements/v1.md`. Coding style, naming, error / log idioms, test layout,
tooling config, and build / test commands → `docs/architecture/conventions.md`.
Milestone order → `docs/roadmap.md`. Active plan and checkpoints → `planning.md`.
Toolchain versions → `go.mod` and CI.
