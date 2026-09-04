import { type ReactNode } from "react";
import { cx } from "../cx";

export interface FieldProps {
  label: ReactNode;
  htmlFor: string;
  hint?: ReactNode;
  error?: ReactNode;
  required?: boolean;
  className?: string;
  children: ReactNode;
}

/**
 * Label + control + hint/error wrapper. Pass the control's id as `htmlFor` and
 * wire `aria-describedby` on the control to `${htmlFor}-hint` / `${htmlFor}-error`.
 */
export function Field({ label, htmlFor, hint, error, required, className, children }: FieldProps) {
  return (
    <div className={cx("ui-field", className)}>
      <label className="ui-field__label" htmlFor={htmlFor}>
        {label}
        {required && (
          <>
            {" "}
            <span aria-hidden="true">*</span>
            <span className="ui-visually-hidden">(required)</span>
          </>
        )}
      </label>
      {children}
      {hint && !error && (
        <span className="ui-field__hint" id={`${htmlFor}-hint`}>
          {hint}
        </span>
      )}
      {error && (
        <span className="ui-field__error" id={`${htmlFor}-error`} role="alert">
          {error}
        </span>
      )}
    </div>
  );
}
