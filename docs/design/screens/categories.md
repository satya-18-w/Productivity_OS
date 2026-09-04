# Screen — Categories

**Reference:** `docs/design/references/categories.png` (also panel 7 of `overall.png`)
**Purpose:** Create, rename, and archive categories; see the active list.
**Proposed route:** `/categories`

---

## V1 scope alignment

Maps to `requirements` §2 (categories).

| Reference element | V1 status |
|---|---|
| List/grid of categories with a name | in scope — create, rename, archive, list active |
| **"N items" + breakdown: Tasks / Notes / Habits / Goals / Events counts** | ⚠ **not V1.** A category in V1 may be referenced by **a time block only** ("Tasks, habits, and goals carry no category in V1"; Notes/Events don't exist). The only meaningful count is *planned/actual blocks referencing this category* — and even that isn't a stated requirement. |
| Per-category **colour** (icon tile, dot rows) | allowed as decoration — "a frontend may show a colour; nothing depends on it" (§2). Not a stored product attribute with meaning. |
| Category **icon / glyph** | not a V1 field — decorative only |
| Category **tabs** (All / Personal / Work / …) filtering by category | the named categories are sample data; a filter-by-self tab strip is odd for a management screen — likely just the "All" state matters |
| "＋ New Category" | in scope |
| Archive (implied via kebab) | in scope — §2: archive so it's no longer offered, without changing existing assignments. **No rename of the archive; no hard delete; no merge.** |
| Rail: "Category Overview" donut (94 items by category) | not V1 (depends on cross-entity counts) |
| Rail: "Quick Actions" (New / Manage / Set Colors / Import / Export) | Import/Export of categories alone is not V1 (export is a single full snapshot, §14) |
| Rail: "Recently Used Categories" | not a V1 concept |

**Recommendation:** build a simple management surface — a list (or light grid) of
**active** categories, each with **name**, an optional decorative colour swatch, and
actions **rename** / **archive**. A separate view or tab for **archived** categories with
**unarchive** (if unarchive is desired — §2 only guarantees archive; confirm). **Drop:**
item counts, cross-entity breakdowns, icons, donut, import/export, recently-used.

---

## Layout

- Shell (§4.1). Header: eyebrow "CATEGORIES" + H1 + subtitle; title icon badge; header
  illustration.
- Right: **primary** "＋ New Category".
- Optional segmented control: **Active / Archived**.
- **Body:** a grid of **category cards** (or a plain list — a list is closer to §2's "flat
  list"):
  - card: colour swatch / glyph tile · **name** · kebab (Rename / Archive).
  - the reference's counts/breakdown block is **omitted** for V1.
- Right rail: minimal — a quote card, or nothing. (No donut / quick-actions / recents.)

## Screen-specific components

- **Category card / row** — glyph-or-swatch · name · kebab.
- **New / rename form** (§4.16) — field: **name** only. Whether a category stores a colour
  is **unresolved** (`design-system.md` register item C1); until ratified, treat any
  colour as a client-side presentation choice, not a persisted product attribute (§2, D2).
- **Archived list** — same rows with an "Unarchive" action **if** unarchive is confirmed
  to exist (C1 — §2 only guarantees *archive*).

## Interactions

- "＋ New Category" → name form. Kebab → rename (inline or form) / archive.
- Archiving removes the category from assignment pickers elsewhere but leaves existing
  block→category references intact (§2).

## Responsive

- Grid 4→2→1, or a single-column list throughout. Right rail drops first.

## Unresolved — deferred to a product requirement (`design-system.md` register item C1)

- Whether categories can be **unarchived** (§2 only states archive).
- Whether a category **persists a colour**, or colour is purely client-side presentation
  (§2: nothing depends on it; D2 forbids logic on colour).
- Whether **"item counts"** are shown at all, and if so only *time-block* usage (the one
  real V1 relationship).
- The sidebar **"Spaces"** concept (category-as-workspace) — not built until C1.

## Design-system references

§4.1 shell · §4.2 header · §4.3 view switcher (Active/Archived) · §4.5 buttons · §4.7 card
· §4.9 chips · §4.16 create/edit form · `requirements` §2 · `visual-principles.md` VP10.
