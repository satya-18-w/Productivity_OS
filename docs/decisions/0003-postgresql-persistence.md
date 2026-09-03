# 0003 — PostgreSQL Persistence

**Status:** Accepted — 2026-09-03
**Applies to:** V1 (and beyond unless superseded)

## Context

PostgreSQL is the only datastore (N1). Before M1, the project needs a fixed driver,
query-code approach, migration tool, and transaction model. Constraints: E3 (boring
tech, justify dependencies, no ORMs that hide SQL), E4 (forward-only migrations); N3
(every query account-scoped — mechanism in ADR-0004); N4 (query shapes defined by
ADR-0005). Sits inside ADR-0001 / ADR-0002.

## Decision

- **Driver:** `jackc/pgx/v5`, used via `pgxpool` (its native API), not through
  `database/sql`.
- **Query code:** `sqlc` generates type-safe Go from hand-written SQL. SQL is explicit
  and lives in versioned `.sql` files. **No ORM.** `sqlx` is not adopted. A query that
  is genuinely dynamic and awkward in `sqlc` may use raw `pgx` — but only after
  evaluating that specific use case, not as a blanket fallback.
- **Migrations:** `golang-migrate`, plain `NNNN_name.up.sql` / `.down.sql` files in a
  single `db/migrations/` directory, globally ordered. **Forward-only:** `.down.sql`
  exists for local convenience; a mistake in a shipped migration is corrected by a new
  forward migration, never by editing the old one. On deploy, migrations run as a
  one-shot step before the server starts.
- **Transactions:** one transaction per writing use-case, opened and owned by the
  service layer. Store methods accept a `DBTX` value satisfied by both the pool and a
  transaction. Read-only paths use the pool directly.
- Migrations are the single source of schema truth; `sqlc` reads the migrated schema.

## Alternatives considered

- **`database/sql` + pgx stdlib adapter.** Keeps the swappable standard interface.
  Rejected: N1 fixes PostgreSQL permanently, so portability buys nothing, and native
  `pgxpool` gives better types and context handling.
- **`database/sql` + `lib/pq`.** Rejected: `lib/pq` is in maintenance mode.
- **An ORM (GORM, ent, Bun).** Rejected by E3 — hides the SQL and adds a query-building
  layer to learn and debug around.
- **Raw `pgx` everywhere, no code generation.** Maximum control, but repetitive row
  scanning that is easy to get subtly wrong across many queries. `sqlc` keeps the SQL
  fully visible while removing the rote scanning.
- **`sqlx` as the approach or a pre-declared fallback.** Trims scanning boilerplate but
  gives weaker guarantees than `sqlc` and still hand-rolls scanning. Not adopted; raw
  `pgx` covers the rare dynamic case.
- **`goose` / `atlas` for migrations.** `goose` is near-equivalent; `atlas` is a larger
  declarative tool with a paid tier. `golang-migrate` is already on the machine and its
  model matches forward-only.
- **Autocommit / per-request transaction middleware.** Autocommit breaks multi-
  statement use-cases; middleware hides control flow (E3). Explicit per-use-case
  transactions in the service layer are clearer and testable.

## Consequences

- One well-maintained driver dependency (`pgx`) and one dev-time code generator
  (`sqlc`, no runtime dependency).
- `sqlc` adds a code-generation step to the build and to CI; a `sqlc diff` check keeps
  generated code honest (ADR-0007).
- All SQL is reviewable in one place and versioned alongside the migrations that define
  the schema it targets.
- The `DBTX` seam is a small amount of plumbing that makes transaction boundaries
  explicit and lets tests run against real transactions.
- Forward-only migrations keep the production schema history linear and reproducible;
  there is no production "rollback" story, by design.

## Revisit conditions

- Report queries (M7) or another feature need dynamic filtering that `sqlc` handles
  poorly → evaluate that specific case (raw `pgx` for those queries, or reconsider
  `sqlx`) via a documented decision.
- `sqlc` project friction (build speed, feature gaps) outweighs its safety benefit →
  reconsider the query-code approach.
- A migration-tooling limitation appears → reconsider `goose` / `atlas` (migration SQL
  files are portable).

## Related documents

- `docs/product/principles.md` — E3, E4
- `docs/requirements/v1.md` — N1, N3, N4
- `docs/architecture/conventions.md` — query-file layout, naming, `DBTX` usage
- ADR-0002, ADR-0004, ADR-0005, ADR-0007
