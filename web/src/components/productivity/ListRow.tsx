import { type HTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";

export interface ListRowProps extends Omit<HTMLAttributes<HTMLLIElement>, "title"> {
  /** Leading slot — checkbox, toggle, icon tile, drag handle. */
  lead?: ReactNode;
  title: ReactNode;
  /** Secondary line under the title. */
  meta?: ReactNode;
  /** Trailing slot — chips, dates, kebab. */
  trail?: ReactNode;
  /** Strike-through + mute the title (completed items). */
  done?: boolean;
}

/**
 * Generic entity row (design-system.md §4.8). Presentation only — the parent
 * owns state and wires the controls passed into the slots.
 */
export function ListRow({ lead, title, meta, trail, done, className, ...rest }: ListRowProps) {
  return (
    <li className={cx("ui-row", className)} {...rest}>
      {lead != null && <span className="ui-row__lead">{lead}</span>}
      <span className="ui-row__main">
        <span className={cx("ui-row__title", done && "ui-row__title--done")}>{title}</span>
        {meta != null && <span className="ui-row__meta">{meta}</span>}
      </span>
      {trail != null && <span className="ui-row__trail">{trail}</span>}
    </li>
  );
}

export interface ListGroupHeaderProps {
  children: ReactNode;
  count?: number;
  tone?: "neutral" | "success" | "danger" | "brand";
}

/** Section header for a grouped list, with a coloured left accent bar. */
export function ListGroupHeader({ children, count, tone = "neutral" }: ListGroupHeaderProps) {
  return (
    <div className={cx("ui-row-group", tone !== "neutral" && `ui-row-group--${tone}`)}>
      <span>{children}</span>
      {count != null && <span>({count})</span>}
    </div>
  );
}
