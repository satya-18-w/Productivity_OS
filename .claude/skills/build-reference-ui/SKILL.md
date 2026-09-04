---
name: build-reference-ui
description: Implement a Productivity OS frontend screen accurately from a visual reference, screen specification, and shared design system.
disable-model-invocation: true
---

# Build Reference UI

## Purpose

Use this skill when implementing a frontend screen from a supplied reference image.

The goal is not to create a UI that is merely similar.

The goal is to reproduce the intended:

- hierarchy
- geometry
- spacing
- alignment
- typography
- color system
- component structure
- interaction model
- responsive behavior

while maintaining the existing Productivity OS architecture.

---

# Required Inputs

Before implementation, identify:

1. Reference image
2. Screen specification
3. Existing design system
4. Existing reusable components
5. Existing application route

Expected locations:

Reference:

docs/design/references/<screen>.png

Specification:

docs/design/screens/<screen>.md

Design system:

docs/design/design-system.md

Visual principles:

docs/design/visual-principles.md

---

# Context Budget

Do not load the entire project into context.

For the current screen, inspect only:

- relevant screen specification
- relevant reference image
- relevant design-system sections
- relevant components
- relevant route/page
- relevant styles

Do not read unrelated:

- backend modules
- database code
- other screen implementations
- historical decisions
- unrelated documentation

unless the current implementation genuinely depends on them.

The reference image does not require reading every other reference image.

---

# Phase 1 — Understand

Before editing code, identify:

## Page structure

Determine:

- global shell
- sidebar
- topbar
- page header
- primary content area
- secondary content
- cards
- controls
- footer if present

## Visual hierarchy

Determine:

- primary heading
- secondary headings
- supporting text
- primary action
- secondary actions
- important data
- supporting data

## Geometry

Determine:

- container width
- column structure
- card dimensions
- gaps
- padding
- alignment
- vertical rhythm

Do not guess blindly.

Use the reference image and existing design tokens.

---

# Phase 2 — Inspect Existing Code

Before creating components:

1. inspect the route;
2. inspect existing layout components;
3. search for reusable UI primitives;
4. search for existing productivity components;
5. inspect existing design tokens.

Do not create duplicate components.

If an existing component is close to the required design:

- reuse it;
- extend it;
- or refactor it carefully.

Only create a new component when necessary.

---

# Phase 3 — Implementation Plan

Before making substantial changes, establish a short implementation plan.

The plan should identify:

- files to modify
- components to reuse
- components to create
- layout strategy
- responsive strategy
- verification strategy

Do not perform unrelated refactoring.

---

# Phase 4 — Implement

Implement the screen using:

- React
- TypeScript
- existing design tokens
- existing component architecture
- CSS Grid
- Flexbox
- normal document flow

Avoid:

- arbitrary colors
- arbitrary spacing
- arbitrary typography
- arbitrary shadows
- arbitrary radii
- excessive absolute positioning
- duplicate components

---

# Layout Integrity

The following are hard requirements:

- no overlapping elements
- no accidental clipping
- no unexpected horizontal scrolling
- no content escaping containers
- no broken grid/flex behavior
- no text overflowing controls
- no buttons colliding with other controls
- no inconsistent card alignment

Never solve structural layout problems with random margins.

Fix the layout model.

---

# Visual Accuracy

Compare implementation against the reference in this order:

## 1. Macro layout

Check:

- sidebar
- topbar
- page container
- columns
- major sections

## 2. Component geometry

Check:

- cards
- buttons
- inputs
- tabs
- rows
- blocks

## 3. Spacing

Check:

- page padding
- section gaps
- card gaps
- internal padding
- text spacing

## 4. Typography

Check:

- heading sizes
- weight
- line height
- hierarchy
- text density

## 5. Color

Check:

- background
- surfaces
- primary color
- semantic accents
- borders
- muted text

## 6. Details

Check:

- icons
- radii
- borders
- shadows
- hover states
- focus states

Always fix higher-level structural deviations before small visual details.

---

# Browser Verification

After implementation:

1. start the application;
2. open the relevant route using Playwright;
3. verify the page actually renders;
4. inspect the layout;
5. interact with important controls;
6. capture a screenshot;
7. compare it with the reference;
8. identify deviations;
9. fix the highest-impact deviations;
10. repeat.

Do not claim visual completion without browser verification.

---

# Screenshot Comparison

When comparing the implementation to the reference, prioritize:

### Critical

- overlapping elements
- broken layout
- incorrect page structure
- missing major components
- unusable interactions
- severe responsive breakage

### Major

- incorrect dimensions
- major alignment differences
- incorrect spacing
- incorrect typography hierarchy
- significant color differences

### Minor

- small icon differences
- subtle shadow differences
- tiny spacing differences
- small radius differences

Fix in that order.

---

# Responsive Implementation

Do not implement desktop first and assume everything else will work.

Use the responsive-ui skill when responsive behavior is part of the screen.

The layout should adapt intentionally.

Typical transformations include:

desktop:
sidebar + multi-column content

tablet:
reduced columns + compressed navigation

mobile:
single-column content + mobile navigation strategy

Do not simply shrink desktop components until they fit.

---

# Interaction Verification

Verify relevant interactions such as:

- navigation
- buttons
- tabs
- dropdowns
- dialogs
- forms
- task completion
- habit completion
- timeline interactions

A visually accurate component that does not behave correctly is incomplete.

---

# No Scope Creep

While implementing a screen:

DO NOT:

- redesign other screens
- introduce new product features
- modify backend architecture
- add unnecessary libraries
- add unrelated animations
- add speculative functionality
- redesign the entire design system

If a requirement is missing:

stop and identify the missing decision.

Do not invent product behavior.

---

# Dependency Discipline

Do not add a frontend dependency merely to solve a small visual problem.

Before adding a dependency:

1. check existing dependencies;
2. determine whether the requirement can be implemented using existing tools;
3. determine whether the dependency materially improves the architecture;
4. obtain explicit approval before introducing an unplanned dependency.

---

# Verification Before Completion

Run the relevant:

- frontend build
- tests
- browser verification
- responsive checks
- visual QA

Inspect:

- git diff
- git status

Ensure unrelated files were not modified.

Do not commit automatically.

---

# Completion Statement

Only report the screen as complete when:

- implementation is complete;
- build passes;
- required tests pass;
- browser verification passes;
- no layout overlap exists;
- no unexpected overflow exists;
- reference comparison has been performed;
- responsive behavior has been checked where required;
- no unrelated changes were introduced.