import { forwardRef, type ButtonHTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";

export interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  /** Required — the icon carries no text, so it needs an accessible name. */
  label: string;
  size?: "sm" | "md" | "lg";
  children: ReactNode;
}

export const IconButton = forwardRef<HTMLButtonElement, IconButtonProps>(function IconButton(
  { label, size = "md", className, children, type, ...rest },
  ref,
) {
  return (
    <button
      ref={ref}
      type={type ?? "button"}
      aria-label={label}
      title={label}
      className={cx("ui-icon-btn", size !== "md" && `ui-icon-btn--${size}`, className)}
      {...rest}
    >
      {children}
    </button>
  );
});
