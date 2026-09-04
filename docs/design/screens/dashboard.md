# Screen — Dashboard

**Reference:** `docs/design/references/dashboard.png` (also panel 1 of `overall.png`)
**Purpose:** A single landing overview: today's schedule, quick stats, recent items across
areas.
**Proposed route:** `/` (currently `/` renders Timeline).

---

## V1 scope alignment

⚠ **This entire screen is outside approved V1 scope.** `docs/requirements/v1.md` defines
no dashboard/home aggregate; it lists 14 discrete capability areas and a fixed report set.
Nearly every widget here is post-V1:

| Widget in reference | V1 status |
|---|---|
| "Good morning, Satyajit" greeting + rotating quote | post-V1 (P3 tension) |
| KPI row: Tasks 5/8, Habits 3/5, Focus Time 4h20m /8h goal, Goals 2/4 | Tasks/Habits counts *derivable* from V1 data; **Focus Time** and **time goals** are not V1; **Goals "on track"** is not a V1 state |
| Today's Schedule (timeline preview) | a preview of V1 timeline data — plausible |
| Quick Actions (Add Task / Add Note / Set Goal / Start Focus) | Notes & Focus are not V1 |
| Recent Tasks list | derivable from V1 tasks |
| Habit Tracker week strip | derivable from V1 habit completions |
| Goals Progress with % and "3/7" | V1 goals have **no %** — 4 manual states only |
| Focus Mode / Pomodoro card | **not V1** (V2 candidate) |
| This Week day strip with task counts | derivable |
| Mini month calendar | plausible helper |
| Notes list | **Notes are not a V1 feature at all** |

**Ratified 2026-09-04: Dashboard is reference-only and MUST NOT be implemented**
(`design-system.md` §6.4). There is no V1 home/overview screen. If a lightweight "Today"
screen is later wanted it would need a requirements entry first; it could then reuse the
timeline preview, today's habit toggles, today's tasks, and the goal list (state chips,
no %). Nothing here is built in the meantime.

---

## Layout

- Shell: `design-system.md` §4.1 (sidebar + top bar + right rail).
- Header: date line + greeting H1 + quote subtitle; header illustration bleeds top-right.
- Main column, vertically stacked:
  1. **KPI row** — 4 stat cards (§4.6), one per area, each tinted (green/rose/blue/violet).
  2. Two-column region: left **"Today's Schedule"** card (timeline row list, each row =
     time + title + sub + completion dot); right **"Quick Actions"** 2×2 icon-button grid,
     then **"Recent Tasks"** card, then **"Habit Tracker"** card (week toggle grid).
  3. **"Goals Progress"** card — progress-bar rows (§4.11).
  4. **"Focus Mode"** dark card — post-V1.
  5. **"This Week"** 7-day strip.
- Right rail: mini month calendar (§4.13), a quote card, **"Today's Focus"** category-time
  list (post-V1), a Notes list (post-V1).

## Screen-specific components

- **Quick-action tile** — large icon button, soft tint, label under icon. (new)
- **Schedule preview row** — reduced timeline row: no tags, no kebab, just time + title +
  sub + done dot.
- **Day-strip cell** — weekday + date + count, today emphasised.

## Interactions (inferred)

- KPI card → navigates to that area. "View all →" links → area screens. Mini-calendar day
  → Timeline on that date. Habit toggle → marks completion (same control as Habits grid).

## Responsive

Right rail drops first; then the two-column region stacks; KPI row wraps 4→2→1.

## Cannot be inferred / ambiguous

- Whether the greeting/quote is fixed or rotates; source of quotes.
- "Focus Time … of 8h goal" — no such goal object exists in V1.
- Ordering / which widgets are configurable.

## Design-system references

§3.6 shell grid · §4.1 shell · §4.2 header · §4.6 KPI card · §4.7 card · §4.11 progress
bar · §4.13 mini calendar · §4.14 quote card · `visual-principles.md` VP3, VP10.
