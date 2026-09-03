# Engineering Conventions

> Living document. It records idiomatic implementation guidance — how to write code
> that fits this codebase. Architectural decisions and their rationale live in
> `docs/decisions/` (ADRs); this file points to them rather than repeating them.

## Where the rules live

| Topic | Owned by |
|-------|----------|
| Deployment shape, single-origin, no distributed infrastructure | ADR-0001 |
| Go toolchain, module/package structure, the dependency rule, HTTP layer, API & error contract, validation stance | ADR-0002 |
| pgx driver, `sqlc`, migrations (forward-only), transaction boundaries | ADR-0003 |
| Sessions, password-hashing algorithm, account-scoping mechanism, no RLS in V1 | ADR-0004 |
| Timezone storage and all date / week bucketing | ADR-0005 |
| Frontend stack and single-origin asset serving | ADR-0006 |
| Local dev topology, testing approach, CI | ADR-0007 |
| Spec-before-code, the eight-stage workflow | `CLAUDE.md`; principle E4 |
| ₹0 filter, "boring tech", justify every dependency | `docs/product/principles.md` E1, E3 |

## Idioms to be established in Milestone M1

Filled in here when the M1 skeleton lands. These are style choices *within* the ADRs,
not new decisions:

- Error creation and wrapping idiom — how a `code` / `message` / `fields` error is
  built and propagated to the central writer defined in ADR-0002.
- Structured logging (`log/slog`): field names, levels, what every request logs.
- Naming: packages, exported module interfaces, `sqlc` query files and methods.
- Internal module layering in practice (`service` / `store` / handler) — the shape
  ADR-0002 leaves to convention.
- Test file layout, test naming, and how to run a single test (approach set by
  ADR-0007).
- `gofmt` / `golangci-lint` configuration and how it is run locally.
- Frontend: component structure, styling approach, and any libraries added (per
  ADR-0006, only on demonstrated need).

## Commands

To be added when the M1 `Makefile` exists (see ADR-0007 for the intended entry points).
