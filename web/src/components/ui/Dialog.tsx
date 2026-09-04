import { useEffect, useId, useRef, type ReactNode } from "react";
import { cx } from "../cx";
import { IconButton } from "./IconButton";
import { CloseIcon } from "./icons";

export interface DialogProps {
  open: boolean;
  onClose: () => void;
  title: ReactNode;
  /** Footer actions (buttons). */
  actions?: ReactNode;
  /** Close when the backdrop is clicked (default true). */
  dismissable?: boolean;
  className?: string;
  children: ReactNode;
}

/**
 * Modal dialog built on the native <dialog> element — this gives us the focus
 * trap, initial focus, `Esc` to close and inert background for free.
 */
export function Dialog({
  open,
  onClose,
  title,
  actions,
  dismissable = true,
  className,
  children,
}: DialogProps) {
  const ref = useRef<HTMLDialogElement>(null);
  const titleId = useId();

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    if (open && !el.open) el.showModal();
    else if (!open && el.open) el.close();
  }, [open]);

  return (
    <dialog
      ref={ref}
      className={cx("ui-dialog", className)}
      aria-labelledby={titleId}
      onCancel={(e) => {
        e.preventDefault();
        onClose();
      }}
      onClose={onClose}
      onClick={(e) => {
        if (!dismissable) return;
        // Click landed on the <dialog> itself (the backdrop area), not content.
        if (e.target === ref.current) onClose();
      }}
    >
      <div className="ui-dialog__body">
        <div className="ui-dialog__header">
          <h2 id={titleId} className="ui-dialog__title">
            {title}
          </h2>
          <IconButton label="Close" onClick={onClose}>
            <CloseIcon width={18} height={18} />
          </IconButton>
        </div>
        <div>{children}</div>
        {actions && <div className="ui-dialog__actions">{actions}</div>}
      </div>
    </dialog>
  );
}
