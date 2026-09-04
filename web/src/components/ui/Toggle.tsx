import { forwardRef, useId, type InputHTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";
import { CheckIcon } from "./icons";

type BaseProps = Omit<InputHTMLAttributes<HTMLInputElement>, "type" | "size">;

export interface SwitchProps extends BaseProps {
  /** Visible label text. */
  label?: ReactNode;
}

/** Binary on/off setting (settings surfaces). Native checkbox under the hood. */
export const Switch = forwardRef<HTMLInputElement, SwitchProps>(function Switch(
  { label, id, className, ...rest },
  ref,
) {
  const autoId = useId();
  const inputId = id ?? autoId;
  const control = (
    <span className="ui-switch">
      <input ref={ref} id={inputId} type="checkbox" role="switch" {...rest} />
      <span className="ui-switch__track" aria-hidden="true" />
      <span className="ui-switch__thumb" aria-hidden="true" />
    </span>
  );
  if (label === undefined) return <span className={cx("ui-field-check", className)}>{control}</span>;
  return (
    <label className={cx("ui-field-check", className)} htmlFor={inputId}>
      {control}
      <span>{label}</span>
    </label>
  );
});

export interface ToggleCircleProps extends BaseProps {
  /** Required accessible name (e.g. "Meditation — Mon 1 Sep"). */
  label: string;
}

/**
 * Circular completion control — the V1 habit / day completion toggle
 * (design-system.md §4.10). Filled green check = done; hollow ring = not.
 */
export const ToggleCircle = forwardRef<HTMLInputElement, ToggleCircleProps>(function ToggleCircle(
  { label, className, ...rest },
  ref,
) {
  return (
    <span className={cx("ui-toggle-circle", className)}>
      <input ref={ref} type="checkbox" aria-label={label} title={label} {...rest} />
      <span className="ui-toggle-circle__ring" aria-hidden="true">
        <CheckIcon />
      </span>
    </span>
  );
});
