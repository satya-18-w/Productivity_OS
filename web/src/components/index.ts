/**
 * Productivity OS design-system foundation.
 *
 * One visual system: every component here is presentation-only and draws its
 * values from src/styles/tokens.css (governed by docs/design/design-system.md).
 * Feature/business behaviour is added by the screens that use these — not here.
 */
export * from "./ui";
export * from "./layout";
export * from "./productivity";
export { cx } from "./cx";
