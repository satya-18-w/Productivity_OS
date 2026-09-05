# Categories — reference-vs-actual gap plan

**Date:** 2026-09-04
**Reference:** `docs/design/references/categories.png`
**Actual:** `/tmp/opencode/pg-categories.png` (`CategoriesScreen` in `web/src/features/categories/`)
**Scope anchor:** task-fixed V1 — flat active list, create/rename/archive (**name only**).
Spec: `docs/design/screens/categories.md` (+ Phase 8 status note, already COMPLETE).

> Note: `docs/requirements/v1.md` §2 has since been amended with ADR-0009
> (colour/icon keys, cross-entity assignment, per-category counts). This plan
> deliberately implements the **task-fixed** scope (pre-amendment §2), which is
> what `screens/categories.md` and the live screen were built against. Anything
> from the amendment is recorded under "Backend gaps noticed" and **not** built.

---

## 1. Reference vs actual

| Reference element | Actual | Verdict |
|---|---|---|
| Eyebrow "CATEGORIES" + H1 + subtitle; primary "＋ New Category" top-right | Eyebrow + H1 "Categories" + subtitle "Flat labels for your time blocks." + "New category" button | ✅ match |
| Category tab strip (All / Personal / Work / …) | None — flat list, no tabs | ✅ deliberate (spec: named tabs are sample data; filter-by-self strip is odd) |
| 8 category cards: icon tile, name, "N items", Tasks/Notes/Habits/Goals/Events breakdown, dot row, arrow | Plain rows: decorative colour tile (first-letter glyph) + name + kebab | ✅ deliberate (counts/breakdowns are not V1 — §2 time-block-only; Notes/Events don't exist) |
| Kebab → Rename / Archive (archive confirmed) | `Menu` with Rename · Archive + `window.confirm` on archive | ✅ match (`window.confirm` is the established cross-screen pattern — tasks/goals/board/habits all use it) |
| New / rename form, name only | `CategoryDialog`, name only, 409 duplicate-name + validation errors wired | ✅ match |
| Right rail: "Category Overview" donut (94 items) | No rail | ✅ deliberate (depends on cross-entity counts — not V1) |
| Right rail: Quick Actions (New / Manage / Set Colors / Import / Export) | None | ✅ deliberate (category import/export is not V1 — §14 is one full snapshot) |
| Right rail: Recently Used Categories | None | ✅ deliberate (not a V1 concept) |
| Bottom banner "Categories bring clarity." + header illustration + title icon badge | None | ✅ deliberate (D6: omit decorative surfaces when unsure; `PageHeader` deliberately omits illustration) |
| Loading / empty / error states | Loading text, `EmptyState`, `ErrorState` with retry | ✅ match (`<p class="muted">Loading…</p>` is the shared cross-screen pattern) |
| Active / Archived segmented tabs | None — active list only | ✅ deliberate (no backend to list archived or unarchive — C1 open, see §3) |
| Responsive single column, no page h-scroll; light + dark | Single-column list, token-driven, dark tokens present | ✅ match (verified in Phase 8) |

## 2. Deliberate deviations kept (with justification)

1. **No item counts / breakdowns** — pre-amendment §2: a category may be referenced by a time block only; even block counts aren't a stated requirement. (`screens/categories.md` V1 table.)
2. **Colour is presentation-only decoration** — D2; `categoryColor(id)` hash → palette hue, first-letter glyph. Nothing persisted, nothing depends on it.
3. **No icons, no colour picker, no import/export, no recently-used, no donut, no banner/illustration** — all reference-only per design-system §6.4 / D6.
4. **No Archived tab / unarchive** — §2 guarantees archive only; API has no archived-list or unarchive (unlike Habits). Tracked in `docs/left.md` Phase 8 + register item C1.
5. **`window.confirm` for archive + `<p class="muted">Loading…</p>`** — kept for consistency: identical patterns in Tasks/Goals/Board/Habits screens. Changing one screen alone would fork a second pattern (design-system rule 4).

## 3. Gaps to build (Phase B — all inside `web/src/features/categories/` + `web/src/styles/categories.css`)

| # | Gap | File | Fix |
|---|---|---|---|
| G1 | Raw `12px` vertical padding — token scale (`--sp-1…--sp-7`) must be the only spacing source (design-system §5 rule 1) | `web/src/styles/categories.css` | `padding: 12px var(--sp-4)` → `padding: var(--sp-3) var(--sp-4)` (12px == `--sp-3`, no visual change) |
| G2 | Dead first-row rule: `.category-list li:first-child { border-top: 0; }` targets the `li`, but the border lives on `.category-row` — first row keeps a doubled top border against the container | `web/src/styles/categories.css` | `.category-list > li:first-child > .category-row { border-top: 0; }` |
| G3 | Dialog submits the raw name: whitespace-only input passes `required` and round-trips to a 422; padded names (`" Work "`) are sent untrimmed while the backend trims server-side (`internal/categories/service.go` `validate`) | `web/src/features/categories/CategoryDialog.tsx` | Trim before submit; disable Create/Save while trimmed name is empty. Tests: trim assertion + blank-disabled. |

Out of scope for Phase B (explicitly excluded): counts, donut, icons/colour persistence,
import/export, recently-used, Archived tab, header illustration/banner.

## 4. Backend gaps noticed (list only — never implement from this task)

1. **No archived-categories listing / unarchive** (`GET /api/categories?state=archived` or similar + `POST /api/categories/:id/unarchive`) — C1 open; already tracked in `docs/left.md` Phase 8 (swap point: `CategoriesScreen.tsx` + `SegmentedControl` tab).
2. **Rename wipes ADR-0009 colour/icon**: `PATCH /api/categories/{id}` decodes a full `categoryRequest` ("empty clears them"), but the frontend sends `{name}` only — so renaming a category that has a stored colour/icon would clear both. Either the frontend must echo colour/icon (needs `Category` type fields — out of this scope) or PATCH should treat omitted keys as unchanged. Flag for the categories owner.
3. **Frontend `Category` type omits `colour`/`icon`** though `GET /api/categories` now returns them — deliberate under this task's scope; revisit if ADR-0009 fields reach the UI.
4. **`GET /api/categories/overview` exists** (`cmd/server/overview.go`) but the frontend deliberately doesn't call it — per-category counts are not V1 in this task's scope.
