# Engineering Conventions

> Living document. It records the conventions actually in force. Conventions not yet
> decided are marked as such and are settled by the milestone that first needs them —
> recorded here, and in an ADR if the choice has lasting consequences.

## In force now (from the approved product baseline)

- **Spec before implementation.** No milestone is implemented without an approved
  `spec.md` and `plan.md` (`docs/product/principles.md` E4).
- **Forward-only migrations.** Schema changes only move forward. A mistake is corrected
  by a new migration, never by editing one that has shipped (E4).
- **Account scoping is mandatory.** Every query is scoped to the authenticated account
  (`docs/requirements/v1.md` N3).
- **Unambiguous time.** Instants are stored without ambiguity; dates and weeks are
  resolved in the account's timezone (`v1.md` N4).
- **Deterministic V1.** No heuristic or AI-driven behaviour in V1 (P6).
- **Justify dependencies.** Every dependency beyond the standard library and PostgreSQL
  is justified in the `plan.md` that introduces it; a lasting one also gets an ADR (E3).

## To be established in Milestone M1

Not yet decided. Filled in here when M1 is planned:

- Go module and package layout
- Error handling and wrapping idiom
- Logging approach
- HTTP handler and routing structure
- Test layout, naming, and how to run a single test
- Formatting and linting toolchain, and how it is run

## Commands

To be added when the build / test / lint toolchain exists (Milestone M1).
