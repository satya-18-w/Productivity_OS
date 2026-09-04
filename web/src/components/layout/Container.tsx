import { type HTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";

export interface ContainerProps extends HTMLAttributes<HTMLDivElement> {
  width?: "default" | "wide" | "narrow";
  children: ReactNode;
}

/** Centered, max-width, gutter-padded content column. */
export function Container({ width = "default", className, children, ...rest }: ContainerProps) {
  return (
    <div
      className={cx("ui-container", width !== "default" && `ui-container--${width}`, className)}
      {...rest}
    >
      {children}
    </div>
  );
}
