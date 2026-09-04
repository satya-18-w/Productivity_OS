# Productivity OS — Visual Principles

> Decision filters for frontend visual work, derived from the reference images in
> `docs/design/references/` and the product principles in `docs/product/principles.md`.
> When more than one visual treatment is defensible, apply these and pick the option that
> best satisfies them. Each principle states the rule, why it exists, and what it rules
> out. Tokens and components live in `design-system.md`; this file is about *judgement*.
>
> `references/overall.png` is a **global visual reference only** — a poster of many
> panels. It is not a V1 functional screen and defines no scope. V1 functional scope is
> governed by `docs/requirements/v1.md`.

---

## VP1 — One calm surface, one primary action

**Rule:** A screen has a paper-calm background, content in white cards, and **exactly one**
`--brand`-filled button (the "+ Add …" for that screen's entity). Everything else is
text, ghost, or icon buttons.

**Why:** The references are dense with information but never noisy — colour is rationed,
and the eye always finds the one thing to do. Supports P1 (low friction: the capture
action is unmissable).

**Rules out:** Multiple filled primary buttons competing on one screen; brand colour used
for decoration; toolbars of equally-weighted buttons.

## VP2 — Category colour is the only data language

**Rule:** Hue carries exactly one meaning: **which category / series** something belongs
to. Semantic states (success / warning / danger) reuse three fixed hues. Nothing else is
coloured.

**Why:** Consistent across every reference screen — a rose dot means "Personal"
throughout. Makes the UI learnable and keeps `design-system.md` §3.1 honest.

**Rules out:** A per-screen accent; colour-coding priority *and* category *and* status in
three unrelated palettes on the same row; gradients as identity.

**Guard (P4):** category colour is a legibility aid, never load-bearing. Every coloured
element also carries a text label or shape difference, so the screen works in greyscale
and for colour-blind users.

## VP3 — Honest over motivational (reconciling the quote surfaces)

**Rule:** Motivational quotes and nature illustration are **optional chrome in fixed
decorative slots** (sidebar card, right-rail quote card, header backdrop, bottom banner).
They never touch the data region, never wrap a number, and never change based on how
"well" the user is doing.

**Why:** P3 forbids motivational distortion of the data and gamification beyond a plain
streak count; P4 forbids anything implying a second reader ("A better you", "Keep showing
up — you're doing great!"). The references are saturated with this copy; left unbounded it
violates both principles.

**Rules out:** Encouragement text adjacent to totals or streaks; celebratory copy that
appears on good weeks and hides on bad ones; progress framed as praise; badges/levels;
"you're on fire!" next to the streak figure. The streak is shown as a **number**, plainly.

**Status:** **D6 is APPROVED** (ratified 2026-09-04) exactly as stated above — restrained
decorative surfaces only. The hard exclusions are binding: no productivity score/rating/
level, no fake or adaptive encouragement, no gamification (badges, XP, celebrations,
streak-as-achievement), no motivational copy next to any figure or chart, no
second-person aspirational identity copy. When in doubt, leave the slot empty or show
only the brand mark. See `design-system.md` §3.8.

## VP4 — Comfortable density, not cramped

**Rule:** ~14–15px body, ~24px card padding, hairline row separators, generous 24–32px
section rhythm. Tables (habit grid) and multi-column views (week timeline) may go tighter
but keep a 32px+ touch target on interactive cells.

**Why:** The references are information-rich yet breathable. Matches P1 — scanning is
fast, nothing is fiddly.

**Rules out:** 12px body text for content; zero-padding cards; list rows under ~40px tall;
data crammed edge-to-edge.

## VP5 — Structure with flow layout

**Rule:** Page and component layout uses CSS Grid / Flexbox / normal flow. Absolute
positioning is allowed **only** where the domain is spatial — time-proportional blocks on
the day timeline against an hour axis — and even there the container is a normal grid
cell.

**Why:** Project `CLAUDE.md` mandates it; the reference layouts are all
grids (KPI rows, card grids, calendar grids, the 3-pane shell).

**Rules out:** Positioning cards with `top/left`; overlap hacks; pixel-nudged headers.

## VP6 — The three-pane shell is the frame, the middle pane is the screen

**Rule:** Sidebar and top bar are **global** and identical on every authenticated screen
(one implementation). The right rail is **contextual** but built from shared widget
components. A screen spec only describes the **main column** plus which right-rail widgets
it shows.

**Why:** Enables single-screen implementation without loading other specs (the stated goal
of this documentation set). Keeps navigation and shell consistent (project `CLAUDE.md`).

**Rules out:** Per-screen re-styling of the sidebar; a screen inventing its own header
chrome; right-rail widgets that aren't in the component catalogue.

## VP7 — Consistency across screens beats local optimisation

**Rule:** If two screens show the same kind of thing (a list of entities with a status), 
they use the same row component, the same group-header pattern, the same chip. Differences
must be justified by the data, not by taste.

**Why:** the reference screens share one language; a future agent implementing screen N
must be able to trust that patterns from screen 1 apply.

**Rules out:** A bespoke card for goals that could have been the task row with different
fields; re-ordering header elements screen to screen.

## VP8 — Accessibility is not optional

**Rule:** AA contrast for text and meaningful UI; visible `:focus-visible` ring
(`--brand`); ≥ 40px primary hit targets (≥ 32px in dense grids); state conveyed by more
than colour; motion respects `prefers-reduced-motion`; full light/dark parity.

**Why:** A personal daily tool used for years. Also a hard constraint on the category
palette (VP2 guard) and the dark-mode token set.

**Rules out:** Grey-on-grey meta text below AA; colour as the only "completed" signal;
focus outlines removed; dark mode as an afterthought.

## VP9 — Responsive: shed context, keep the core

**Rule:** As width shrinks, drop in this order: right rail → sidebar labels (icon-only) →
sidebar (to drawer). The main column's primary content and its one primary action survive
to the smallest screen. Wide tables/grids scroll horizontally in their own container or
fall back to the agenda/list form; the page never scrolls sideways.

**Why:** N5 requires responsive down to mobile; the references are desktop-first and must
degrade predictably.

**Rules out:** Hiding the "+ Add" on mobile; a horizontally-scrolling page body; showing a
7-column month grid unusably squeezed instead of falling back to a list.

## VP10 — Don't draw what V1 can't do

**Rule:** Every element in a reference is checked against `docs/requirements/v1.md` before
it is built. Out-of-scope affordances (priority, tags, assignees, focus timer, notes,
generic calendar events, %-progress, milestones, insights) are recorded in the screen
spec's "V1 scope alignment" section and **not implemented**.

**Why:** Project `CLAUDE.md` and requirement #12 of this task: do not invent
functionality that can't be inferred from the references *or approved requirements*. The
references are a visual spec, not a scope expansion.

**Rules out:** Building a priority picker because the mock shows chips; adding a notes
route; shipping a Pomodoro timer.
