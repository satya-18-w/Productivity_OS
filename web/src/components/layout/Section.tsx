import { type HTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";

export interface SectionProps extends HTMLAttributes<HTMLElement> {
  children: ReactNode;
}

/**
 * A vertical content group with consistent internal spacing and top margin
 * when stacked with a sibling <Section>. Renders a semantic <section>.
 */
export function Section({ className, children, ...rest }: SectionProps) {
  return (
    <section className={cx("ui-section", className)} {...rest}>
      {children}
    </section>
  );
}
