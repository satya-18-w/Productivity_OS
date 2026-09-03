# Productivity OS — Roadmap

> Ordered by dependency and value. **No dates, no resourcing.** Each milestone is a
> vertical slice that can be demonstrated on its own. A milestone is complete only when
> it has passed the full workflow through ACCEPTANCE.
>
> Section references (§1–§14, N1–N7, Q1–Q11) point at `docs/requirements/v1.md`.

## Milestone 0 — Foundation

**Goal:** the project can be developed under Spec-Driven Development.

- Product documents (vision, principles, V1 requirements, this roadmap) reviewed and
  approved.
- `CLAUDE.md` updated to defer to these documents.
- `docs/` structure for feature SPECs and ADRs agreed.
- Foundational ADRs: Go toolchain version, frontend framework, database access
  approach, local development infrastructure (Docker Compose), migration tooling,
  deployment / hosting target.
- Repository skeleton: module layout, build / test / lint entrypoints, CI decision.
- No application behaviour.

**Depends on:** nothing. **Unblocks:** everything.

## V1 milestones

### M1 — Skeleton and authentication

**Delivers:** a deployed, reachable application. A user can create an account, log in,
log out, set their timezone, and change their password. Account-scoped data isolation
is enforced from the first table.

**Covers:** §1, N1, N3, N5 (shell).

**Depends on:** M0.

**Why first:** every other feature stores per-account data and needs the isolation
boundary and the account timezone in place before it can be built correctly.

### M2 — Categories and the timeline core

**Delivers:** the user can manage categories; add, edit, and delete planned blocks and
actual blocks for a date, including blocks that cross midnight; view a single date's
timeline with planned and actual shown together; and see per-date planned-vs-actual
totals by category.

**Covers:** §2, §3, §4, §5, §6, N4.

**Depends on:** M1.

**Why here:** this is the product's core. Everything after it is either a different
lens on time or a supporting list.

**Blocked by open questions:** Q9 — whether actual time blocks may be recorded for
dates that have not yet occurred.

### M3 — Tasks and the Kanban board

**Delivers:** the user can create, edit, and delete tasks and move them across the four
fixed columns; the board view.

**Covers:** §7, §8.

**Depends on:** M1. Independent of M2; sequenced after it by value.

### M4 — Habits and streaks

**Delivers:** the user can create and archive habits, mark completion per date, and see
the current streak for each active habit.

**Covers:** §9.

**Depends on:** M1.

**Blocked by open questions:** Q9 — whether habit completions may be marked for dates
that have not yet occurred.

### M5 — Goals

**Delivers:** the user can create, edit, and delete goals and set a manual progress
state.

**Covers:** §10.

**Depends on:** M1.

**Note:** the smallest milestone, deliberately minimal per the locked V1 decision.

### M6 — Daily and weekly reviews

**Delivers:** the user can complete, edit, and view daily reviews (per date) and weekly
reviews (per ISO week), each against a fixed prompt set, with that period's totals
shown for reference.

**Covers:** §11, §12.

**Depends on:** M2 (time totals), M4 (habit completion data), and — for the weekly
review's task figure — M3.

**Blocked by open questions:** Q1, Q2, Q9, Q10. Q9 covers whether daily and weekly
reviews may be completed for dates that have not yet occurred. Q10 is required because
the weekly review's "tasks that entered `DONE` that week" figure depends on how a
`DONE` transition is defined.

### M7 — Reports

**Delivers:** the five fixed reports over a user-chosen date range.

**Covers:** §13.

**Depends on:** M2, M3, M4 — each report aggregates one of those areas.

**Blocked by open questions:** Q7, Q8, Q10.

### M8 — Data export

**Delivers:** the user can export all of their data in one open, documented format.

**Covers:** §14, P5.

**Depends on:** every V1 entity existing, so it comes last.

**Blocked by open questions:** Q3.

### V1 Definition of Done

- Every capability in §1–§14 has passed ACCEPTANCE.
- All V1 open questions (Q1–Q11) are resolved and reflected in the relevant SPECs.
- N1–N7 are verified: single binary; runs within free tiers; isolation tested;
  timezone, midnight-spanning, and DST cases covered; responsive on desktop and mobile
  viewports; a database restore performed successfully; interactive views feel
  immediate.
- No item from the V1 non-goals list is present in the product.
- Export round-trips: the exported file contains every entity the user created.

## V2 — Assisted

Theme: reduce manual effort with reusable structure. Still deterministic, still no AI,
still no social features.

- **Recurring tasks** and the recurrence engine deferred from V1.
- **Day-plan templates** and "copy a previous day".
- **External calendar import** (e.g. Google Calendar) into planned or actual blocks.
- **Live timer capture** for actual blocks (candidate).
- **Expanded analytics:** longer trends, week-over-week comparison, per-category time
  budgets and progress against them.
- **Account-lifecycle gaps** from V1: password-reset email, email verification,
  self-service account deletion (candidates).

## V3 — Intelligent

Theme: AI-assisted planning and insight. The user accepts or rejects every suggestion;
the system never acts unattended.

- **Suggested day plans** from history and stated goals.
- **Pattern detection:** observations about when and how the user works, phrased as
  findings, not instructions.
- **Productivity insights** linking time, habits, tasks, and goals.
- **Natural-language capture and review assistance.**

## What this roadmap deliberately does not do

- No milestone introduces a V2 or V3 capability early. M6 uses only the fixed review
  prompt sets from V1 requirements; M7 produces only the five fixed reports.
- No milestone adds recurrence, templates, calendar import, timers, or AI.
- Milestones are not dated and not resourced here. Sequencing is by dependency first,
  then value.
