# Screen — Habits

**Reference:** `docs/design/references/habits.png` (also panel 3 of `overall.png`)
**Purpose:** Track daily habits; mark completion per date; see current streaks.
**Proposed route:** `/habits`

---

## V1 scope alignment

Maps to `requirements` §9 (habits and streaks).

| Reference element | V1 status |
|---|---|
| Habit grid: name + 7 weekday toggle circles + streak | in scope. Daily cadence only; each date completed/not; streak = consecutive completed dates ending today/yesterday. |
| Habit sub-label ("30 minutes", "At least 20 pages") | ⚠ a habit has **only a name** in V1 (§9). The sub-label is not a V1 field. Could be folded into the name. |
| Streak column with flame + number | in scope — but show the **number plainly** (P3; VP3). Flame icon is borderline gamification; keep minimal or drop. |
| "Actions" kebab (archive/unarchive/edit/delete) | archive/unarchive in scope (§9); edit name plausible; delete not stated but harmless |
| Period switcher Today / This Week / This Month / All Habits | "Today" and "All Habits" map to §9 ("which habits completed for a chosen date"; "see active habits"). Week/Month expansions of the grid are reasonable renderings. |
| KPI row: Habits completed 5/7, Current streak 12, **Longest streak 28**, Weekly consistency 71% | "completed today" and "current streak" are V1. **Longest streak** and **weekly consistency %** are **not V1** (§9 scope boundary — a completion has no aggregate beyond the current streak; consistency % belongs to Reports §13). |
| Rail: "Your Streak" with week dots, "Habit Completion" bar chart, "Top Habits" completion-rate list | completion rate over a range = **Reports §13 ("Habit completion")** — legitimate but belongs on Analytics/Reports, not necessarily here |
| "Habit Categories" section | ⚠ **habits carry no category in V1** (core concepts). Out of scope. |
| Motivational banner "Consistency compounds… you're doing great!" | VP3 violation — drop |
| Today column header highlighted | in scope |

**Recommendation:** build the grid: habit name · 7 dated weekday toggle circles · current
streak (number) · kebab (archive / unarchive / edit / delete). A period switcher for
Today (single-column check-off) vs Week (7 columns) is fine. **Drop:** longest streak,
consistency %, habit categories, sub-labels (or merge into name), motivational copy.

---

## Layout

- Shell (§4.1). Header: eyebrow "HABITS" + H1 + subtitle; header illustration.
- **View switcher** (§4.3): Today / This Week / This Month / All Habits.
- Right: **split primary** "＋ Add Habit ▾".
- KPI row (§4.6) — trimmed to V1: "Completed today N / M", "Current streak" (longest
  streak). (Drop the 3rd/4th cards or replace with V1-safe figures.)
- **Habit grid** (card, §4.7): a table —
  - columns: **Habit** (glyph tile + name [+ optional sub]) · **Mon–Sun** (short name +
    date; today emphasised) · **Streak** · **Actions**.
  - rows: one per **active** habit. Each weekday cell = **toggle-circle** (§4.10): filled
    green check if completed that date, hollow ring if not, clickable to toggle.
- Right rail: "Your Streak" (current streak number + this-week dots), a completion bar
  chart (post-V1 / Reports), a quote card.

## Screen-specific components

- **Habit grid table** — CSS Grid or `<table>`; sticky first column on narrow widths;
  scrolls horizontally in `.tl-scroll`-style container.
- **Weekday header cell** — short name + date; today variant (dark pill on date).
- **Toggle-circle** — §4.10; the primary V1 interaction.
- **Add-habit form** (§4.16) — field: **name** only.
- **Archived habits** — a separate list/section or an "All Habits" tab showing archived
  with an "Unarchive" action (history preserved — `requirements` Q11 resolution).

## Interactions

- Click a weekday circle → mark / unmark that habit for that date (any date, incl. future
  — Q9 resolution; the streak still anchors on today/yesterday).
- Kebab → archive / unarchive / edit name / delete.
- Period switcher changes the grid's date span; date stepper (if present) moves the week.

## Responsive

- Grid scrolls horizontally below `tablet`; first column (habit name) stays put.
- On mobile prefer the **Today** view: a simple checklist of habits with one toggle each.
- Right rail drops first.

## Cannot be inferred / ambiguous

- ~~Week semantics~~ — **RESOLVED (D8):** current ISO week, Monday-first.
- "This Month" grid shape (31 columns is impractical → likely a mini heatmap; see
  Analytics "Habit Consistency").
- Whether editing a habit name is offered (implied by kebab; not explicit in §9).
- Flame icon: keep as a small non-animated glyph or drop (P3).

## Design-system references

§4.1 shell · §4.2 header · §4.3 view switcher · §4.5 buttons · §4.6 KPI card · §4.7 card ·
§4.10 toggle-circle · §4.16 create/edit form · `requirements` §9 (+ Q9, Q11) ·
`visual-principles.md` VP3, VP4, VP9, VP10.
