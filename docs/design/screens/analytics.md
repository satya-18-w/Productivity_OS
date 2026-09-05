# Screen — Analytics

**Reference:** `docs/design/references/analytics.png` (also panel 8 of `overall.png`)
**Purpose:** Charts and aggregates over a chosen date range.
**Proposed route:** `/analytics` (or `/reports`)

---

## V1 scope alignment

V1 **does** have a reporting surface — `requirements` §13 — but it is a **fixed,
exhaustive list of five deterministic reports**, not an open analytics dashboard.

**The V1 report set (§13), over a user-chosen date range:**
1. **Time by category** — total actual time per category.
2. **Planned vs actual by category** — planned, actual, and difference per category.
3. **Habit completion** — per habit: completed days + completion rate across the range.
4. **Task throughput** — count of tasks that entered `DONE` within the range.
5. **Daily actual totals** — total actual time for each day in the range.

§13 scope boundary: *"This list is exhaustive for V1. No custom report builder, no
arbitrary group-by, no saved reports, no comparison between two ranges, no trend lines or
forecasts, no correlation between metrics."*

| Reference element | V1 status |
|---|---|
| Date-range picker | **in scope** (§13 — user-chosen range) |
| Tabs: Overview / Productivity / Habits / Tasks / Goals / Time / Categories | over-structured vs 5 fixed reports; a single page or light grouping is enough |
| KPI row with **"+12% vs last month"** deltas | ⚠ **not V1** — "no comparison between two ranges" |
| "Productivity Trend" combo bar+line ("over the last 30 days") | ⚠ **not V1** — "no trend lines"; also "Focus Time" isn't a V1 measure. The nearest V1 chart is **report 5 (daily actual totals)** as a bar chart. |
| "Time Distribution" donut (time per category) | ✅ = **report 1** (time by category) |
| "Habit Consistency" heatmap | ✅ can render **report 3** (habit completion) though a heatmap is one interpretation; a per-habit "X days / Y%%" table is the literal form |
| "Top Categories by tasks completed" | ⚠ tasks have no category in V1 → not possible. "Top categories by actual time" (= report 1 sorted) is the V1 version. |
| "Goal Progress" donut / "Productivity Streak" / "Insights" cards | **not V1** — insights are heuristic (P6); goal progress has no % (§10); streak belongs to Habits |
| "Export Report" button | export in V1 is a **single full-data snapshot** (§14), not a per-report export |

**Recommendation:** build a **Reports** screen presenting exactly the five §13 reports for
a chosen date range. **The visualisation for each report is PENDING the Reports
specification** (`design-system.md` register item R1) — the pairings below are *candidates*
for that spec to confirm, not decisions:
- Report 1 (time per category, incl. explicit "Uncategorized" — Q8) → donut *or*
  horizontal bar.
- Report 2 (planned / actual / diff per category) → table *or* grouped bar; the existing
  `table.totals` with `.pos`/`.neg` (`web/src/styles.css`) is a ready fit.
- Report 3 (habit completion) → table: habit · completed days · completion rate.
- Report 4 (task throughput) → a single stat card.
- Report 5 (daily actual totals) → vertical bar chart, one bar per day.

**Drop (ratified — `design-system.md` §6.4):** period-over-period deltas, trend lines,
focus-time, goal analytics, insights, streak widgets, the 7-tab structure, per-report
export.

---

## Layout

- Shell (§4.1). Header: title icon badge + H1 "Reports" (or "Analytics") + subtitle;
  header illustration.
- **Toolbar:** a **date-range picker** pill (the one required control). Optional light
  section nav (Time / Habits / Tasks) — not the 7-tab set.
- **Body** (card grid, §4.7): one card per report (see recommendation). Each card: H3
  title + one-line description + the chart/table + a plain caption of the underlying
  numbers (P3 — the figures are the point, the chart is secondary).
- Right rail: minimal (a quote card or omit). No insights / streak / goal widgets.

## Screen-specific components

- **Date-range picker** — pill trigger → a range popover (inferred; §4 create/edit
  patterns apply). Default range TBD (last 7 / 30 days / this month — confirm).
- **Report card** — card + a data-viz primitive (§4.12) + a small figures table/caption.
- Reuse **`table.totals`** (`web/src/styles.css`) for report 2.
- Follow the **`dataviz` skill** for every chart (palette, light/dark, legends, a11y).

## Interactions

- Change the date range → all reports recompute (deterministically; §13, N4 — range may
  cross a DST boundary and must still total correctly).
- No drill-down, no saved views, no range-vs-range.

## Responsive

- Report cards 2-up → 1-up. Charts keep a min height; wide tables scroll in-container.
- Right rail drops first.

## Cannot be inferred / ambiguous

- Default date range and the allowed presets.
- Whether "Time by category" is shown as donut, bars, or both.
- Whether report 3's "completion rate" denominator is calendar days in range or days the
  habit was active (define in the reports SPEC).
- Chart interactivity (hover tooltips assumed; no more).

## Design-system references

§4.1 shell · §4.2 header · §4.7 card · §4.12 data-viz primitives · existing `table.totals`
in `web/src/styles.css` · `dataviz` skill · `requirements` §13 (+ Q7, Q8, Q10), N4 ·
`visual-principles.md` VP3, VP10.

---

## Phase 9 — Reports — Status: ✅ COMPLETE (2026-09-04)

Route `/reports` → `ReportsScreen` (`web/src/features/reports/`). **Backend: none exists**
— the entire screen runs on deterministic mock data pending the reports API (full spec in
`docs/left.md`, "Phase 9 — Reports").

**R1 resolved** (design-system.md §6.2, now moved to §6.1) — the five reports render as:
1. Time by category → **horizontal bars**, one per category incl. explicit "Uncategorized"
   (Q8), coloured via the existing `categoryColor()` palette, literal duration always
   shown as text (never bar-only), plus a total-time caption.
2. Planned vs actual by category → **table**, reusing `table.totals`/`.pos`/`.neg` exactly
   as Timeline's comparison card does, with a totals row.
3. Habit completion → **table + `ProgressBar`** per habit (a legitimate progress-bar use,
   unlike Goals — §13 explicitly wants a completion rate here), "`completed / range days
   (rate%)`" as literal text alongside the bar.
4. Task throughput → a single **`StatCard`** (count of tasks that entered `DONE` in range).
5. Daily actual totals → **vertical CSS bar chart**, one bar per day, horizontally
   scrollable for long ranges, native `title` tooltips carrying the literal value.

No charting library added — all five are plain HTML/CSS per the `dataviz` skill's
"table/stat-tile first, simplest mark that does the job" guidance and P3 (figures are the
point, charts secondary).

- [x] `PageHeader` + `DateRangePicker` (From/To date inputs + Last 7/30/90-day presets),
      range read/written to `?from=&to=` so a link to a specific range is shareable.
      Default range: trailing 30 days (the "Cannot be inferred" question above — resolved
      as part of this phase; no persisted last-used range, no "this month" preset per §13
      minimalism).
- [x] A persistent **"⚠ Sample data"** notice (`role="note"`) while the mock stands in.
- [x] **Dropped per spec:** donuts, the 7-tab structure, period-over-period deltas, trend
      lines, insights/streak/goal widgets, per-report export — none are V1 (§13, §6.4).
- [x] No rail (minimal, per D6/VP3 — same call as every other Phase 6+ screen).
- [x] Responsive — `reports-grid` 2-up → 1-up; the daily-totals bar chart scrolls in its
      own container rather than the page. Light + dark verified.
- [x] Tests (25) — `reportsData` (determinism, inclusive date range, Uncategorized bucket,
      planned-vs-actual excludes it, `differenceSeconds` math, habit `rangeDays`, default
      30-day range), one test file per report component, `ReportsScreen` (all 5 headings +
      sample-data notice render, URL param read/default/write). Full suite green.
- [x] Browser-verified — date range + presets, all 5 reports with plausible mock figures,
      category colours match the palette used elsewhere, dark, mobile (no h-scroll, cards
      stack, bar chart scrolls), no console errors.
- [ ] Committed — pending product owner.

`App.tsx`'s `/reports` route now renders `ReportsScreen` in place of the Phase-9
`<Placeholder>`.
