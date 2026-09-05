import type { ReactNode } from "react";
import { cx } from "../cx";
import { Menu, type MenuItem } from "./Menu";
import type { ButtonSize, ButtonVariant } from "./Button";

export interface SplitButtonProps {
  /** Main action label (e.g. "Add block"). */
  children: ReactNode;
  /** Fired by the main (left) segment. */
  onPrimary: () => void;
  /** Accessible name of the menu list. */
  menuLabel: string;
  /** Menu items for the caret segment. */
  items: MenuItem[];
  variant?: ButtonVariant;
  size?: ButtonSize;
  disabled?: boolean;
  /** Accessible name of the caret toggle. Defaults to "<label>, more actions". */
  caretLabel?: string;
  className?: string;
}

/**
 * Split primary action: a main button joined to a caret that opens a menu
 * (design-system §4.5 — Timeline "Add ▾", Tasks "Add Task ▾").
 * Presentation only — the caller owns `onPrimary` and the menu `items`.
 * The caret reuses the WAI-ARIA `Menu` pattern (arrow keys, Esc, focus return).
 */
export function SplitButton({
  children,
  onPrimary,
  menuLabel,
  items,
  variant = "primary",
  size = "md",
  disabled,
  caretLabel,
  className,
}: SplitButtonProps) {
  return (
    <div className={cx("ui-split", className)}>
      <button
        type="button"
        disabled={disabled}
        onClick={onPrimary}
        className={cx(
          "ui-btn",
          `ui-btn--${variant}`,
          size !== "md" && `ui-btn--${size}`,
          "ui-split__main",
        )}
      >
        {children}
      </button>
      <Menu
        label={menuLabel}
        trigger={
          <button
            type="button"
            disabled={disabled}
            aria-label={caretLabel ?? `${typeof children === "string" ? children : "Actions"}, more actions`}
            className={cx(
              "ui-btn",
              `ui-btn--${variant}`,
              size !== "md" && `ui-btn--${size}`,
              "ui-split__caret",
            )}
          >
            <span aria-hidden="true">▾</span>
          </button>
        }
        items={items}
      />
    </div>
  );
}
