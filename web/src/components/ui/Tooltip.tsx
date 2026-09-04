import { useId, useState, type ReactElement, type ReactNode } from "react";

export interface TooltipProps {
  /** Tooltip text. */
  content: ReactNode;
  /** Single focusable/hoverable trigger element. */
  children: ReactElement;
}

/**
 * Lightweight tooltip. Shows on hover and keyboard focus; hides on blur / Esc.
 * The trigger is described by the tip via `aria-describedby`.
 */
export function Tooltip({ content, children }: TooltipProps) {
  const [open, setOpen] = useState(false);
  const id = useId();

  return (
    <span
      className="ui-tooltip-wrap"
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
      onFocusCapture={() => setOpen(true)}
      onBlurCapture={() => setOpen(false)}
      onKeyDown={(e) => {
        if (e.key === "Escape") setOpen(false);
      }}
    >
      <span aria-describedby={open ? id : undefined}>{children}</span>
      {open && (
        <span role="tooltip" id={id} className="ui-tooltip">
          {content}
        </span>
      )}
    </span>
  );
}
