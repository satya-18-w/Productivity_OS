import { forwardRef, type InputHTMLAttributes } from "react";
import { cx } from "../cx";

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  invalid?: boolean;
}

/** Text input. Wrap with <Field> for a label/hint/error. */
export const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  { invalid, className, ...rest },
  ref,
) {
  return (
    <input
      ref={ref}
      className={cx("ui-input", className)}
      aria-invalid={invalid || undefined}
      {...rest}
    />
  );
});
