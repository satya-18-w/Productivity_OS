import { forwardRef, useId, type InputHTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";
import { CheckIcon } from "./icons";

export interface CheckboxProps
  extends Omit<InputHTMLAttributes<HTMLInputElement>, "type" | "size"> {
  /** Visible label. Omit only if an external <label> / aria-label is provided. */
  label?: ReactNode;
}

/**
 * A native checkbox with a styled box. Controlled or uncontrolled — it just
 * forwards props. State/keyboard behaviour is the browser's.
 */
export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(function Checkbox(
  { label, id, className, disabled, ...rest },
  ref,
) {
  const autoId = useId();
  const inputId = id ?? autoId;

  const control = (
    <span className="ui-checkbox">
      <input ref={ref} id={inputId} type="checkbox" disabled={disabled} {...rest} />
      <span className="ui-checkbox__box" aria-hidden="true">
        <CheckIcon />
      </span>
    </span>
  );

  if (label === undefined) {
    return <span className={cx("ui-field-check", className)}>{control}</span>;
  }

  return (
    <label className={cx("ui-field-check", className)} htmlFor={inputId}>
      {control}
      <span>{label}</span>
    </label>
  );
});
