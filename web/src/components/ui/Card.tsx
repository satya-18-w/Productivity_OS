import { type HTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";

export interface CardProps extends Omit<HTMLAttributes<HTMLElement>, "title"> {
  /** Render as <section> with an accessible heading link when `title` is set. */
  as?: "div" | "section" | "article";
  padding?: "default" | "compact" | "flush";
  raised?: boolean;
  /** Optional header: a title and an actions slot on the right. */
  title?: ReactNode;
  headingLevel?: 2 | 3 | 4;
  actions?: ReactNode;
  children: ReactNode;
}

export function Card({
  as: Tag = "div",
  padding = "default",
  raised,
  title,
  headingLevel = 3,
  actions,
  className,
  children,
  ...rest
}: CardProps) {
  const Heading = `h${headingLevel}` as "h2" | "h3" | "h4";
  return (
    <Tag
      className={cx(
        "ui-card",
        padding === "compact" && "ui-card--compact",
        padding === "flush" && "ui-card--flush",
        raised && "ui-card--raised",
        className,
      )}
      {...rest}
    >
      {(title || actions) && (
        <div className="ui-card__header">
          {title ? <Heading className="ui-card__title">{title}</Heading> : <span />}
          {actions}
        </div>
      )}
      {children}
    </Tag>
  );
}
