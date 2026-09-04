---
name: productivity-ui
description: Apply the Productivity OS visual language, design system, reusable UI architecture, and interaction principles when building or reviewing frontend UI.
disable-model-invocation: true
---

# Productivity OS UI System

## Purpose

Use this skill when implementing or modifying the Productivity OS frontend.

This skill defines HOW the product UI should be built.

The detailed visual values are NOT stored here.

The source of truth for detailed design values is:

- docs/design/design-system.md
- docs/design/visual-principles.md

Read only the relevant sections needed for the current task.

Do not load the entire design documentation into context unless explicitly required.

---

# Product UI Philosophy

Productivity OS should feel:

- calm
- professional
- human
- focused
- trustworthy
- organized
- low-friction
- nature-inspired without becoming decorative

The interface should reduce cognitive load rather than compete for attention.

Avoid:

- excessive gradients
- excessive animations
- excessive glassmorphism
- unnecessary decoration
- visual noise
- aggressive gamification
- excessive saturated colors
- inconsistent component styling
- unnecessary badges
- unnecessary shadows
- UI elements that exist only for visual novelty

The product should feel like a serious productivity tool designed for long-term daily use.

---

# Visual Source of Truth

When a screen has a reference image:

1. Read the corresponding screen specification.
2. Inspect the reference image.
3. Read only the relevant design-system sections.
4. Reuse existing components.
5. Reuse existing design tokens.
6. Implement according to the reference.
7. Verify the rendered result in a browser.

Never treat the reference image as loose inspiration.

Treat it as a visual specification.

---

# Design System Rules

Always use the existing design system for:

- colors
- typography
- spacing
- radii
- borders
- shadows
- component dimensions
- breakpoints
- interaction states

Do not invent a new visual value merely because it looks convenient.

If a required value genuinely does not exist:

1. determine whether an existing token can be reused;
2. if not, document the new token in the design system;
3. then use it consistently.

Never create one-off values silently.

---

# Color Rules

The visual system uses a restrained palette.

Prefer semantic roles rather than raw colors.

Examples:

- page background
- surface
- elevated surface
- primary text
- secondary text
- muted text
- primary action
- success
- warning
- danger
- informational
- category accents

Do not introduce arbitrary colors directly into components.

Do not use different shades of the same semantic color across screens without a documented reason.

Color should communicate hierarchy and meaning, not decoration.

---

# Layout Rules

Prefer:

- CSS Grid
- Flexbox
- normal document flow

Avoid absolute positioning for primary layout.

Absolute positioning may be used only when the element genuinely requires positional anchoring.

Examples where absolute positioning may be appropriate:

- popovers
- tooltips
- badges anchored to an element
- overlays
- decorative elements
- floating controls

Do not use absolute positioning to compensate for an incorrect layout structure.

---

# Alignment Rules

Every screen must have a deliberate alignment system.

Check:

- page container alignment
- sidebar alignment
- topbar alignment
- section alignment
- card alignment
- text alignment
- icon alignment
- button alignment
- form alignment

Elements that belong to the same visual group should share consistent alignment.

Do not fix alignment problems with arbitrary margins.

Correct the underlying layout.

---

# Component Rules

Before creating a component:

1. Search for an existing reusable component.
2. Determine whether the existing component can support the requirement.
3. Extend the existing component when appropriate.
4. Only create a new component when the behavior or visual responsibility is genuinely distinct.

Preferred architecture:

components/
├── ui/
├── layout/
└── productivity/

Shared components must remain reusable.

Do not create page-specific copies of global components.

---

# Component Responsibility

Components should have clear responsibilities.

Prefer:

- Button handles button behavior.
- Card handles surface structure.
- Dialog handles modal behavior.
- Sidebar handles navigation.
- TimelineBlock handles timeline block presentation.
- TaskCard handles task presentation.
- HabitRow handles habit presentation.

Avoid components that become giant containers responsible for unrelated behavior.

---

# Data vs Presentation

Keep visual components as independent from backend details as reasonably possible.

Prefer:

data/state
    ↓
page/container
    ↓
presentational components

Do not put API-specific logic into low-level UI primitives.

---

# Interaction Rules

Interactions should be:

- predictable
- discoverable
- responsive
- visually consistent

Use existing interaction patterns.

Do not introduce custom interaction behavior when an existing component already solves the problem.

Important interaction states include:

- default
- hover
- focus
- active
- disabled
- loading
- error
- empty

---

# Accessibility

UI must remain usable with:

- keyboard navigation
- visible focus states
- semantic HTML
- appropriate labels
- accessible interactive controls

Do not rely only on color to communicate state.

Icons should not replace accessible labels when the meaning is ambiguous.

---

# Context Discipline

This skill is intentionally lightweight.

DO NOT:

- read the entire repository
- read unrelated backend code
- read every design reference
- read every screen specification
- inspect unrelated frontend components
- load the complete project roadmap
- repeatedly reread CLAUDE.md

For a UI task, normally read only:

1. CLAUDE.md if not already available
2. the relevant screen specification
3. relevant design-system sections
4. relevant existing components
5. the relevant reference image

Do not expand context unless the current task requires it.

---

# Change Discipline

When modifying an existing screen:

1. inspect the current implementation;
2. identify the smallest relevant set of files;
3. make focused changes;
4. avoid unrelated refactoring;
5. verify the result.

Do not redesign unrelated screens while implementing one screen.

---

# Completion Requirement

A UI change is not complete merely because the code compiles.

For visual changes:

- application must run;
- relevant route must render;
- browser verification must be performed;
- layout must be checked;
- responsive behavior must be checked;
- visual deviations must be reviewed.

Use the `visual-qa` skill when visual verification is required.

Use the `responsive-ui` skill when responsive behavior is part of the task.