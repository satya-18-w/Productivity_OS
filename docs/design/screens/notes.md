# Screen — Notes

**Reference:** `docs/design/references/notes.png` (also panel 5 of `overall.png`)
**Purpose:** Capture and organise free-form notes; view one in detail.
**Proposed route:** `/notes`

---

## V1 scope alignment

⚠ **Entirely outside V1 scope.** `docs/requirements/v1.md` has **no Notes feature** — the
14 capability areas are account, categories, day planning, activity logging, timeline,
planned-vs-actual, tasks, board, habits, goals, daily review, weekly review, reports,
export. Notes is listed as an **explicit non-goal-adjacent** idea only in the reference's
own "Ideas for v2 features" note.

The nearest V1 concept is the **daily / weekly review** (§11–§12): fixed-prompt free-text
answers tied to a date / ISO week. That is *not* a free-form note system.

**Ratified 2026-09-04: Notes is reference-only and MUST NOT be implemented** — no feature,
no screen, no route, no nav entry (`design-system.md` §6.4). Documented below only so
that, if Notes is added to a future requirements version, it starts from the shared
language.

Post-V1 sub-features shown: pin, favourite/star, archive, trash, tags, related-notes
links, checklists inside notes, word/character count, rich text.

---

## Layout (for reference only)

- Shell (§4.1). Header: title icon badge + H1 "Notes" + subtitle; header illustration.
- **View switcher** (§4.3): All Notes / Pinned / Favorites / Archive / Trash.
- Right: **primary** "＋ New Note".
- **Two-pane body:**
  - left/main: a **masonry grid** of **note cards** (~3 columns) — glyph tile, title,
    preview (text / bullet list / checklist), tag chips, timestamp, kebab, pin/star icon
    top-right. Selected card gets a `--brand` border.
  - right (wider rail): **note detail** — title, tags, timestamp, full body (with
    checklist items), an "Inspiration" callout (green left-border blockquote), "Related
    Notes" list, "Words / Characters" counts, copy / delete / more actions, an **Edit**
    button.

## Screen-specific components

- **Note card** — masonry item; variable height; hover lift to `--shadow-md`.
- **Note detail pane** — occupies the right rail width; scrolls independently.
- **Inline callout** — green left border + italic text.

## Interactions (inferred)

- Click card → load detail pane. Pin/star toggles. Kebab → archive / trash / duplicate.
- "＋ New Note" → editor. "Edit" → editable detail.

## Responsive

- Masonry collapses 3→2→1 columns; detail pane becomes a full-screen route/overlay on
  mobile.

## Cannot be inferred

- Everything about persistence, format, and search — no backend concept exists.
- Whether "tags" and "categories" are the same thing here.

## Design-system references

§4.1 shell · §4.2 header · §4.3 view switcher · §4.5 buttons · §4.7 card ·
`visual-principles.md` VP10 (this screen is the clearest example of "don't draw what V1
can't do").
