import { forwardRef, type ButtonHTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";

export type ButtonVariant = "primary" | "secondary" | "ghost" | "danger";
export type ButtonSize = "sm" | "md" | "lg";

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  /** Stretch to the width of the container. */
  block?: boolean;
  /** Show a spinner and mark the control busy. Also disables it. */
  loading?: boolean;
  /** Leading icon node (not a substitute for a text label). */
  icon?: ReactNode;
  children: ReactNode;
}

/**
 * Primary action control. Presentation only — pass an `onClick`.
 * `type` defaults to "button" so it never submits a form by accident.
 */
export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { variant = "primary", size = "md", block, loading, icon, disabled, className, children, type, ...rest },
  ref,
) {
  return (
    <button
      ref={ref}
      type={type ?? "button"}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      className={cx(
        "ui-btn",
        `ui-btn--${variant}`,
        size !== "md" && `ui-btn--${size}`,
        block && "ui-btn--block",
        className,
      )}
      {...rest}
    >
      {loading ? <span className="ui-btn__spinner" aria-hidden="true" /> : icon}
      {children}
    </button>
  );
});
