# 0007 — Testing & Local Development

**Status:** Accepted — 2026-09-03
**Applies to:** V1 (and beyond unless superseded)

## Context

Before M1, the project needs a consistent way to build, run locally, test, and gate
changes. The defect classes that matter most here — SQL correctness, account isolation
(N3), and timezone / DST correctness (N4) — all live at the database boundary.
Constraints: E1 (free tooling), E3 (boring tooling); N6 (backups + a tested restore
before V1 is done). Production hosting, the production image, and the backup mechanism
are deferred (ADR-0008).

## Decision

**Local development.**
- Docker Compose runs **only PostgreSQL** (a pinned image, a named volume). Go and the
  frontend toolchain run natively on the host.
- A `Makefile` provides the entry points (`db-up`, `db-down`, `migrate`, `run`, `test`,
  `lint`, `check`, `sqlc`). Commands are also documented in
  `docs/architecture/conventions.md` and `CLAUDE.md` once they exist.
- Configuration via a git-ignored local `.env`, with a committed `.env.example`.

**Docker.** Used for local PostgreSQL only. The backend and frontend are not
containerized for development; the frontend is never containerized. A production image,
if needed, is decided with hosting (ADR-0008).

**Testing.**
- The standard library `testing` package, with `testify/require` for assertions.
- **Unit tests** for pure logic: streak calculation, planned-vs-actual arithmetic,
  timezone / date bucketing, validation, report aggregation.
- **Integration tests against a real PostgreSQL** — a dedicated test database in the
  local Compose instance, migrations applied, state reset between tests. Mocking the
  database is not the default; the risky logic is the SQL itself.
- **Cross-account isolation tests are mandatory** for every module that owns account
  data (verifying N3 — see ADR-0004).
- **HTTP-level tests** through the real router and a real database for critical paths
  (authentication, isolation).
- **Browser end-to-end tests are not part of V1 by default, and are not permanently
  excluded.** They are introduced for a milestone where acceptance genuinely benefits
  from real-browser verification.

**CI.**
- **GitHub Actions** (the repository is on GitHub; the free tier covers this project).
  One workflow runs on push and pull request: `go build`, `go vet`, `golangci-lint`,
  `go test` with a PostgreSQL service container, `sqlc diff` (fails if generated code
  is stale), and the frontend typecheck + build.
- **No deployment automation in V1.** CI verifies; it does not deploy.
- CI does not get its own ADR: the tool choice follows from the repository host, and
  the only positions worth recording — "integration tests run against a real database"
  and "CI does not deploy" — are captured here.

**Backups (N6).** A backup and at least one successfully tested restore are required
before V1 is considered done. The mechanism depends on the production database
(ADR-0008) and is deferred.

## Alternatives considered

- **Mock the database; few or no integration tests.** Faster tests, but they would not
  exercise the SQL, the isolation guarantees, or the timezone math — the areas most
  likely to break.
- **`testcontainers-go` from the start.** Fully hermetic per-test databases, at the
  cost of speed and requiring Docker in every test run. A shared local test database is
  simpler for now; `testcontainers` can be adopted later if run-to-run isolation
  becomes a problem.
- **Everything in Docker Compose for local dev (backend + frontend containers).**
  Slower edit-run loop, more configuration, and it hides the toolchain.
- **A different CI system.** No reason to leave GitHub-native tooling; the free tier is
  sufficient.
- **Establishing "no E2E in V1" as a fixed rule.** Rejected: E2E has real value for
  specific acceptance criteria; the decision is made per milestone.

## Consequences

- Fast local iteration (native Go / frontend) with a reproducible database.
- Integration tests are slower than pure unit tests but catch this project's actual
  failure modes; CI needs a PostgreSQL service and runs migrations.
- The `sqlc diff` gate keeps generated query code in lockstep with the SQL and schema.
- No automated deploys means releases are a deliberate manual action in V1 — acceptable
  and simpler; revisit once hosting is stable.
- The backup requirement is tracked here but cannot be fully satisfied until ADR-0008.

## Revisit conditions

- Test-run cross-contamination or parallelism needs → adopt `testcontainers-go`.
- The frontend grows enough logic to warrant its own unit tests → add a frontend test
  runner (milestone decision).
- Hosting is selected (ADR-0008) → decide the production image and the backup
  mechanism, and reconsider deploy automation.
- CI minutes or constraints change materially → revisit the pipeline.

## Related documents

- `docs/product/principles.md` — E1, E3
- `docs/requirements/v1.md` — N3, N4, N6
- `docs/architecture/conventions.md` — test layout, how to run a single test
- ADR-0002, ADR-0003, ADR-0004, ADR-0005; ADR-0008 (deferred — production hosting)
