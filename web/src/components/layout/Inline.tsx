import { type ElementType, type HTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";
import { type GapStep } from "./Stack";

const gapVar = (g: GapStep) => (g === 0 ? "0" : `var(--sp-${g})`);

export interface InlineProps extends HTMLAttributes<HTMLElement> {
  as?: ElementType;
  gap?: GapStep;
  wrap?: boolean;
  align?: "start" | "center" | "end" | "baseline" | "stretch";
  justify?: "start" | "center" | "end" | "space-between";
  children: ReactNode;
}

/** Horizontal flex row with a token-based gap. */
export function Inline({
  as: Tag = "div",
  gap = 2,
  wrap,
  align = "center",
  justify,
  className,
  style,
  children,
  ...rest
}: InlineProps) {
  return (
    <Tag
      className={cx("ui-inline", wrap && "ui-inline--wrap", className)}
      style={{
        gap: gapVar(gap),
        alignItems: align === "start" || align === "end" ? `flex-${align}` : align,
        justifyContent:
          justify === "start" || justify === "end" ? `flex-${justify}` : justify,
        ...style,
      }}
      {...rest}
    >
      {children}
    </Tag>
  );
}
