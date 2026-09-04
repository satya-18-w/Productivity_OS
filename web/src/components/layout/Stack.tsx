import { type ElementType, type HTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";

/** Spacing-scale step (maps to --sp-N). 0 = no gap. */
export type GapStep = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7;

const gapVar = (g: GapStep) => (g === 0 ? "0" : `var(--sp-${g})`);

export interface StackProps extends HTMLAttributes<HTMLElement> {
  as?: ElementType;
  gap?: GapStep;
  align?: "start" | "center" | "end" | "stretch";
  children: ReactNode;
}

/** Vertical flex column with a token-based gap. */
export function Stack({ as: Tag = "div", gap = 4, align, className, style, children, ...rest }: StackProps) {
  return (
    <Tag
      className={cx("ui-stack", className)}
      style={{ gap: gapVar(gap), alignItems: align, ...style }}
      {...rest}
    >
      {children}
    </Tag>
  );
}
