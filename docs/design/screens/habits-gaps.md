# Habits — reference vs actual gap plan (2026-09-04)

Source: `docs/design/screens/habits.md` (V1 scope alignment) + `docs/requirements/v1.md` §9
vs live screenshot `/tmp/opencode/pg-habits.png` (Today view).
Reference: `docs/design/references/habits.png`. Design system: `design-system.md`
§3 tokens · §4.2 header · §4.3 view switcher · §4.5 buttons · §4.6 KPI card ·
§4.7 card · §4.10 toggle-circle · §4.16 form (+ §5 rules, VP3/D6).

V1 scope (fixed): create (name only), mark/unmark per date, archive/unarchive
(history kept), current streak display, last-30 count.
Excluded (do not build): longest streak, consistency %, habit categories,
targets/icons, rename/delete (no backend).

## 1. Deviations from the reference KEPT (v1.md-justified, do not build)

| Reference element | Actual | Justification |
|---|---|---|
| Playful H1 "Small habits. Big results." + "Consistency today builds…" subtitle + header illustration | Factual `PageHeader`: eyebrow "Habits", title "Habits", subtitle "Your daily habits and current streaks."; no illustration | VP3/D6: no motivational copy adjacent to data; illustration is decorative-only, omitted when unsure |
| Split primary "＋ Add Habit ▾" | Plain primary "Add habit" (`Button`) opening `HabitDialog` | V1 create has one field (name, §9) — a split menu implies options that do not exist |
| KPI row of 4 (completed 5/7, current 12, longest 28, 71% consistency) with icons + sparklines | 3 V1-safe `StatCard`s: "Completed today N/M" · "Active habits" · "Best current streak … days"; no deltas/sparks | §9 scope boundary + DS §4.6: no longest-streak, no consistency % (§13 Reports), no period-over-period delta/spark |
| Habit sub-labels ("30 minutes", "At least 20 pages") | Name only everywhere | §9: "a habit has only a name" |
| Per-habit glyph tile (dumbbell, book, lotus…) | None (name-led rows) | V1 has no habit icon/target/category field; a tile would invent data |
| "Actions" kebab with edit/delete | `HabitActions` offers Archive only (Unarchive lives in All view) | §9 grants create/mark/unmark/archive/unarchive/view — not rename/delete (`docs/left.md` Phase 6, item (c)) |
| Right rail: "Your Streak" week dots, "Habit Completion" bar chart, "Top Habits" completion-rate list | No rail | Completion rate over a range is Reports §13 ("Habit completion"); charts wait for R1 (`habits.md` Deferred) |
| "Habit Categories" section | Absent | Habits carry no category in V1 (core concepts; DS §6.4) |
| Motivational banner "Consistency compounds…" + quote cards | Absent | VP3/D6 violation — dropped per spec recommendation |
| Weekday header "Mon 31 Aug" (month on every column) | Weekday short + day number; month range lives in the week-nav label | Spec requires "short name + date; today emphasised" — satisfied; per-column month repeats add width for no V1 value |
| "Actions" column header text | Visually-hidden "Actions" (`ui-visually-hidden`) | Same meaning, more accessible; geometry unchanged |

No layout defects observed in the screenshot: toolbar wraps (`Flex`), KPI 3→1,
grids scroll in-container (sticky habit-name column), page never scrolls
sideways; `ToggleCircle` click defect already fixed (DS §4.10 note).

## 2. Gaps to build (this pass — frontend scope only)

1. **Last-30 count never rendered (G1 — the only functional V1 gap).**
   `GET /api/habits` returns `last_30_days` per habit and the task scope lists
   "last-30 count" as V1, but no view displays it (grep: only used as mock
   density in `habitData.ts`). Add a `Last30` presentational bit in
   `HabitBits.tsx` (feature-local, per scope rules — NOT promoted to
   `components/ui`), rendered as plain tabular `N/30` with a muted token style
   and an accessible name:
   - Today checklist: next to `Streak` per row.
   - Week grid: new "Last 30" column per row (keeps the 7 day columns + Streak untouched).
   - All habits: next to `Streak` on active rows.
   - Month heatmap: new "Last 30" column (reads the real `last_30_days`, not the mock cells).
   Styles in `web/src/styles/habits.css` only (`.habit-last30`), tokens only,
   Flex/inline flow, no absolute layout, no new deps. Tests: extend
   `HabitTodayList` + `HabitWeekGrid` suites, add `HabitAllList` suite for the
   new cell (all colocated `*.test.tsx`).

No other functional gaps: name-only create dialog, per-date toggle (incl.
future, Q9) with optimistic UI + refetch, archive/unarchive with history-kept
hint (Q11), plain-number streak with static flame (VP3), view switcher
(`?view=` + `?week=`), week nav, sample-data note on the mock month view, and
empty/error/loading states are all present.

## 3. Backend gaps noticed (list only — not implemented)

1. **`GET /api/habits/history?from=&to=` — (a) required for "This month".**
   Heatmap runs on `mockHabitHistory()` + visible sample-data note. Swap point:
   `habitData.ts` → `mockHabitHistory`. Tracked in `docs/left.md` Phase 6.
2. **`GET /api/habits/week?date=` — (b) optimisation.** Week grid makes 7
   parallel `GET /api/habits?date=` calls (`fetchWeek`). Swap point:
   `habitData.ts` → `fetchWeek`. Tracked in `docs/left.md` Phase 6.
3. **Rename / delete habit — (c) not V1.** No endpoint, no `api.ts` method;
   kebab correctly offers Archive only. Confirm before building. Tracked in
   `docs/left.md` Phase 6.
