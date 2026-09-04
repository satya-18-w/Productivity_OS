import { type ReactNode } from "react";
import { cx } from "../cx";

export interface ChipProps {
  children: ReactNode;
  /** A colour for the leading dot. Presentation only — never logic (D2). */
  dotColor?: string;
  className?: string;
  /** When set, the chip renders as a toggle button (filter chips). */
  onToggle?: () => void;
  active?: boolean;
  disabled?: boolean;
  title?: string;
}

/**
 * Generic pill. For a category label use <CategoryChip> which wires the
 * category-colour convention; this is the unstyled-by-meaning base.
 */
export function Chip({ children, dotColor, className, onToggle, active, disabled, title }: ChipProps) {
  const content = (
    <>
      {dotColor && <span className="ui-chip__dot" style={{ background: dotColor }} aria-hidden="true" />}
      <span className="ui-chip__label">{children}</span>
    </>
  );

  if (onToggle) {
    return (
      <button
        type="button"
        className={cx("ui-chip", active && "ui-chip--active", className)}
        aria-pressed={active}
        disabled={disabled}
        onClick={onToggle}
        title={title}
      >
        {content}
      </button>
    );
  }

  return (
    <span className={cx("ui-chip", className)} title={title}>
      {content}
    </span>
  );
}
