import { type ReactNode } from "react";
import { cx } from "../cx";
import { Chip } from "../ui/Chip";
import { categoryColor } from "./categoryColor";

export interface CategoryIndicatorProps {
  /** Category name (shown when `variant` includes a label). */
  name?: string;
  /** Stable key (category id) used to pick a hue. Omit → uncategorized. */
  colorKey?: string | null;
  /** Explicit colour override (CSS colour or token var). Presentation only. */
  color?: string;
  variant?: "dot" | "tile" | "chip";
  /** Glyph for the tile variant. */
  glyph?: ReactNode;
  className?: string;
}

/**
 * Shows which category something belongs to. Colour is decorative (D2) — the
 * name is always the source of truth, so this reads correctly without colour.
 */
export function CategoryIndicator({
  name,
  colorKey,
  color,
  variant = "dot",
  glyph,
  className,
}: CategoryIndicatorProps) {
  const hue = color ?? categoryColor(colorKey);

  if (variant === "chip") {
    return (
      <Chip dotColor={hue} className={className}>
        {name ?? "Uncategorized"}
      </Chip>
    );
  }

  if (variant === "tile") {
    return (
      <span
        className={cx("ui-cat-tile", className)}
        style={{ background: `color-mix(in oklab, ${hue} 16%, var(--surface))`, color: hue }}
        role={name ? "img" : undefined}
        aria-label={name}
      >
        {glyph}
      </span>
    );
  }

  return (
    <span
      className={cx("ui-cat-dot", className)}
      style={{ background: hue }}
      role={name ? "img" : undefined}
      aria-label={name}
    />
  );
}
