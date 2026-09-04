---
name: visual-qa
description: Perform browser-based visual quality assurance for Productivity OS screens against their reference images and design system.
disable-model-invocation: true
---

# Productivity OS Visual QA

## Purpose

Use this skill to inspect an implemented frontend screen.

This is a QA skill.

Do not redesign the application during visual QA.

Do not modify production code while performing a QA-only pass.

The goal is to identify deviations objectively.

---

# Context Discipline

Do not inspect the entire project.

Read only:

1. relevant screen specification;
2. relevant reference image;
3. relevant design-system sections;
4. relevant route/page implementation if necessary.

Do not read unrelated screens or backend modules.

Do not load every reference image.

Do not compare unrelated screens.

---

# Browser Requirement

Visual QA must use the actual running application.

Prefer Playwright for:

- navigation
- screenshots
- viewport changes
- interaction
- accessibility inspection
- console/network inspection when useful

Do not judge visual quality solely from source code.

Source code describes intent.

The browser rendering is the result that matters.

---

# QA Sequence

Perform QA in this order:

1. application availability
2. route correctness
3. macro layout
4. component geometry
5. spacing
6. typography
7. colors
8. alignment
9. interaction states
10. responsive behavior
11. overflow/clipping
12. accessibility basics

---

# 1. Application Availability

Verify:

- application starts
- route loads
- no fatal runtime error
- no blank page
- no unexpected error screen

If the page cannot render, classify the issue as CRITICAL.

---

# 2. Route Correctness

Verify:

- expected route opens
- navigation works
- page title/header is correct
- correct screen is rendered

---

# 3. Macro Layout

Compare the reference and rendered screen.

Inspect:

- overall page width
- sidebar
- topbar
- main content
- major sections
- columns
- section ordering
- container alignment

Macro-layout problems take priority over cosmetic problems.

---

# 4. Geometry

Inspect:

- card dimensions
- button dimensions
- input dimensions
- row heights
- timeline blocks
- navigation dimensions
- dialog dimensions
- chart areas

Look for:

- overlap
- clipping
- unexpected stretching
- unexpected shrinking
- misaligned edges
- inconsistent dimensions

---

# 5. Spacing

Check:

- page padding
- section spacing
- card gaps
- card internal padding
- text spacing
- button spacing
- icon/text spacing

Look for inconsistent rhythm.

Do not recommend random pixel changes without understanding the underlying spacing system.

---

# 6. Typography

Check:

- font family
- font size
- font weight
- line height
- heading hierarchy
- body text density
- muted text
- labels
- numerical data

Watch for:

- text wrapping unexpectedly
- clipped text
- headings competing with primary content
- inconsistent hierarchy

---

# 7. Color

Compare semantic roles:

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
- category colors

Do not judge only whether a color is "nice."

Judge whether it is consistent with the established design system.

---

# 8. Alignment

Inspect:

- left edges
- right edges
- baselines
- icons
- text
- buttons
- cards
- headers
- navigation items

Related elements should align intentionally.

---

# 9. Interaction States

Where applicable, test:

- hover
- focus
- active
- disabled
- loading
- error
- open/closed
- selected/unselected

Check that states remain visually consistent.

---

# 10. Responsive QA

Test representative viewports.

At minimum use:

- desktop
- tablet
- mobile

Recommended baseline viewports:

1440 × 900
1024 × 768
768 × 1024
390 × 844

Check:

- navigation
- content width
- cards
- grids
- typography
- buttons
- forms
- dialogs
- horizontal overflow
- vertical clipping

Do not assume responsive correctness from CSS inspection.

Open the actual browser viewport.

---

# 11. Overflow and Collision Detection

Explicitly look for:

- overlapping cards
- overlapping text
- controls covering content
- content escaping cards
- content escaping viewport
- horizontal scrollbars
- clipped buttons
- clipped headings
- clipped icons
- broken flex/grid children

These are high-priority defects.

---

# 12. Accessibility Basics

Check:

- keyboard focus
- visible focus state
- accessible labels
- semantic controls
- sufficient text/background distinction
- interactive elements reachable by keyboard

Do not treat accessibility as separate from visual quality.

---

# Issue Classification

Classify findings:

## CRITICAL

The UI is unusable or structurally broken.

Examples:

- overlapping primary content
- broken page layout
- inaccessible controls
- application crash
- severe mobile breakage

## MAJOR

The screen works but differs materially from the intended design.

Examples:

- incorrect grid
- major spacing problems
- wrong typography hierarchy
- major color mismatch
- major responsive issue

## MINOR

Small deviations that do not materially affect usability.

Examples:

- subtle shadow difference
- small radius difference
- tiny spacing mismatch
- minor icon alignment issue

---

# QA Output

Produce a concise report:

## Visual QA

### Critical
- issue
- location
- evidence
- recommended fix

### Major
- issue
- location
- evidence
- recommended fix

### Minor
- issue
- location
- evidence
- recommended fix

### Passed
- macro layout
- alignment
- typography
- color
- responsive behavior
- interaction
- accessibility

---

# Evidence

Whenever possible, support findings with:

- browser screenshot
- viewport
- route
- affected component
- observable behavior

Do not make claims based solely on assumptions.

---

# QA Discipline

Do not:

- rewrite components during QA
- redesign unrelated areas
- add dependencies
- modify backend code
- invent product behavior
- compare against unrelated screens
- declare success without browser inspection

QA identifies problems.

Implementation fixes problems.

Keep those responsibilities separate.