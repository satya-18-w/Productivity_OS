# Goals — reference vs actual gap plan (2026-09-04)

Source: `docs/design/screens/goals.md` (V1 scope alignment) + `docs/requirements/v1.md` §10
vs live screenshot `/tmp/opencode/pg-goals.png`.
Reference: `docs/design/references/goals.png`. Design system: `design-system.md`
§4.2 header · §4.3 view switcher · §4.5 buttons · §4.6 KPI card · §4.9 status
chip · §4.12 donut · §4.16 form (+ §3 tokens, §5 rules).

V1 scope (fixed): title / description / target-date; 4 manual states
`NOT_STARTED` / `IN_PROGRESS` / `ACHIEVED` / `ABANDONED` via `StatusBadge`.
Excluded: % progress, progress bars, task counts/linkage, milestones,
categories, On-Track/At-Risk wording.

## 1. Deviations from the reference KEPT (v1.md-justified, do not build)

| Reference element | Actual | Justification |
|---|---|---|
| Progress bar + % per goal | Absent (asserted in `GoalRow` + `GoalsScreen` tests) | §10: "no percentage, no numeric target, no progress derived from tasks" |
| "12 / 20 tasks" sub-count | Absent | §10: "A goal is not linked to any other entity in V1" |
| "On Track / At Risk / Completed / Not Started" chips | `StatusBadge` with the four §10 labels verbatim | §10 + DS §4.9: On-Track/At-Risk imply a derived health signal V1 does not compute |
| Category chips on rows, category tabs, "Goal Categories" widget | Absent; filter is All + 4 states only | Goals carry no category in V1 (core concepts; DS §6.4) |
| KPI row "Total / On Track / At Risk / Completed" | "Total / In progress / Achieved / Not started" (`StatCard`s, no %) | Same §10 wording rule; 4th slot is Not started per spec's "(Not Started or Abandoned)" |
| Right-rail "Goal Progress" donut | Plain "By status" count list | Donut-by-state deferred — chart form waits for R1 (`goals.md` Deferred); counts match the Tasks/Habits rail pattern |
| Right-rail "Upcoming Milestones" | Absent | §10: "No milestones, key results, or check-in history" |
| Header illustration, quote cards, sidebar/rail art | Absent | Decorative only (VP3/D6); omitted, not required |
| Per-goal glyph tile | None (title-led rows) | Spec allows "a generic goal glyph or none" — V1 has no goal-icon field |
| View-switcher label "All Goals" | "All" | Trivial wording; consistent with the other list screens |
| Ordering | Newest-first (`filterGoals`) | Ordering unspecified in V1; matches the other list screens |

## 2. Gaps to build (this pass — frontend scope only)

1. **`goals.css` raw spacing values → tokens** (DS §5 rule 1; `tokens.css` is
   canonical). `gap: 5px` → `var(--sp-1)`; `padding: … 2px` horizontals →
   `var(--sp-1)`; `margin-top: 2px` → dropped (parent flex `gap` already
   separates); `gap: 4px` / `padding: 8px …` → `var(--sp-1)` / `var(--sp-2)`
   (same rendered values). No colour/radius/shadow changes needed — the file
   already uses tokens throughout; layout is already Grid/Flex, no absolute
   positioning.
2. **Filter test coverage**: assert the `SegmentedControl` offers exactly
   All + the four V1 states and no category options (locks in exclusion #4
   above). Added to `GoalsScreen.test.tsx`. No new test files; existing
   `GoalRow`/`GoalDialog`/`goalHelpers` coverage already asserts the other
   exclusions.

No functional gaps: `PageHeader` + New-goal primary, `GoalDialog`
(title/description/target-date only), inline state change via kebab
(`Set to <other 3 states>`), edit/delete, `?filter=` URL state, KPI strip,
rail counts, empty/error/loading states, and responsive KPI 4→2→1 are all
present and match the spec.

## 3. Backend gaps noticed (list only — not implemented)

None. `api.goals`, `createGoal`, `updateGoal`, `setGoalProgress`,
`deleteGoal` are all present in `web/src/api.ts` and wired through the
screen; no mock data, no missing endpoint observed from frontend scope.
