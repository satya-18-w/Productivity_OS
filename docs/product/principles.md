# Productivity OS — Principles

> These are decision filters. When more than one design or scope choice is defensible,
> apply these and pick the option that best satisfies them. Each principle states the
> rule, why it exists, and what it rules out. A principle that forbids nothing is not a
> principle.

## Product principles

### P1 — Capture friction is the enemy

**Rule:** Recording what happened must take seconds and must never block on planning,
categorizing, or reflecting.

**Why:** If logging is tedious the actuals data will be incomplete, and every report
built on it becomes misleading.

**Rules out:** Required fields on an actual block beyond its start and end; forcing a
category at capture time; multi-step wizards for routine entry.

### P2 — Plan and actual are both first-class

**Rule:** Planned blocks and actual blocks are stored separately. Neither is ever
silently overwritten or derived from the other.

**Why:** The gap between intent and reality is the core insight the product exists to
show.

**Rules out:** "Editing the plan into reality"; a single mutable schedule; a
"convert plan to actual" action; computing one set of blocks from the other.

### P3 — Honest over motivational

**Rule:** Show what the data says, including when it is unflattering. No encouragement
copy, no hidden bad weeks, no inflated streaks.

**Why:** The user asked for an accurate picture. Motivational distortion destroys trust
in the numbers.

**Rules out:** Gamification beyond a plain streak count; "you've got this" messaging;
rounding away missed days; celebratory nudges; streak freezes that misrepresent history.

### P4 — Personal, not private-social

**Rule:** Every feature is designed for one person looking at their own data.
Authentication and per-account isolation are infrastructure, never a product surface.

**Why:** The product's value is individual reflection. Any social affordance changes
the user's behaviour and the product's purpose.

**Rules out:** Share links, exports addressed to another person, "compare with",
visibility settings, anything that implies a second reader.

### P5 — The user owns their data

**Rule:** The user can export all of their data in an open, documented format at any
time, without assistance.

**Why:** A personal tool that traps your history is not trustworthy. Export is a V1
requirement, not a later addition.

**Rules out:** Proprietary-only storage with no export path; export gated behind a
paid tier or a support request; lossy export.

### P6 — Deterministic before intelligent

**Rule:** No V1 feature may depend on AI, prediction, or heuristics. Intelligence is
added later, on top of a system that already works without it.

**Why:** A manual system that is correct and fast is the foundation. AI layered on an
unreliable base amplifies the unreliability.

**Rules out:** "Smart" defaults that guess; auto-categorization; suggested schedules;
any output that cannot be explained by a rule the user could read.

## Engineering principles

### E1 — ₹0/month is a hard filter

**Rule:** Every infrastructure, hosting, and third-party choice must have a genuine
free tier that covers the expected load (order of tens of accounts, thousands of
records each).

**Why:** The project has a zero-cost operating target. A choice that costs money is
disqualified regardless of its other merits.

**Rules out:** Paid managed services with no free tier; usage-billed dependencies that
charge past a low threshold; anything that needs a card on file to keep running.

### E2 — Modular monolith, no distribution

**Rule:** One deployable Go binary. Modules are separated by explicit interfaces and do
not import each other's internal packages. No network hop between modules.

**Why:** The user base is small and the operational budget is one person's spare
attention. Distributed systems add failure modes with no matching benefit here.

**Rules out:** Microservices; a message broker between modules; separate deployables
for work the monolith could serve; adding a queue, cache, or search engine before a
measured need.

### E3 — Boring, well-understood technology

**Rule:** Prefer the option a competent Go developer recognizes immediately. Justify
every dependency beyond the standard library and PostgreSQL.

**Why:** Maintainability over years by a small team beats momentary convenience.

**Rules out:** Frameworks that hide the request lifecycle; ORMs that obscure the SQL;
niche libraries for problems the standard library already solves; trend-driven choices.

### E4 — Spec before code, forward-only migrations

**Rule:** Every feature has an approved SPEC and PLAN before implementation. Database
migrations only move forward; a mistake is corrected by a new migration.

**Why:** Spec-driven development is the project's chosen method. Forward-only
migrations keep the production schema history reproducible.

**Rules out:** Editing a migration that has shipped; implementing from a verbal
description; "we'll write the spec afterwards."

## How to apply these

When facing a choice:

1. Eliminate every option that violates a **hard** rule: E1, E2, E4, P2, P5, P6.
2. Among what remains, pick the option that best serves **P1** (capture friction) and
   **E3** (boring technology).
3. If still tied, pick the option that is **easiest to remove later**.

### Worked example

*Should V1 auto-suggest a category for an actual block from the time of day and past
behaviour?*

- **P6** rules it out immediately: it is a heuristic the user cannot read as a rule.
- Even setting P6 aside, **P1** is served better by making the category optional than
  by guessing it and making the user correct the guess.
- **Decision:** the category stays optional and manual in V1. Revisit at Horizon 2/3.
