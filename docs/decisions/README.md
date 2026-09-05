# Architecture Decision Records

An ADR records one architectural or engineering decision: the context, the options
considered, the choice, and the consequences. ADRs are how a future reader understands
*why* the system is the way it is.

## What earns an ADR

Record a decision when it is **hard to reverse**, **affects more than one module**, or
**constrains future work**. Examples: language runtime version, frameworks, persistence
approach, infrastructure, hosting, CI, authentication mechanism, module-boundary rules.

## What does not

- Product decisions — those live in `docs/product/` and `docs/requirements/`.
- Routine library choices a single milestone's `plan.md` can make and undo.
- Tool versions, config values, and settings with no lasting consequence.

Do not create an ADR for every tool, dependency, version, or configuration choice.

## Conventions

- **Filename:** `NNNN-kebab-title.md`, zero-padded, assigned once, never reused or
  renumbered.
- **One decision per ADR.**
- **Status:** `Proposed` → `Accepted` → `Superseded` (or `Deprecated`).
- The **Decision** text of an Accepted ADR is never edited. A reversal is a new ADR
  that supersedes the old one; the old one gets a `Superseded by NNNN` line at the top
  and is otherwise left intact.
- Start from `0000-template.md`.

## Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [0001](0001-modular-monolith-and-deployment-shape.md) | Modular Monolith & Deployment Shape | Accepted | 2026-09-03 |
| [0002](0002-backend-http-api-and-module-architecture.md) | Backend HTTP/API & Module Architecture | Accepted | 2026-09-03 |
| [0003](0003-postgresql-persistence.md) | PostgreSQL Persistence | Accepted | 2026-09-03 |
| [0004](0004-authentication-and-account-isolation.md) | Authentication & Account Isolation | Accepted | 2026-09-03 |
| [0005](0005-time-and-timezone-model.md) | Time & Timezone Model | Accepted | 2026-09-03 |
| [0006](0006-frontend-architecture.md) | Frontend Architecture | Accepted | 2026-09-03 |
| [0007](0007-testing-and-local-development.md) | Testing & Local Development | Accepted | 2026-09-03 |
| [0009](0009-categories-as-a-shared-module.md) | Categories as a Shared Module | Accepted | 2026-09-04 |

ADR-0008 (production hosting / deployment) is intentionally not yet written — it
follows a free-tier validation spike.

ADR-0009 opens the design-driven expansion (`planning.md` Appendix A). Further ADRs
for that expansion (unified scheduled-item model, etc.) are written per milestone.
