# Productivity OS — Design System

> **Status:** ratification pass complete 2026-09-04. Extracted from the visual references
> in `docs/design/references/`. This document is the single shared source of visual
> language for reference-driven frontend work. Screen specifications in
> `docs/design/screens/` reference the tokens and components named here instead of
> re-describing them.
>
> **This document does not authorise implementation.** It records design *direction*.
> §6 is the decision register: decisions marked **APPROVED** are settled direction;
> decisions marked **PENDING** are not, and anything gated on a PENDING decision (the
> app-shell restructure, route names, the V1 screen list, exact token values) must wait
> for the owning document. V1 functional scope is always governed by
> `docs/requirements/v1.md`, which outranks every visual reference.

---

## 1. How to use this document

- Treat the reference PNGs as **visual specifications**, not inspiration.
- When implementing one screen, load: that screen's spec, its reference image, the
  **specific** sections of this file it names, and the relevant components — not this
  whole file and not other screens.
- Do **not** introduce a colour, spacing value, radius, shadow, typography step or
  breakpoint that is not defined here. If a screen genuinely needs a new one, add it here
  with a rationale and get product-owner approval before using it (per project
  `CLAUDE.md` → "Design System Changes").
- There is **one** visual system. The existing `web/src/styles.css` token block is its
  current expression; this document proposes a bounded set of changes to bring it in line
  with the references. Do not fork a second system.

## 2. Relationship to existing code and ADRs

| Concern | Owned by | Note |
|---|---|---|
| Token names, base reset, existing component classes | `web/src/styles.css` | Current implementation. Token **names** and scale are kept; some **values** change (below). |
| Frontend stack, routing, single-origin | ADR-0006 | Routes proposed in screen specs are **not** ratified here — they belong in `docs/architecture/conventions.md`. |
| What the user can observe and do | `docs/requirements/v1.md` | The references show substantially more than V1. Each screen spec maps the gap. |
| Decision filters when choices are close | `docs/product/principles.md` | See `visual-principles.md` for the visual reading of P1–P6. |

## 3. Foundations

### 3.1 Colour

The references use a **warm, nature-leaning palette**: an off-white paper ground, a single
**forest-green brand**, and a fixed multi-hue set used *only* to identify categories and
data series.

#### Brand / primary — APPROVED (D1)

The reference **deep forest / dark green** is the Productivity OS brand and action colour
(logo tile, primary buttons, active view-switcher segment, active nav pill text, focus
ring). It **replaces** indigo `--accent` (`#5b5bd6`) in `web/src/styles.css`.

| Token | Value | Role |
|---|---|---|
| `--brand` / `--accent` | forest green, ~`#1f5132`, hover ~`#183f27` | Primary buttons, active segment, focus ring, brand mark |
| `--brand-soft` | pale mint, ~`#e7f0e9` | Active nav pill background, subtle fills |
| `--on-brand` | `#ffffff` | Text/icon on primary |

The hue direction is settled; the **exact hex values are PENDING precise token extraction**
(register item T1). Dark-mode counterparts follow the existing `[data-theme="dark"]`
pattern and are extracted in the same pass.

#### Neutrals / surface — APPROVED (D5), warmer than current

| Token | Approx value | Role |
|---|---|---|
| `--bg` | `#f6f7f4` | App background (very slightly warm off-white) |
| `--surface` | `#ffffff` | Cards, panels, rows |
| `--surface-2` | `#f2f3ef` | Inset areas, table zebra, sidebar wells |
| `--sidebar-bg` | `#fbfbf9` | Left sidebar |
| `--border` | `#e9e9e3` | 1px hairlines, card edges |
| `--border-strong` | `#d7d7cf` | Inputs, dividers needing weight |

Values are direction, not final — **exact hex PENDING token extraction (T1)**.

#### Category / data palette — APPROVED (D2) as a visual/semantic system only

Eight fixed hues. This is the **only** chromatic language for data. Values approximate;
each needs a solid (chip text / dot / series) and a soft (chip background / block fill)
form, verified for AA contrast — **exact values PENDING token extraction (T1)**.

**D2 constraint (binding):** category colour is *identification and legibility only*. It
must **never** drive business logic — no validation, filtering semantics, ordering,
totals, permissions, or any stored product meaning depends on a category's colour. Colour
is a presentation attribute the frontend assigns; the domain does not see it. This
reinforces `requirements` §2 ("no colour or icon carries product meaning").

| Name | Solid (approx) | Used for |
|---|---|---|
| `--cat-personal` | rose `#e0679a` | "Personal" category, series |
| `--cat-work` | blue `#3b82f6` | "Work" |
| `--cat-study` | violet `#8b5cf6` | "Study" |
| `--cat-health` | emerald `#22a565` | "Health" |
| `--cat-projects` | amber `#e6a532` | "Projects" |
| `--cat-ideas` | orange `#ee7326` | "Ideas" |
| `--cat-finance` | deep green `#1f5e3a` | "Finance" |
| `--cat-other` | grey `#9aa0a6` | Uncategorized / "Others" bucket |

> V1 note: categories in V1 are **flat, user-defined labels** with no product meaning
> attached to colour (`requirements` §2). The named set above ("Personal", "Work", …) is
> **sample data in the references**, not a fixed taxonomy. A frontend may assign a colour
> per category for legibility; nothing may depend on it. The palette is a *pool of hues to
> assign from*, not a required list of categories.

#### Semantic

| Token | Approx | Role |
|---|---|---|
| `--success` | `#1c8a4e` | Done, achieved, positive difference |
| `--warning` | `#c9821f` | Needs-attention, negative difference |
| `--danger` | `#e04b4b` | Overdue, destructive |
| `--info` | `#3b82f6` | Informational / neutral emphasis |
| `--goal` | `#8b5cf6` | Goal accent (violet) |

Each has a `-soft` background tint form. Exact values **PENDING token extraction (T1)**.

> A single-hue **streak** accent (orange, ~`#f0642f`) may be used only for the V1 *current
> streak* number (`requirements` §9). It is a colour, not a motivational device — no
> flame animation, no "longest streak", no achievement framing (D6; `visual-principles.md`
> VP3). "Focus time" is **not a V1 measure** — do not add a token or accent for it.

#### Text

| Token | Approx | Role |
|---|---|---|
| `--text` | `#1c241d` | Headings, primary text (warm near-black) |
| `--text-secondary` | `#55605a` | Body, descriptions |
| `--text-muted` | `#8a938c` | Meta, timestamps, axis labels, eyebrows |

### 3.2 Typography

- **Family — APPROVED (D9):** keep the existing `--font-sans` Inter-first system stack.
  **Do not introduce a second typeface.** The reference headings may look like a rounded
  display face; Inter at the weights below is the approved match. Revisiting this needs a
  fresh decision with an E3 justification.
- **Numerals:** tabular (`font-variant-numeric: tabular-nums`) for every stat, count,
  time, date, axis and table cell.

| Step | Size (approx) | Weight | Used for |
|---|---|---|---|
| Display / H1 | 28–32px | 700 | Page title in the header block |
| H2 | 20–22px | 640 | Major section titles |
| H3 / card title | 16–17px | 620 | Card and widget headings |
| Body | 14–15px | 400–450 | Descriptions, row content |
| Body-strong | 14–15px | 600 | Row names, emphasised inline |
| Label | 12–13px | 550 | Form labels, tab labels |
| Eyebrow | 11–12px | 650, `letter-spacing: .06em`, uppercase | Kicker above page title, table column heads, column-group heads |
| Meta | 12px | 400 | Timestamps, counts, hints |
| Stat number | 24–30px | 680 | KPI card figures |

Existing `--fs-*` token names are retained; the scale is widened slightly at the top end
(H1) to match the references.

### 3.3 Spacing

4px base scale — **unchanged** from `web/src/styles.css` (`--sp-1`…`--sp-7` =
4/8/12/16/24/32/48). Relationships observed in the references:

- Card internal padding: `--sp-5` (24px); compact widget cards `--sp-4` (16px).
- Grid gaps between cards: `--sp-4`–`--sp-5`.
- Vertical rhythm between page sections: `--sp-6` (32px).
- List row vertical padding: ~11–12px; row gap 0 with hairline separators.
- Header block → content: `--sp-5`.

### 3.4 Radius

Existing `--radius-*` scale retained (5 / 8 / 11 / 16 / 999). Observed usage:

- Inputs, small chips, table check circles container: `--radius-sm` (8px).
- Cards, panels, columns: `--radius-md`→`--radius-lg` (11–16px).
- Segmented control, filter chips, category chips, count badges, avatars: `--radius-full`.
- Icon tiles (category glyph squares, block icons): ~10–12px (`--radius-md`).

### 3.5 Elevation

Restrained — existing `--shadow-sm/md/lg` retained.

- Cards / widgets: 1px `--border` **or** `--shadow-sm`, rarely both.
- KPI cards and right-rail widgets: often borderless on a soft `--shadow-sm`.
- Hover on interactive cards (task card, note card): lift to `--shadow-md`.
- Overlays (modal, dropdown, date picker) — not shown in references: `--shadow-lg`.

### 3.6 App-shell layout grid — APPROVED (D3); exact widths PENDING (T1)

A **persistent three-region shell**: left sidebar (primary nav) + main content + right
contextual rail. Replaces the current top-nav / 760px column (`web/src/AuthLayout.tsx`).
Build spec: `docs/design/screens/app-shell.md`.

| Region | Indicative width (T1) | Role |
|---|---|---|
| Left sidebar | ~248px (`--sidebar-w`) | Primary navigation; brand lockup; user chip. |
| Main content | fluid | `PageHeader` + screen body. |
| Right rail | ~320px (`--rail-w`) | **Per-screen** contextual widgets; a screen may have none. |
| Outer gutters | `--gutter` (24px) | |

`--content-max: 760px` no longer applies to authenticated screens; auth screens
(`/login`, `/register`) keep the centered narrow layout and no shell.

### 3.7 Breakpoints — shed order APPROVED (D4); thresholds PENDING (T1)

The **responsive shed order is approved**: as width decreases —

1. remove / reduce the **right contextual rail** first;
2. then **collapse the sidebar labels** (icon-only);
3. then, on mobile, the **sidebar becomes a drawer / navigation pattern**.

The main column's primary content and its one primary action survive to the smallest
screen; the page never scrolls sideways (wide tables and any week/month grid scroll inside
their own container or fall back to a list). Exact pixel thresholds below are indicative
and finalised with the token pass:

| Name | Indicative range | Shell behaviour |
|---|---|---|
| `wide` | ≥ 1280px | Sidebar + main + right rail. |
| `laptop` | 1024–1279px | Right rail drops below main or hides behind a toggle. |
| `tablet` | 640–1023px | Sidebar icon-only or a slide-over drawer; single column. |
| `mobile` | < 640px | Sidebar → drawer / nav pattern; segmented controls scroll; wide grids fall back to list. |

### 3.7 Iconography

- Line icons, ~18–20px, `1.5px` stroke, `currentColor`.
- Nav icons sit left of the label.
- **Category / entity glyph tiles:** a rounded square (~28–36px) filled with the soft
  category tint, holding a solid-colour glyph or an emoji. Used in the sidebar "Spaces"
  list, category cards, goal rows, note cards, timeline blocks, habit rows.
- No icon set is chosen here — that is a frontend decision (record it in conventions).

### 3.8 Illustration & motivational surfaces — APPROVED (D6), tightly bounded

The references lean heavily on soft nature artwork (mountains, forests, sunrise, leaf) and
motivational quotes in the sidebar card, right-rail quote cards, page-header background
bleeds, and full-width bottom banners. **D6 permits these only as restrained decorative
surfaces** in fixed slots, with no data role.

**D6 forbids** (these are hard, from P3 / P4 / P6):

- Any productivity *score*, rating, grade, level, or index.
- Fake or adaptive encouragement — copy that praises the user, or that changes / hides
  based on how "well" they are doing.
- Gamification: badges, medals, XP, celebrations, confetti, streak-as-achievement, "on
  fire" / "great job" adornment next to figures.
- Motivational or coaching copy adjacent to any number, total, streak, or chart.
- Second-person aspirational identity copy that implies another reader ("a better you").

Decorative art and a neutral, static quote in a dedicated slot are acceptable; when in
doubt, leave the slot empty or show only the brand mark. See `visual-principles.md` VP3.

---

## 4. Shared components

Each entry: what it is, its variants/states, and where it recurs. Screen specs reference
these by name. Components not yet in `web/src/styles.css` are marked **(new)**; ones only
*inferred* (never fully shown) are marked **(inferred)**.

### 4.1 Shell

- **Sidebar** *(new shell — D3 approved; build spec `screens/app-shell.md`)* — brand lockup (leaf glyph in a `--brand` tile +
  "Productivity OS" + "Plan · Focus · Grow" micro-tagline + a collapse chevron); primary
  nav list; optionally a decorative card (D6 bounds); a user chip (avatar, name, plan,
  gear).
- **Sidebar nav item** *(new)* — icon + label, full-width hit area, `--radius-sm`.
  States: default (muted text), hover (`--surface-hover`), **active** (`--brand-soft`
  pill, `--brand` text/icon). Optional trailing **count badge**. Label collapses to
  icon-only per the D4 shed order.
- **Spaces list** *(reference-only — DO NOT BUILD)* — the references show colour-dot +
  label rows under a "SPACES" heading acting as a category/workspace switcher. There is
  **no V1 concept** for this (categories are flat labels on time blocks; §2). Excluded
  until a product requirement introduces it — see register item C1.
- **User chip** *(new)* — avatar initials circle, name, plan label, settings gear.
- **Top bar** *(partly reference-only)* — the shell may carry a theme toggle and the user
  avatar. **Global search and the notification bell are NOT V1** (no search feature, no
  notifications — §V1 non-goals) — do not build them.
- **Right rail** *(new — D3 approved)* — vertical stack of widget cards; contents are
  strictly screen-specific and limited to what that screen's spec lists. A screen may
  have no rail. First region to drop as width shrinks (D4).

### 4.2 Page header

- **Eyebrow** (uppercase kicker) *or* a **date line** (e.g. "Thu, 4 Sep 2026").
- **Title** (Display/H1), often with a trailing emoji/leaf and a playful phrasing.
- **Subtitle** — one line: a description or a short quote.
- **Header illustration** *(optional)* — a faint mountain/forest scene bleeding off the
  top-right of the header band (dashboard, tasks, habits, goals, calendar, categories,
  analytics). Decorative only.
- **Title icon badge** *(optional)* — green rounded tile + white glyph, left of the title
  (tasks, calendar, categories, analytics, notes).

### 4.3 View switcher (segmented control) *(new)*

Pill group; exactly one segment active (`--brand` fill, `--on-brand` text); others are
plain text on a `--surface-2` track. Used for: Timeline & Calendar Day/Week/Month/Agenda;
Tasks status filter; Habits period; Goals & Categories & Analytics category/section tabs.
Horizontally scrollable on mobile.

### 4.4 Filter-chip row *(new)*

Row of rounded-full chips, each: optional colour dot + label + count in parens. One or
many active (active = `--brand` fill or category-tinted). Seen on Timeline Agenda and
implied on Tasks/Goals/Categories.

### 4.5 Buttons

- **Primary** — `--brand` fill, `--radius-sm`, `+ Label`. Top-right of content.
- **Split primary** *(new)* — primary with a trailing caret for a menu (Timeline "Add ▾",
  Tasks "Add Task ▾").
- **Ghost / secondary** — `--surface`, `--border-strong`, `--text`.
- **Icon button** — square, borderless, muted icon, hover `--surface-hover` (kebab `⋯`,
  chevrons, theme toggle).
- **Link button** — `--brand` text, no chrome ("View all →", "Manage →").

### 4.6 Stat / KPI card *(new)*

Soft tinted background (semantic or category hue) **or** white; a circular icon badge; a
large tabular number; a label; an optional sub-label. Appears as a **row of 4** on several
reference screens.

> The references also show a **trend delta** ("+12% vs last month") and a **corner
> sparkline** on KPI cards. Period-over-period comparison is **not V1** (`requirements`
> §13: "no comparison between two ranges"). Build the card as number + label only; no
> delta, no trend spark.

### 4.7 Card / widget card

White `--surface`, `--radius-md`, `--border` or `--shadow-sm`, `--sp-5` padding. Optional
header row: H3 title + a "View all →" link. Right-rail widgets are the compact variant.

### 4.8 List row + group header *(partly new)*

- **Row** — leading control (checkbox / drag handle / icon tile), primary name
  (Body-strong), inline chips (category, status, priority), trailing meta (date, count,
  avatars), kebab. Hairline separator between rows.
- **Group header** *(new)* — a section label with a **coloured left accent bar** and a
  count (e.g. "Overdue (2)" red, "Today (4)" green, "Completed (5)" green). Used on Tasks;
  the pattern is reusable.

### 4.9 Chips & badges *(partly new)*

- **Category chip** — pill, colour dot + label, soft category tint background. Colour is
  presentation only (D2).
- **Status chip** *(new)* — soft semantic tint. In V1 this renders **goal progress state
  only**, using the four `requirements` §10 labels verbatim: **Not started / In progress /
  Achieved / Abandoned**. Do **not** use the reference's "On Track" / "At Risk" wording —
  those imply a derived health signal V1 does not compute. See `screens/goals.md`.
- **Priority chip** *(reference-only — DO NOT BUILD)* — High / Medium / Low. **Task
  priority is explicitly excluded from V1** (`requirements` §7 scope boundary).
- **Tag chip** *(reference-only — DO NOT BUILD)* — free-text tags on blocks / notes.
  **Tags are not a V1 concept** (a time block has only start / end / category).
- **Count badge** — small rounded-full number (nav items, column heads, tab labels).

### 4.9a Menu (kebab / actions) — **built** (`components/ui/Menu.tsx`)

A WAI-ARIA menu-button. `trigger` (usually an `IconButton` with `MoreIcon`) is cloned
with `aria-haspopup="menu"` / `aria-expanded`; `items` is a flat list of
`{ label, onSelect, danger?, disabled? }` or `{ separator: true }`. Opens on click /
Enter / Space / ArrowDown; arrow keys move; Enter selects and closes; Esc / outside-click
close and return focus to the trigger. Used for row actions (Tasks, later Goals/Habits/
Categories).

### 4.10 Checkbox / toggle-circle *(partly new)*

- **Checkbox** — rounded square; checked = `--brand`/`--success` fill + white check;
  completed list items get a struck-through, muted name.
- **Toggle-circle** *(new)* — circular; used in the habit grid and week strips. Filled
  green check = completed for that date; hollow ring = not. This is the V1 habit
  completion control.
- **Bug fixed 2026-09-04** (found during a reference-accuracy audit): the decorative ring
  (`.ui-checkbox__box` / `.ui-toggle-circle__ring` / `.ui-switch__track`+`__thumb`) is an
  absolutely-positioned sibling painted *after* the real `<input>`, so without
  `pointer-events: none` it silently sat on top and swallowed every mouse/touch click —
  the input only ever toggled via keyboard or a wrapping `<label>`. This affected the
  Tasks row checkbox and every Habit completion toggle (Today/Week views); `Switch` had
  the same defect but wasn't in use yet. Vitest/RTL didn't catch it because `userEvent`
  dispatches directly to the target element without real hit-testing. Fixed in
  `web/src/styles/primitives.css`.

### 4.11 Progress bar *(new)*

Thin rounded track + fill. Used in V1 only where a genuine ratio exists — e.g. a report's
proportion bar. Fill draws from the semantic / category palette.

> **Not for goals.** Goal progress in V1 is one of **four manual states**, not a
> percentage (`requirements` §10) — no progress bar, no `%`, no "12 / 20 tasks". The
> reference's goal progress bars are **excluded**. See `screens/goals.md`.

### 4.12 Data-viz primitives *(new)* — follow the `dataviz` skill before building

- **Donut / ring** — proportion of a whole with a centre total and a legend of
  label + value + %.
- **Vertical bar chart** — one bar per period (e.g. daily actual totals).
- **Horizontal bar list** — ranked label + bar + value.
- **Table** — the literal figure form; often the honest primary presentation (P3).

> Which of these renders each V1 report is **PENDING the Reports specification** (register
> item R1) — do not pre-commit a chart choice here. The reference's *combo trend line*,
> *habit-consistency heatmap*, and *period-delta* visuals imply trends / range comparison
> that V1 §13 excludes; treat them as reference-only. All data-viz draws series colours
> from the category / semantic palette and must read correctly in light and dark; follow
> the `dataviz` skill.

### 4.13 Mini month calendar — **built** (`components/date/MiniCalendar`)

Right-rail widget: month label + prev/next, 7-col weekday grid **Monday-first (D8)**,
today ringed in `--brand`, selected day filled. `value` (ISO date) + `onChange`. Each day
is a `<button>` with a full-date `aria-label` and `aria-pressed`/`aria-current="date"`.
Selecting a day drives the screen's date. (Activity dots per day — deferred.)

> **Week starts Monday — RESOLVED (D8).** ISO week semantics (`requirements` N4) are
> authoritative everywhere: the mini-calendar, any week/month grid, and all date/week
> bucketing are **Monday-first**. The Calendar reference's Sunday-first grid is a
> **visual defect** — do not reproduce it.

### 4.14 Decorative surface (quote card / banner) — APPROVED (D6), bounded

- **Quote card** — small rounded card, optional nature art, one short **static, neutral**
  quote. A dedicated slot in the sidebar or right rail. No progress bar, no data.
- **Banner** — full-width nature image with a single **navigational** CTA (e.g. a link to
  another screen). Not a "keep going!" nag.

See §3.8 for what D6 forbids. When unsure, omit the surface. Never place this adjacent to
a number, streak, total, or chart.

### 4.15 Empty state *(exists)*

Centered icon + short message + a single primary action ("No tasks yet — Add a task and
get started"). Keep the existing `.empty` pattern, add the illustrated variant from
`overall.png` panel 14.

### 4.16 Create / edit form *(inferred)*

Never fully shown (only "+ Add" affordances and inline "Add a task" / "Add a new block"
rows). A modal or slide-over with the entity's fields is implied. Fields per entity come
from `docs/requirements/v1.md`, **not** from the references (which show out-of-scope fields
like priority, tags, assignees). Mark all create/edit UI as inferred; spec it per screen
against V1 fields only.

---

## 5. Rules (anti-patterns)

1. No new colour / spacing / type / radius / shadow / breakpoint token without adding it
   here **and** getting approval (project `CLAUDE.md`).
2. Category / semantic hues are the **only** decorative colour. No arbitrary per-screen
   accent. Category colour is presentation only — no logic depends on it (D2).
3. Prefer Grid / Flexbox / normal flow. Absolute positioning only where the model demands
   it (e.g. time-proportional timeline blocks) and never for page layout.
4. Reuse the components above. Do not build a second "card", "chip" or "button".
5. Do not implement affordances for features outside approved V1 scope; document them in
   the screen spec's "V1 scope alignment" section and stop. The reference-only list in §6
   is a hard exclusion list.
6. No motivational scoring, adaptive encouragement, or gamification (D6 / §3.8).
7. Nothing may be built against the three-region shell or a specific route until D3 / D10
   land in `docs/architecture/`.
8. Verify every implemented screen in a real browser (Playwright) before claiming it done.

---

## 6. Decision register

Ratification pass 2026-09-04; D3 / D10 approved 2026-09-04 (product owner). **APPROVED**
items are settled and reflected throughout this document. **PENDING** items are not
settled; nothing gated on them may be implemented.

### 6.1 Approved

| # | Decision | Reflected in |
|---|---|---|
| **D1** | Primary brand / action colour is the reference **deep forest green**, replacing indigo `--accent`. Exact hex via T1. | §3.1 Brand |
| **D2** | Adopt the **8-hue category palette** as a visual/semantic identification system **only** — category colour must never drive business logic. Exact hex via T1. | §3.1 Category palette, §5 rule 2 |
| **D3** | **Three-region app shell** — left sidebar (primary nav) + main content + right contextual rail — replacing the current top-nav / 760px column. Rail is per-screen; drops first (D4). Exact `--sidebar-w` / `--rail-w` via T1. | §3.6, `conventions.md` → Frontend, `screens/app-shell.md` |
| **D4** | Responsive **shed order**: right rail → collapse sidebar labels → mobile sidebar drawer. Main content + primary action always survive; no sideways page scroll. Thresholds via T1. | §3.7, `visual-principles.md` VP9 |
| **D5** | **Warmer off-white / paper** neutral direction (not the current cool greys). | §3.1 Neutrals |
| **D6** | Motivational surfaces exist **only** as restrained decoration in fixed slots. **No** productivity scoring, fake/adaptive encouragement, gamification, or anything conflicting with P3 / P4 / P6. | §3.8, §4.14, §5 rule 6, `visual-principles.md` VP3 |
| **D8** | **Monday-first / ISO week** semantics are authoritative everywhere. The Sunday-first calendar reference is a visual defect. | §4.13, `screens/calendar.md`, `screens/timeline-month.md` |
| **D9** | Keep the existing **Inter** font stack. Do not introduce another typeface. | §3.2 |
| **D10** | **SPA routes ratified**: `/` → Timeline (today) · `/timeline` · `/tasks` (list) · `/board` (Kanban) · `/habits` · `/goals` · `/categories` · `/reports` · `/reviews/daily` · `/reviews/weekly` · `/account` · `/export` · `/login` · `/register`. Tasks and Board are **separate** routes over the same task model. `/` landing is Timeline — **no dashboard** (D7 / §6.4). | `conventions.md` → Frontend, `screens/*.md` route lines |
| **G1** | **Timeline block geometry (approved 2026-09-04):** blocks are **time-proportional** (height = duration) positioned against a 24-hour axis; two **labelled lanes** (Planned \| Actual). Block fill/border = its **category colour** (VP2); **planned** blocks are dashed-border + lighter fill, **actual** blocks solid — so planned/actual read from lane + line-style, not hue. Midnight-spanning blocks show ▲/▼ markers on the day boundary. Full 00:00–24:00 range, vertically scrollable. | `screens/timeline.md`, existing `.tl-*` in `web/src/styles.css` |
| **R1** | **Report visualisation (approved 2026-09-04):** time-by-category → horizontal bars (category colour); planned-vs-actual → table (`table.totals`/`.pos`/`.neg`); habit completion → table + `ProgressBar`; task throughput → single `StatCard`; daily actual totals → vertical bar chart, scrollable. No charting library; literal values always shown as text alongside every mark (dataviz skill, P3). | `screens/analytics.md` §"Phase 9", `web/src/features/reports/` |
| **G2** | **Timeline Week/Month + focus timer (approved 2026-09-05, product owner):** Week and Month join Day/Agenda in the view switcher. **Week** = 7 day-columns (Monday-first, D8), each a chronological stack of that day's blocks (category colour + dashed/solid per G1, not hour-proportional — a week is too dense for that). **Month** = a calendar grid (Monday-first), each day cell a compact list/count of that day's blocks, opening the Day view on click for detail. Neither view adds a KPI row, donut, or any other dashboard widget (§6.4 still excludes those). **Focus timer**: a standalone `Card` in the Day view's rail — preset durations, start/pause, countdown — with no persistence and no link to block data (`v1.md §4`). | `screens/timeline-week.md`, `screens/timeline-month.md`, `web/src/features/timeline/` |

### 6.2 Pending — do not implement against these

| # | Decision | Owner / gate |
|---|---|---|
| **D7** | Which screens are in the **V1 frontend**. Governed entirely by `docs/requirements/v1.md` — the reference set does not expand scope. See §6.3 / §6.4. | `docs/requirements/v1.md`; a requirements revision if scope is to change. |
| **T1** | Precise extraction / ratification of **exact token values** — brand, category, semantic, neutral hues (light + dark), final breakpoint pixel thresholds, and `--sidebar-w` / `--rail-w`. | A dedicated token-extraction pass. Until then, all hex in §3 is direction only. |
| **C1** | Category **persistence model and detail** — whether a category stores a colour; whether it can be unarchived; whether categories ever attach to entities beyond time blocks; the sidebar "Spaces" concept. | A ratified product requirement. Until then: categories are flat labels on time blocks (§2); "Spaces" is not built. |

### 6.3 V1 screens eligible for implementation (D3 / D10 approved 2026-09-04)

All are governed by `docs/requirements/v1.md`:

| Screen | Spec | V1 requirement |
|---|---|---|
| Timeline — **Day** | `screens/timeline.md` | §3, §4, §5 |
| Timeline — **Agenda** (single-day list rendering) | `screens/timeline-agenda.md` | §5 (alternate rendering of one day) |
| Timeline — **Week** (G2, approved 2026-09-05) | `screens/timeline-week.md` | §5 (amended) |
| Timeline — **Month** (G2, approved 2026-09-05) | `screens/timeline-month.md` | §5 (amended) |
| **Tasks** (list) | `screens/tasks.md` | §7 |
| **Board** (Kanban) | reference not provided; requirement §8 | §8 |
| **Habits** | `screens/habits.md` | §9 |
| **Goals** | `screens/goals.md` | §10 |
| **Categories** (management) | `screens/categories.md` | §2 |
| **Reports** | `screens/analytics.md` (reframed) | §13 (five fixed reports) |
| **Planned vs actual comparison** (per-date) | part of Timeline / Agenda | §6 |
| **Daily review** / **Weekly review** | no reference image; requirements §11 / §12 | §11, §12 |
| **Account** (email, password change, timezone) | no reference image; requirements §1 | §1 |
| **Auth** (login / register) | panel 12 of `overall.png` (global ref only) | §1 |
| **Data export** | requirements §14 | §14 |

### 6.4 Reference-only — MUST NOT be implemented

Present in the references (or `overall.png`) but **absent from V1 requirements**. Do not
build, and do not add affordances for:

- **Dashboard / home overview** screen (aggregate landing page).
- **Notes** (feature and screen) — no V1 concept.
- **Calendar** as a separate feature/screen, and any generic "event" entity.
- **Analytics** beyond the five fixed reports — trend lines, period-over-period deltas,
  "insights", heatmaps as required output, focus-time metrics, per-report export.
- **Recurring tasks** / any recurrence engine.
- **Task priorities**, **task tags**, **assignees / collaborators**.
- **Categories on tasks, habits, or goals**; the **"Spaces"** sidebar switcher.
- **Goal milestones**, **linked tasks**, **numeric / % goal progress**.
- **Habit** "longest streak", "weekly consistency %", habit categories, habit sub-labels.
- **Social / collaboration / sharing** of any kind.
- **AI planning / suggestions / auto-categorisation**.
- **Calendar synchronisation / import** (Google Calendar etc.).
- **Notifications**, reminders, the **notification bell**.
- **Global search** (the top-bar "Search… ⌘K").

**Timeline Week/Month and a focus timer are no longer on this list** — amended
2026-09-05 (product owner), see `v1.md §4`/`§5`. They are now approved (G2, §6.1) with
one caveat carried over unchanged: the **dashboard-style widgets** that surround them in
the references (KPI/sparkline rows, donuts, "Insights", "Upcoming Events", Weekly
Goals/Habit-Tracker cross-widgets) are **still excluded**, for the reasons already given
above (P3, P6, VP3) — this amendment reopens only the timeline-of-blocks rendering and
the standalone timer, not the wider dashboard aesthetic.
- Motivational **scoring / badges / gamification**.
