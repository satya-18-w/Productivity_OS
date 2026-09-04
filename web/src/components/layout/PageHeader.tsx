import { type ReactNode } from "react";
import { cx } from "../cx";

export interface PageHeaderProps {
  /** Uppercase kicker above the title (optional). */
  eyebrow?: ReactNode;
  title: ReactNode;
  /** One line: a plain description. Keep it factual — no motivational copy (VP3). */
  subtitle?: ReactNode;
  /** Right-aligned actions (typically one primary button). */
  actions?: ReactNode;
  className?: string;
}

/**
 * Standard screen header (design-system.md §4.2). Deliberately omits the
 * decorative header illustration — that is an opt-in decorative surface (D6),
 * not part of the structural header.
 */
export function PageHeader({ eyebrow, title, subtitle, actions, className }: PageHeaderProps) {
  return (
    <header className={cx("ui-page-header", className)}>
      <div className="ui-page-header__row">
        <div>
          {eyebrow && <div className="ui-page-header__eyebrow">{eyebrow}</div>}
          <h1 className="ui-page-header__title">{title}</h1>
        </div>
        {actions && <div className="ui-page-header__actions">{actions}</div>}
      </div>
      {subtitle && <p className="ui-page-header__subtitle">{subtitle}</p>}
    </header>
  );
}
