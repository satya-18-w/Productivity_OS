import { type HTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  tone?: "neutral" | "brand" | "danger";
  children: ReactNode;
}

/** Small count / label pill (nav items, column heads, tab labels). */
export function Badge({ tone = "neutral", className, children, ...rest }: BadgeProps) {
  return (
    <span
      className={cx("ui-badge", tone !== "neutral" && `ui-badge--${tone}`, className)}
      {...rest}
    >
      {children}
    </span>
  );
}
