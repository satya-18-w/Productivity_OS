# Screen — Daily / Weekly Review

**Reference:** none (no reference image for either review screen; built from shared
primitives + form patterns, like Board).
**Requirement:** `docs/requirements/v1.md §11` (daily) `§12` (weekly), Q1 (daily prompts,
resolved), Q2 (weekly prompts, resolved), Q9 (future dates allowed).
**Routes:** `/reviews/daily?date=<ISO>` · `/reviews/weekly?date=<ISO-in-week>` (D10).

---

## V1 scope

Both screens share one shape: a **date/week selector** → a **read-only reference panel**
(that period's real totals) → a **fixed, non-editable, four-prompt free-text form** → Save.
Editing a past review overwrites it; there is no history of edits, no ratings/scores, and a
review produces no follow-up actions or carry-over items (§11/§12 scope boundary).

**Daily reference panel (§11):** that date's actual time per category, and which habits
were completed.
**Weekly reference panel (§12, not yet built — Phase 11):** that ISO week's actual time per
category, habit completion counts, and count of tasks that entered `DONE` that week.

**Daily prompts (Q1, resolved — placeholder wording, replace freely):**
1. What went well today?
2. What didn't go as planned?
3. What will you do differently tomorrow?
4. One thing you're grateful for.

**Weekly prompts (Q2, resolved — for Phase 11):**
1. What were the highlights of this week?
2. What did you struggle with?
3. Did your time go where you intended?
4. What is the one priority for next week?

**Q9 (resolved):** a review may be created/edited for a date that hasn't occurred yet — no
"not after today" restriction, unlike Reports' range picker.

**Scope boundary (build none of these):** prompt wording is not user-editable; answers are
free text only (no ratings, scores, or structured fields); no generated output/summary; no
review history/versioning; no linking a review to tasks/goals/habits.

## Layout

- Shell (§4.1), no rail (a single-column form flow, same call as every other Phase 6+
  screen — D6/VP3).
- Header: eyebrow "Reviews" + H1 ("Daily review" / "Weekly review") + one-line subtitle.
- **`DateStepper`** (shared primitive, `components/date/DateStepper.tsx`) — `‹ date ›` +
  native date input + "Today", no max-date (Q9). Weekly will need an ISO-week variant
  (step by 7 days, label = week range) — not yet built.
- **Reference card** — read-only, two columns on wide screens (chips for category time,
  a checklist for habits) stacking to one column on mobile.
- **Prompts card** — one `Field` + `Textarea` per fixed prompt, in order, then a Save
  button. A factual "Saved" note appears next to the button once saved and unedited since
  (no motivational framing — VP3).

## Screen-specific components

- `DateStepper` — extracted from Timeline's toolbar in this phase (`design-system.md`
  register — see "DateStepper" component note in `frontend-implementation-plan.md`); now
  shared by Timeline and Daily Review.
- Reuses `Card`, `Field`, `Textarea`, `Button`, `Chip` (category dot + label, read-only),
  `ErrorState`. No new primitive needed for the reference panel — it's list markup, not a
  chart (P3 — the figures are the point).

## Interactions

- Change the date → reference panel and prompt answers both reload for that date.
- Type in a prompt → the "Saved" note clears until the next successful save.
- Save → upserts the review for that date (create or overwrite — the screen doesn't
  distinguish the two at the API level).

## Responsive

- Reference panel's two columns → one column on mobile. No page h-scroll.

## Design-system references

§4.1 shell · §4.7 card · `requirements` §11, §12, Q1, Q2, Q9 · `visual-principles.md` VP3.

---

## Phase 10 — Daily Review — Status: ✅ COMPLETE (2026-09-04)

Route `/reviews/daily` → `DailyReviewScreen` (`web/src/features/reviews/`). **No reviews
backend exists** — the review record itself runs on an in-memory mock
(`docs/left.md`, "Phase 10 — Daily Review"); the reference panel is **real data** (existing
`api.comparison(date)` + `api.habits(date)`, same calls Timeline/Habits already make).

- [x] Extracted **`DateStepper`** (`components/date/DateStepper.tsx`) from Timeline's
      inline toolbar code and refactored `TimelineScreen` to use it too — one shared
      component instead of two copies of `‹ date › + Today` (CLAUDE.md: no duplicate
      visual systems). `label="Review date"` on this screen for a distinct accessible name.
- [x] Reference panel: actual time per category as `Chip`s (category-colour dot, D2,
      `fmtDuration` value), filtered to categories with actual time > 0; habits list with a
      check mark for `completed_on_date`. Loading + error (retry) states.
- [x] Four fixed `Field`+`Textarea` prompts (Q1 wording), free text only, 5000-char cap.
      "Save review" (new) / "Save changes" (existing) button label; a "Saved" note clears
      on the next edit.
- [x] **Dropped per spec:** ratings/scores, generated summaries, review history, linking to
      other entities — none are V1.
- [x] No rail. Responsive — reference panel 2-col → 1-col, no page h-scroll. Light + dark
      verified.
- [x] Tests — `DateStepper` (3, incl. Today disabled-when-current), `reviewData` (4, incl.
      overwrite-on-resave), `DailyReviewScreen` (4, incl. zero-time category filtered out,
      save round-trip, date-stepper reloads that date's own state). Full suite green.
- [x] Browser-verified — reference chips + habit checklist populate from real fixture data,
      typing + saving an answer, "Saved" note, navigating dates resets the form, dark,
      mobile; no console errors.
- [ ] Committed — pending product owner.

Old placeholder route removed from `App.tsx`.
