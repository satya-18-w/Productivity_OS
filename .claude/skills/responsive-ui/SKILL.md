---
name: responsive-ui
description: Implement and verify responsive behavior for Productivity OS across desktop, tablet, and mobile viewports.
disable-model-invocation: true
---

# Productivity OS Responsive UI

## Purpose

Use this skill when implementing or reviewing responsive behavior.

The goal is not to make the desktop UI smaller.

The goal is to create an intentional responsive layout.

---

# Context Discipline

Only inspect:

- current screen specification
- current reference image
- responsive design rules
- relevant components
- current route
- relevant styles

Do not inspect the entire frontend.

Do not inspect unrelated screens.

Do not read unrelated backend code.

Do not load every reference image.

---

# Responsive Principle

Responsive behavior should preserve:

- information hierarchy
- usability
- visual rhythm
- accessibility
- interaction clarity

while adapting layout to available space.

---

# Layout Strategy

Prefer:

- CSS Grid
- Flexbox
- fluid widths
- max-width containers
- responsive gaps
- responsive typography where necessary

Avoid:

- fixed widths that cause overflow
- desktop-only assumptions
- unnecessary JavaScript viewport detection
- excessive breakpoint-specific duplication
- absolute positioning for primary layout

---

# Responsive Transformations

When moving from desktop to smaller screens, intentionally decide:

## Navigation

Desktop may use:

- persistent sidebar

Smaller screens may use:

- collapsed navigation
- drawer
- compact navigation

Do not simply force the desktop sidebar into a narrow viewport.

---

## Content

Desktop may use:

- multi-column layout

Tablet may use:

- fewer columns

Mobile should generally use:

- single-column content

unless the content genuinely requires otherwise.

---

## Cards

Cards should:

- shrink naturally
- maintain readable internal spacing
- avoid text collision
- avoid fixed heights when content is dynamic

Avoid forcing cards into fixed dimensions when the content can vary.

---

## Tables

For wide data:

prefer an intentional responsive strategy such as:

- horizontal scrolling inside the table container
- alternate mobile representation
- reduced columns where appropriate

Do not allow the entire page to horizontally overflow accidentally.

---

## Forms

Forms should adapt:

Desktop:
- multiple fields may share a row

Mobile:
- fields may stack vertically

Controls must remain easy to tap and read.

---

# Breakpoint Discipline

Do not create breakpoints arbitrarily.

Use the project's documented breakpoint tokens.

If a new breakpoint is genuinely required:

1. verify existing breakpoints cannot solve the problem;
2. document the reason;
3. add it to the design system;
4. use it consistently.

Do not create one-off breakpoints for individual elements without justification.

---

# Viewport Verification

At minimum verify:

## Desktop

1440 × 900

## Tablet

1024 × 768

768 × 1024

## Mobile

390 × 844

These are baseline verification sizes, not mandatory permanent product breakpoints.

---

# Responsive QA Checklist

For every viewport check:

### Layout

- no overlap
- no clipping
- no unintended horizontal overflow
- no broken grid
- no broken flex layout

### Navigation

- navigation remains usable
- active state remains visible
- controls remain accessible

### Typography

- headings remain readable
- text does not clip
- line lengths remain reasonable
- hierarchy remains clear

### Cards

- cards fit viewport
- internal content fits
- buttons do not collide
- content does not overflow

### Forms

- inputs fit
- labels remain visible
- controls remain usable
- errors remain visible

### Dialogs

- dialog fits viewport
- content can scroll when necessary
- close action remains accessible

### Interactions

- touch-friendly controls
- keyboard navigation
- visible focus
- no hover-only critical interaction

---

# Mobile Is Not An Afterthought

Do not implement:

desktop
    ↓
shrink everything
    ↓
mobile

Instead implement:

shared design system
    ↓
responsive layout model
    ↓
desktop composition
    ↓
tablet composition
    ↓
mobile composition

The information hierarchy should remain consistent even when the layout changes.

---

# Visual Verification

Use Playwright to:

1. open the route;
2. set viewport;
3. inspect the rendered page;
4. capture a screenshot;
5. check for overflow/collision;
6. interact with important controls;
7. repeat for each required viewport.

Do not claim responsive completion based only on CSS inspection.

---

# Fix Priority

When responsive problems exist, fix in this order:

1. application unusable
2. content overlap
3. horizontal overflow
4. clipped controls/content
5. broken navigation
6. broken grid/columns
7. typography problems
8. spacing problems
9. cosmetic details

---

# Scope Discipline

Do not redesign desktop while solving a mobile issue.

Do not modify unrelated screens.

Do not add a responsive library unless explicitly approved.

Do not introduce product behavior that is not specified.

If the correct responsive behavior cannot be inferred from the specification or design system, stop and report the ambiguity.