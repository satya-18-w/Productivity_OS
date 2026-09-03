# Architecture Overview

> Living document. It describes how Productivity OS is built as it stands today.
> Decisions with lasting consequences are recorded in `docs/decisions/`; the change
> history is git. When this document and an ADR disagree, the ADR is the record of
> intent and this document is out of date and should be fixed.

## Architectural style

Productivity OS is a **modular monolith**: one deployable Go binary, backed by a single
PostgreSQL database, serving a modern web frontend. No microservices, no message
broker, no cache, no second datastore. This holds until a *demonstrated* need is
recorded in an ADR (`docs/product/principles.md`, E2).

## The dependency rule

Modules are separated by explicit, published interfaces. A module may depend on another
module only through that interface — never by importing its internal packages, reading
its tables, or sharing its types. Modules are packages in one process; there is no
network hop between them.

A violation of this rule is a defect, not a style preference.

## Modules

The module catalogue is populated as milestones deliver modules; it is not designed up
front. When the first modules land (Milestone M1) they are listed here, or this section
points to `docs/architecture/modules.md` once that file has become necessary.

Currently: none.

## Cross-cutting concerns

These apply to every module and are specified in detail by the milestone that first
needs them:

- **Account scoping.** Every read and write is scoped to the authenticated account
  (`docs/requirements/v1.md` N3). No code path returns or mutates another account's
  data.
- **Time and timezone.** Instants are stored unambiguously; "a date" and "a week" are
  always resolved in the account's timezone (`v1.md` N4).
- **Determinism.** No V1 behaviour depends on AI, prediction, or heuristics
  (`docs/product/principles.md` P6).

## Decided vs pending

**Decided (product baseline):** modular monolith, Go, PostgreSQL, single binary,
responsive web frontend, ₹0/month hosting target, deterministic V1.

**Pending — to be settled by ADR before or during Milestone M1:** Go toolchain
version, frontend framework, database access approach, local development
infrastructure, schema migration tooling, deployment / hosting target, CI approach.
These are opened for review as a set; not every one necessarily becomes an ADR.

## Related documents

- `docs/product/vision.md`, `docs/product/principles.md` — why, and the decision filters
- `docs/requirements/v1.md` — what V1 must do
- `docs/roadmap.md` — milestone order (M0–M8)
- `docs/decisions/` — architectural decisions
- `docs/specs/` — per-milestone specifications
