# Specifications

A specification is the written output of the SDD stages that produce documents, for one
**milestone** (M0–M8, see `docs/roadmap.md`). It is created immediately before the
milestone is planned and implemented, so it stays current and high-signal. Milestone
specs are not written in advance.

## Layout

```
docs/specs/
├── templates/        spec.md, plan.md, acceptance.md skeletons
├── v1/               one folder per V1 milestone
│   └── M<n>-<slug>/
│       ├── spec.md        SPECIFICATION stage
│       ├── plan.md        PLAN stage
│       └── acceptance.md  ACCEPTANCE stage
└── v2/               created when V2 begins
```

One folder per milestone. If a milestone's specification grows too large to stay
high-signal, that is a signal to discuss splitting the milestone with the product
owner — not to invent a sub-numbering scheme. Milestone identifiers stay M0–M8.

## Lifecycle

`Draft` → `Approved` → `Implemented` → `Accepted`

- **Draft** — being written; open questions it consumes are still open.
- **Approved** — every open question the specification consumes has been resolved
  (recorded in `docs/requirements/v1.md`), and the product owner has approved
  `spec.md` and `plan.md`. Implementation may start.
- **Implemented** — implementation and automated verification are complete.
- **Accepted** — the implementation has been checked against the specification's
  acceptance criteria and signed off by the product owner.

## How a spec references requirements

`spec.md` names the requirement sections it covers, by number — e.g. `v1.md §3`,
`v1.md §6`, `v1.md N4`. Narrower clauses are cited by quotation. A spec may only
*cover* requirements;
it must not add, widen, or reinterpret them. If a milestone needs something not in
`v1.md`, that is a product decision and goes back to the product owner.

## Traceability

The chain, each link carried by an artifact that already exists — no separate
traceability database:

```
product requirement            docs/requirements/v1.md
  → requirement section          e.g. v1.md §3
    → milestone specification     docs/specs/v1/M<n>-<slug>/spec.md  (Covers: v1.md §3)
      → implementation            commit / PR message references the milestone (M<n>)
        → tests                   test name or comment references an acceptance criterion (AC-M<n>-04)
          → acceptance criteria   acceptance.md: each AC-M<n>-NN lists its requirement
                                  section and its proving test, ticked by the product owner
```

`AC-M<n>-NN` is a criterion identifier scoped to a milestone, not a competing
numbering scheme.
