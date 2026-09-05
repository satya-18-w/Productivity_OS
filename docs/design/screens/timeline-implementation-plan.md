# Timeline (Day) — Implementation & Audit Plan

**Date:** 2026-09-04. **Target:** `docs/design/references/timeline.png`. **Requirement
source of truth:** `docs/requirements/v1.md §3–§6`. **Current code:**
`web/src/features/timeline/{TimelineScreen,TimelineGrid,AgendaList,BlockDialog,
ComparisonCard,TodayTasks}.tsx`, `web/src/styles/timeline.css`.

This audit re-checks the already-shipped Timeline Day screen (`docs/design/screens/
timeline.md` Phase 2, plus later follow-ups: `SplitButton`, `TodayTasks` rail widget,
`DateStepper`) against the reference image and against V1 scope, with the app actually
running (Playwright, 5 viewports, live fixture data). It does **not** redesign the
screen — G1 (two lanes, time-proportional, category-colour, dashed/solid) stays as
ratified in `design-system.md §6.1`.

---

## Executive summary

- **1 CRITICAL bug found and fixed**: category labels were clipped or fully invisible
  on any block ≤ ~50 minutes — a large share of realistic blocks, including several of
  the reference's own examples (a 45-min "Team Sync", "Morning Routine").
- **1 MAJOR gap found and fixed**: the page always opened on empty midnight hours
  instead of the populated part of the day.
- **1 MINOR cosmetic issue found, deferred**: the "now" line can cross through an
  active block's text.
- **Everything else** the reference shows that's in V1 scope was already present and
  correct (two lanes, planned/actual distinction, category colour, midnight markers,
  date navigation, split-button add, mini-calendar rail, Today's-tasks rail, comparison
  table). Everything the reference shows that's **out** of V1 scope was already
  correctly absent.
- **Backend**: fully supports everything Timeline Day needs. No `backend-planning.md`
  entry required for this screen. One **adjacent, out-of-scope** finding is recorded
  below (category colour/icon) because it directly affects block colour accuracy, even
  though fixing it means touching the Categories screen, which this task does not.

---

## Phase 1 — Gap checklist (reference vs. current)

Legend: ✅ matches / already correct · 🔧 fixed in this pass · ⏸ deferred (documented,
not fixed) · 🚫 correctly excluded (V1 scope boundary).

### Shell

| Item | Status | Note |
|---|---|---|
| Sidebar (icon nav, collapses at narrow widths) | ✅ | Unchanged, shared shell |
| Main content width / column proportions | ✅ | Matches shell grid (§3.6) |
| Right contextual rail | ✅ | Mini-calendar + Today's tasks, both V1-legitimate |
| Page margins / gutters | ✅ | Shared `ScreenLayout` |
| Column widths (two lanes vs. reference's one list) | ✅ (ratified) | G1 — two lanes is the approved deviation, not a gap |
| Vertical alignment (header → toolbar → grid → comparison) | ✅ | Consistent `sp-*` rhythm |

### Header

| Item | Status | Note |
|---|---|---|
| Eyebrow | ✅ | "Timeline" — matches the system-wide eyebrow=screen-name convention used by every V1 screen |
| Page title | ✅ (ratified) | Full date, not a personalised greeting — VP3 (honest, not motivational); already documented as a deliberate divergence in Phase 2 |
| Subtitle | ✅ | Factual, not motivational (VP3) |
| Primary action | ✅ | `SplitButton` "Add block ▾" → Add planned / Add actual — matches reference's "+ Add ▾" pattern closely |
| Date controls | ✅ | `DateStepper` (‹ date › + Today), shared with Daily Review |
| View switcher | ✅ (ratified) | Day / Agenda only — Week/Month excluded (`design-system.md §6.4`, confirmed with product owner this session) |

### Timeline body

| Item | Status | Note |
|---|---|---|
| Time labels | ✅ | Every 3h on the axis, tabular-nums |
| Timeline axis | ✅ | Fixed-width gutter, hour ticks |
| Hour/grid structure | ✅ | Full 00:00–24:00, 24 hourlines |
| Planned/actual distinction | ✅ (ratified) | Lane + dashed(planned)/solid(actual) border — not colour alone (§5 requirement, G1) |
| **Block geometry — text clipping** | 🔧 **CRITICAL, fixed** | See below |
| **Default scroll position** | 🔧 **MAJOR, fixed** | See below |
| Block spacing (gap between adjacent blocks) | ✅ | 4px inset, fine at all tested durations post-fix |
| Category indicators | ✅ (ratified) | `categoryColor(id)` hash — deterministic, presentation-only (D2); see the adjacent finding on real category colour below |
| Typography | ✅ | Tokens only (`--fs-caption`, `--fw-medium`, tabular-nums for times) |
| Icons | 🚫 | Reference shows a per-activity icon (🌱💻🍴…) — no icon field exists on a V1 time block or category; not built (ties to open item C1) |
| Current-time indicator | 🔧 **bug fixed** | Time-pill selector was dead (`:first-of-type` matched by tag, not class — a pre-existing bug from an earlier phase, not from the reference); also **now line can overlap block text** — MINOR, deferred |
| Empty states | ✅ | "Nothing planned" / "Nothing actual" per lane |
| Overflow behaviour (mobile) | ✅ | `.tl2__scroll` (`overflow-x:auto`) — verified by manually scrolling it at 390px; content is fully reachable, not lost |

### Right rail

| Item | V1 status |
|---|---|
| Mini-calendar (date picking) | ✅ in scope, present |
| **Today's tasks** | ✅ in scope (tasks due that date, read-only) — present, hides itself on load/error/empty so it never breaks the screen |
| "Agenda Overview" donut | 🚫 reference-only — dashboard-style chart, excluded (§6.4) |
| "Focus Mode" / Pomodoro | 🚫 reference-only — not V1 (no timers) |
| Motivational quote cards | 🚫 reference-only — VP3/D6, omitted like every other screen |

### Visual system

| Item | Status |
|---|---|
| Colours | ✅ tokens only; category hue via `categoryColor()`, no arbitrary literals found in `timeline.css` |
| Typography scale | ✅ tokens only |
| Spacing | ✅ `--sp-*` tokens throughout |
| Radii | ✅ `--radius-xs`/`--radius-md` |
| Borders | ✅ `--border` / `--border-strong` |
| Shadows | ✅ none used on the grid (correct — a dense data grid shouldn't float); `Card`/`Dialog` use the shared `--shadow-sm` |
| Icon sizing | ✅ 16px nav icons, consistent with the rest of the app |
| Density | 🔧 improved | The text-clipping fix increases how many realistic block durations render legibly without changing the hour scale |

### Responsive

| Viewport | Result |
|---|---|
| 1440×900 | ✅ no h-scroll, all content reachable |
| 1280×800 | ✅ no h-scroll |
| 1024×768 | ✅ no h-scroll, sidebar collapses to icon rail, rail stacks below main |
| 768×1024 | ✅ no h-scroll, rail stacks below main, calendar + Today's tasks both render |
| 390×844 | ✅ no **page** h-scroll; the two-lane grid scrolls **inside** `.tl2__scroll` as designed (verified: scrolling it right reveals the Actual column and all block text, nothing is unreachable) |

*(One screenshot artifact, not a real bug: a `fullPage` Playwright screenshot bakes
`position:fixed` elements — the skip-link, the mobile top bar — in at their
scroll-offset-relative position rather than the true viewport top, because a full-page
capture temporarily expands the viewport. Confirmed by taking a normal (non-fullPage)
screenshot at the same scroll position: neither element appears out of place. No code
change made or needed.)*

### Interaction

| Item | Status |
|---|---|
| Add block (split button → planned/actual) | ✅ |
| Date navigation (‹ ›) | ✅ |
| Today | ✅ (disabled when already on today) |
| Day/Agenda switching | ✅ (`?view=` param) |
| Block click → edit | ✅ (`BlockDialog`) |
| Delete | ✅ (in the edit dialog) |
| Midnight-spanning blocks | ✅ ▲/▼ markers verified with a live Sleep-style fixture block |

---

## Fixes made in this pass

### 1. CRITICAL — block category label clipped/invisible on short blocks

**Symptom** (found live, with fixture blocks of realistic durations — 30, 45, 60+ min):
a block's two-line stacked layout (time on one line, category on the next) needs ~40px
of height to avoid `overflow:hidden` clipping the second line. At the existing
40px/hour scale, that means **any block of ~50 minutes or less** had its category name
partially or **completely** invisible — only the time range remained legible. This
affected 5 of the 15 fixture blocks used for this audit, including exact analogues of
the reference's own "Team Sync" (45 min) and "Morning Routine" (45 min).

**Fix**: `web/src/features/timeline/TimelineGrid.tsx` + `web/src/styles/timeline.css`
— the block now renders time and category on **one line** (`06:00–06:45 · Personal`),
with the category name ellipsizing if truly out of room, instead of a stacked
two-line layout. Single-line content needs ~21px, comfortably under the block's
existing minimum-height floor (2.4% of the day ≈ 23px), so every duration renders
legibly, including the shortest blocks the grid allows.

### 2. MAJOR — default scroll position was empty midnight

**Symptom**: the grid always renders the full 00:00–24:00 range (G1, correct), but the
page loaded scrolled to the very top — 6 empty hours before the first realistic block
of the day. A user had to scroll past dead space to see anything.

**Fix**: `TimelineGrid` now scrolls the page to a useful anchor after loading —
"now" for today, the earliest block's start time for a non-today date with blocks, or
06:00 as a last-resort default — with an hour of lead-in context. This doesn't change
the range or add a new UI affordance; it only picks a better starting scroll position,
matching the reference's apparent effect of opening onto the populated part of the day.

### 3. MINOR, deferred — "now" line can cross an active block's text

The current-time line renders above blocks (`z-index:3`) so it stays visible when
"now" falls inside an active block — which happens often, since checking your timeline
while something is in progress is a common case. When that happens, the thin red line
visually crosses the block's text. Making the line duck *under* blocks isn't a good
fix either — blocks have a semi-opaque tint, so the line would simply disappear
whenever it matters most (you're inside an active block right now). A correct fix
needs the current-time indicator split into independently-stacked pieces (a line that
can duck under blocks, a time-pill that stays on top), which is a larger, riskier
change than this pass's other two fixes. Recorded here rather than attempted blind;
the block's own border/fill still clearly identifies it even with the line crossing it.

### Also fixed, found via the same audit (pre-existing, unrelated to the reference)

`.tl2__lane:first-of-type` never matched (it matches by tag, not class — `.tl2__lane`
is not actually the first `<div>` child of `.tl2__grid`), so the current-time pill's
`10:24`-style time label had never rendered since it was introduced. Fixed by adding
explicit `tl2__lane--planned`/`tl2__lane--actual` classes and retargeting the
selector. *(This was already fixed and documented in an earlier session pass —
re-verified live in this audit and still correct.)*

---

## Phase 2 — Backend audit

**Result: no gap.** Timeline Day's full data need is already served by real,
account-scoped, tested backend code — confirmed by reading (not assuming)
`internal/timeline/{timeline,block,http,service,readmodel}.go`:

| Requirement | Backend support |
|---|---|
| Planned blocks: create/edit/delete/list for a date (§3) | `Service.AddBlock` / `EditBlock` / `DeleteBlock` / `Timeline` — all account-scoped |
| Actual blocks, same operations (§4) | Same `Service`, `Kind` field distinguishes planned/actual, immutable after create |
| Timeline view for a chosen date (§5) | `Service.Timeline(ctx, accountID, date)` resolves in the account's IANA timezone (`AccountZone`) |
| Midnight-spanning blocks (§5) | Handled — `PositionedBlock.from_prev_day`/`to_next_day`/`ends_next_day` |
| Planned-vs-actual comparison (§6) | `Service.Comparison` — per-category planned/actual/difference, deterministic |
| Category association | `CategoryStore` interface — validates a category is the caller's and active before assigning |
| Account isolation | Every `Service` method takes `accountID` from request context only (ADR-0002/0004) |
| A range-comparison endpoint (`getComparisonRange`, new) | Exists, not needed by Day (Weekly Review's concern) |

**No fixture/mock data was used for anything Timeline Day actually needs.** All
screenshots and functional checks in this pass ran against the real API with a
throwaway QA account (`timeline-day-audit@example.com`, deleted after the audit) and
real created categories/blocks/tasks — never invented client-side data.

### Adjacent finding — out of scope for this task, recorded for visibility

`v1.md §2` was amended (ADR-0009): **a category now has a colour and an icon**, and
the backend fully implements this — `internal/categories/{categories,http,service}.go`
store and return real `colour`/`icon` string fields (a key from a fixed client-side
set; the backend never interprets them). **The frontend has not been updated at all**:
`web/src/api.ts`'s `Category` interface is still `{id, name}`, and nothing renders a
category's real colour or icon anywhere — Timeline blocks are coloured by
`categoryColor(id)`, a deterministic **hash**, not the user's actual chosen colour.

This is real, backend-ready data the frontend is ignoring — exactly the kind of thing
this task's Phase 2 instructions ask to flag. It is **not fixed here** because doing so
properly means: extending `api.ts`'s `Category` type, deciding the fixed colour-key →
CSS-token mapping and the icon set (a Categories-screen design decision, not a Timeline
one), and updating the Categories screen's create/edit form to offer them — none of
which is "Timeline Day," and CLAUDE.md/this task both say not to touch unrelated
screens. Recommend a follow-up task scoped to Categories + a shared colour/icon
picker, after which Timeline's block colour source would swap from the hash to the
real value with no change to Timeline's own code beyond one line in `categoryColor()`
or its call site.

---

## Phase 3 — Reference/fixture data used for this audit

V1-valid only — categories (a name, nothing else) and blocks (kind/start/end/category,
nothing else). No titles were used on blocks because **time blocks have no title field
in V1** (`v1.md §3/§4` — a block is start, end, and an optional category; the
reference's per-block activity names like "Morning Routine" or "DSA Practice" are the
mock's own invention, not a V1 field). The closest faithful approximation is to give
blocks the *category* that activity would plausibly carry:

| Category created | Approximates reference activities |
|---|---|
| Personal | Morning Routine, Lunch Break, Plan Tomorrow, Read a Book |
| Study | DSA Practice, Read Research Papers |
| Project | Work on Productivity OS, Build Timeline UI |
| Work | Team Sync |
| Health | Workout, Sleep |

10 planned + 5 actual blocks were created (including one midnight-spanning actual
"Sleep" block, 22:00→06:00) via the real Add-block dialog against the real API, plus
two real tasks due today (one completed) to populate the Today's-tasks rail widget.
All fixture data belonged to a throwaway account and was not left behind — the account
was deleted after the audit (`DELETE FROM accounts WHERE email =
'timeline-day-audit@example.com'`).

No unsupported metadata was used: no assignees, no tags, no priorities, no
per-block "done" state, no Focus/Pomodoro, no generic "events".

---

## Phase 6 — Functional QA

| Check | Result |
|---|---|
| Day view | ✅ renders both lanes, all 15 fixture blocks in the right lane/position |
| Agenda view | ✅ switching via `?view=agenda` still works (not touched by this pass's fixes — spot-checked, not regressed) |
| Previous / next date | ✅ |
| Today | ✅ disabled on today, jumps back from another date |
| Add block (planned + actual, via the split-button menu) | ✅ |
| Existing blocks render | ✅ real API, not fixtures |
| Timezone/date handling | ✅ blocks land on the correct date; midnight-spanning block shows ▼ correctly |
| Account isolation | Not independently re-tested this pass (unchanged code path, already covered by backend tests) |
| Loading state | ✅ "Loading…" text while the three calls are in flight |
| Empty state | ✅ "Nothing planned"/"Nothing actual" per lane, confirmed on a date with no blocks |
| Error state | Not re-triggered this pass (unchanged `ErrorState` + Retry path from Phase 2) |

---

## Phase 7 — Completion

- **Typecheck**: `pnpm run typecheck` — clean.
- **Build**: pending final run below (post-fix).
- **Tests**: pending full-suite run below (post-fix) — the two changed files
  (`TimelineGrid.tsx`, `timeline.css`) are covered by `TimelineGrid.test.tsx` and
  `TimelineScreen.test.tsx`.
- **Backend tests**: not run — no backend files were changed by this task.
- **Playwright**: 5 viewports captured and inspected (1440×900, 1280×800, 1024×768,
  768×1024, 390×844); no page-level horizontal scroll at any size; mobile's
  intentional inner horizontal scroll verified to reach all content.

### Remaining visual deviations (accepted, not gaps)

- Two lanes instead of one list (G1, ratified).
- Full date title instead of a greeting (system-wide `PageHeader` convention).
- No per-block icon (no icon field on a V1 time block).
- "Now" line can cross an active block's text (MINOR, deferred — see above).

### Remaining V1 decisions / blockers

- None block Timeline Day. The category-colour/icon frontend gap (above) is a
  cross-screen follow-up, not a Timeline blocker.

### Files changed

```
web/src/features/timeline/TimelineGrid.tsx   — single-line block content; auto-scroll to a useful anchor
web/src/styles/timeline.css                  — .tl2__block-line layout (replaces the stacked column layout)
docs/design/screens/timeline-implementation-plan.md   — this file (new)
```

No other screens were modified.
