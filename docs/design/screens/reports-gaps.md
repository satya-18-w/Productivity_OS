# Reports — reference vs actual gap plan (2026-09-04)

Source: `docs/design/screens/analytics.md` (V1 scope alignment + R1) +
`docs/requirements/v1.md` §13 vs live screenshot `/tmp/opencode/pg-reports.png`.
Reference: `docs/design/references/analytics.png`. Design system:
`design-system.md` §3 tokens · §4.2 header · §4.6 Stat card · §4.7 card ·
§4.11 progress bar · §4.12 data-viz · §5 rules · §6.1 R1 · §6.4 exclusions.

V1 scope (fixed): exactly 5 deterministic reports over a user-chosen date
range — time-by-category hbars, planned-vs-actual table, habit completion
table + `ProgressBar`, task throughput `StatCard`, daily actual totals vbar —
plus a date-range picker and a "sample data" notice (no reports backend
exists; `docs/left.md` Phase 9). Excluded: trend lines, period-over-period
deltas, insights, heatmaps (as required output), focus-time, per-report
export. R1 visualisation decisions are settled — kept as-is.

## 1. Deviations from the reference KEPT (v1.md-justified, do not build)

| Reference element | Actual | Justification |
|---|---|---|
| 7 tabs (Overview / Productivity / Habits / Tasks / Goals / Time / Categories) | Single page, no tabs | §13 list is exhaustive; `analytics.md`: 7-tab set is "over-structured vs 5 fixed reports" |
| KPI row with "+12% vs last month" deltas | Absent | §13: "no comparison between two ranges" (§6.4; DS §4.6 builds number + label only) |
| "Productivity Trend" combo bar + line, "Focused Time" KPI | Daily actual totals vbar (report 5) | §13: "no trend lines"; focus time is not a V1 measure (§6.4) |
| "Time Distribution" donut | Horizontal bars incl. explicit "Uncategorized", `categoryColor()` palette, literal duration + total caption | R1 approved (DS §6.1); donut dropped per spec |
| "Habit Consistency" heatmap | Table + `ProgressBar` per habit, "completed / range days (rate%)" literal | R1 approved; the table is §13's literal form |
| "Top Categories by tasks completed" | Absent (report 1 sorted is the V1 version) | §13 list is exhaustive — no such report |
| "Goal Progress" donut, "Productivity Streak", "Insights" cards | Absent | §10: no % goal progress; streak belongs to Habits (§9); insights are heuristic P6 (§6.4) |
| "Export Report" button | Absent | §14: export is a single full-data snapshot, not per-report |
| Header illustration, quote/banner art, right rail | Omitted (no rail) | D6-bounded decoration; `PageHeader` deliberately omits illustration; spec allows omitting the rail |
| Date-range pill + popover | Native From/To date inputs + Last 7/30/90-day presets, range in `?from=&to=` (shareable), trailing-30-day default | §13 requires only "a user-chosen date range"; inputs are the minimal accessible form and resolve the spec's "TBD default" question with no "this month" preset (§13 minimalism) |
| Card headings | `H2` under the `H1` (spec text says H3) | Correct section hierarchy under `PageHeader`; `Card` default is h3, h2 chosen explicitly |
| Task throughput rendering | Shared productivity `StatCard` (number + label, no delta/spark) inside `Card` | DS §4.6 ratified form; reuse, not a second visual system |

## 2. Gaps to build (this pass — frontend scope only)

1. **Planned-vs-actual totals-row diff has no `pos`/`neg` class**
   (`PlannedVsActualReport.tsx`). Body rows colour the difference via
   `table.totals` `.pos`/`.neg` (the Timeline comparison-card pattern); the
   footer total renders plain. Colour the footer diff the same way.
2. **Missing one-line descriptions / captions.** `analytics.md` Layout wants
   title + one-line description + chart/table + plain caption (P3: figures are
   the point). Only report 1 has one. Add `report-caption` description lines
   to reports 2–4 and fold the range total into report 5's lead line
   (`"…{total} across {n} days"`, mirroring report 1's `"…{total} overall"`).
3. **`TimeByCategoryReport` has no empty state.** All-zero/empty rows render
   bare bars. Add "No actual time in this range." (matches the other cards'
   empty states).
4. **`DailyActualTotalsReport` polish.** Empty copy "No range selected." is
   wrong (a range is always present) → "No actual time in this range." Bars
   carry the literal only in hover `title`; add `role="img"` + `aria-label`
   with the literal value per bar (keeps the R1 native tooltips).
5. **`DateRangePicker` To cap contradicts Q9.** `max={todayISO()}` on the To
   input blocks future dates, but Q9 resolved that past-tense data may be
   recorded for future dates (no "not after today" check). Remove the cap
   (`From max={to}` / `To min={from}` stay as picker guardrails).
6. **Inverted range (`from > to`) renders incoherent reports**
   (`ReportsScreen.tsx`). Typed input can invert the range: daily totals go
   empty while the other mocks still show figures. Normalise (swap) on read
   and on write so URL, inputs, and data stay canonical.
7. **`reports-grid` allows 3-up on wide; spec says 2-up → 1-up.**
   At 3-up the 360px-min `table.totals` clips mid-column in the screenshot
   ("DIFFE" header, cut-off difference values). Constrain to 1 column base +
   2 columns ≥ 1024px (the `laptop` value from `breakpoints.ts` mirrored as a
   literal — the established `habits.css` pattern, not a new token).

Tests (colocated, extended): totals-row `pos`/`neg`; TimeByCategory empty;
Daily empty copy + per-bar `aria-label` + totals caption; To input has no
`max`; `ReportsScreen` from > to normalisation. `mockReportData` swap points
stay intact (`reportsData.ts` untouched).

## 3. Backend gaps noticed (list only — not implemented)

- `GET /api/reports?from=<ISO>&to=<ISO>` missing — the entire screen runs on
  `mockReportData` (swap point: `reportsData.ts` → call site in
  `ReportsScreen.tsx`; remove the "⚠ Sample data" notice with it). Full shape
  in `docs/left.md` Phase 9.
- `habit_completion.range_days` semantics: the mock uses full range length;
  the backend must return days the habit was *active* (`left.md` note).
- Task `DONE` re-entry counting is still Q10-partly-open (transition-time
  recording confirmed at M7).
- N4: resolve the range in the account timezone (DST-correct totals),
  server-side bound (e.g. ≤ 366 days, 400 beyond).
