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
| —   | none yet | — | — |
