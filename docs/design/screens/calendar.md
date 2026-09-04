# Screen — Calendar

**Reference:** `docs/design/references/calander.png` *(note the filename misspelling)*
(also panel 6 of `overall.png`)
**Purpose:** Month grid of scheduled items with a day detail rail.
**Proposed route:** `/calendar`

---

## V1 scope alignment

⚠ **Ratified 2026-09-04: Calendar is reference-only and MUST NOT be implemented**
(`design-system.md` §6.4) — no screen, no route, no "event" entity, no nav entry. It is
also redundant with Timeline.

- V1 has **no generic "calendar" or "event" entity.** The only time model is **planned
  blocks** and **actual blocks** with an optional category (`requirements` §3–§4), viewed
  on a **single date's timeline** (§5, which *explicitly* rules out week/month timelines).
- "Add Event", "Event Categories", "Habit review" as calendar items, `+N more`, RSVP-style
  people icons — none are V1.
- This screen is visually the **same artefact as `timeline-month.png`** with cosmetic
  differences (see below).

| Difference vs `timeline-month.png` | Note |
|---|---|
| Week starts **Sunday** here | **Visual defect (D8, ratified).** V1 weeks are ISO → Monday-first (`requirements` N4). Do not reproduce. |
| "Add Event" primary button | V1 creates *blocks*, not events |
| Right rail "Today · <date>" event checklist + "Event Categories" counts | events aren't V1; category counts on time blocks are borderline |

No separate Calendar screen ships in V1. If a month overview is ever wanted it would be a
Timeline Month view (itself out of V1 scope) — one grid, ISO/Monday. Spec retained for
reference only.

---

## Layout (for reference only)

- Shell (§4.1). Header: title icon badge + H1 "Calendar" + subtitle; header illustration.
- **Toolbar:** `‹ ›` + month dropdown ("September 2026 ▾"); **view switcher** Day / Week /
  Month / Agenda; "Today".
- **Month grid:** 7 columns × 5–6 rows. Day cell: date number; today circled `--brand`;
  out-of-month muted; up to ~2 item pills (glyph + title, category-tinted); "+N more".
- Right rail: **mini month calendar** (§4.13); **"Today · <date>"** list of the day's
  items with checkboxes; **"Event Categories"** list with counts + "Manage →".
- Bottom: full-width decorative quote banner.

## Screen-specific components

Same as `timeline-month.md` — **month grid**, **day cell**, **event pill**. Do not build a
parallel set.

## Interactions (inferred)

Prev/next month; day click → day detail; pill click → item edit; "＋ Add Event" → create.

## Responsive

7-col grid → agenda/list fallback below `tablet` (VP9); right rail drops first.

## Cannot be inferred / ambiguous

- Relationship between "events", tasks (with due dates), and time blocks — the reference
  mixes them into one grid with no clear rule.
- Week-start contradiction with the rest of the product.

## Design-system references

`timeline-month.md` (canonical month grid) · §4.13 mini calendar · §4.3 view switcher ·
`visual-principles.md` VP7, VP9, VP10 · `design-system.md` §6.4 (reference-only), D8.
