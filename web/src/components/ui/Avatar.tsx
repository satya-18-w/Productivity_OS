import { cx } from "../cx";

export interface AvatarProps {
  /** Full name — initials are derived, and it becomes the accessible label. */
  name: string;
  src?: string;
  size?: "sm" | "md" | "lg";
  /** Hide from assistive tech — use when an adjacent visible label already names it. */
  decorative?: boolean;
  className?: string;
}

function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  if (parts.length === 1) return parts[0].slice(0, 2);
  return parts[0][0] + parts[parts.length - 1][0];
}

export function Avatar({ name, src, size = "md", decorative, className }: AvatarProps) {
  return (
    <span
      className={cx("ui-avatar", size !== "md" && `ui-avatar--${size}`, className)}
      role={decorative ? "presentation" : "img"}
      aria-label={decorative ? undefined : name}
      aria-hidden={decorative || undefined}
      title={name}
    >
      {src ? <img src={src} alt="" /> : initials(name)}
    </span>
  );
}
