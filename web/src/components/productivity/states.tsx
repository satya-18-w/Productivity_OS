import { type ReactNode } from "react";
import { cx } from "../cx";
import { InboxIcon, AlertIcon } from "../ui/icons";

interface BaseStateProps {
  title?: ReactNode;
  message?: ReactNode;
  /** A single call-to-action (e.g. a <Button>). */
  action?: ReactNode;
  className?: string;
}

/**
 * Empty state (design-system.md §4.15). Icon + short message + one action.
 * Keep copy factual — no motivational framing (VP3).
 */
export function EmptyState({ title, message, action, icon, className }: BaseStateProps & { icon?: ReactNode }) {
  return (
    <div className={cx("ui-state", className)} role="status">
      <span className="ui-state__icon">{icon ?? <InboxIcon width={28} height={28} />}</span>
      {title && <p className="ui-state__title">{title}</p>}
      {message && <p className="ui-state__message">{message}</p>}
      {action}
    </div>
  );
}

/** Loading placeholder. `label` is announced to assistive tech. */
export function LoadingState({ label = "Loading…", className }: { label?: string; className?: string }) {
  return (
    <div className={cx("ui-state", className)} role="status" aria-live="polite">
      <span className="ui-spinner" aria-hidden="true" />
      <span className="ui-visually-hidden">{label}</span>
    </div>
  );
}

/**
 * Error state. Shows a message and, optionally, a retry action.
 * Never surfaces raw error detail (mirrors the API's stance — overview §API).
 */
export function ErrorState({
  title = "Something went wrong",
  message = "Please try again.",
  action,
  className,
}: BaseStateProps) {
  return (
    <div className={cx("ui-state", "ui-state--error", className)} role="alert">
      <span className="ui-state__icon">
        <AlertIcon width={28} height={28} />
      </span>
      <p className="ui-state__title">{title}</p>
      {message && <p className="ui-state__message">{message}</p>}
      {action}
    </div>
  );
}
