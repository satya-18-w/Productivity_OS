## Frontend Reference Implementation

When implementing frontend UI from a reference image:

1. Do not treat the image as inspiration.
2. Treat it as a visual specification.
3. Read only the relevant screen specification.
4. Read only the relevant design-system sections.
5. Reuse existing components.
6. Do not create duplicate visual systems.
7. Do not introduce arbitrary colors, spacing, typography, radii or shadows.
8. Prefer Grid/Flexbox/normal flow over absolute positioning.
9. Verify the actual rendered UI with Playwright.
10. Never claim visual completion without browser verification.

## Frontend Context Discipline

Do not load the entire frontend into context for a single screen.

For a screen task, prefer:

- current screen specification
- current reference image
- relevant design-system sections
- relevant components
- current route
- relevant styles

Do not read unrelated screens or backend modules unless required.

## Visual QA Separation

Implementation and visual QA are separate responsibilities.

Implementation may modify code.

Visual QA identifies defects.

Do not silently modify unrelated UI during visual QA.

## Design System Changes

If implementation requires a new:

- color
- spacing token
- typography token
- radius
- shadow
- breakpoint

do not silently invent one.

Document the change in the design system and obtain approval when it represents a new design decision.