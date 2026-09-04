import { forwardRef, type SelectHTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";
import { ChevronDownIcon } from "./icons";

export interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  invalid?: boolean;
  children: ReactNode;
}

/** Native <select> with a styled shell and a chevron affordance. */
export const Select = forwardRef<HTMLSelectElement, SelectProps>(function Select(
  { invalid, className, children, ...rest },
  ref,
) {
  return (
    <span className="ui-select-wrap">
      <select
        ref={ref}
        className={cx("ui-select", className)}
        aria-invalid={invalid || undefined}
        {...rest}
      >
        {children}
      </select>
      <ChevronDownIcon className="ui-select-wrap__chevron" width={16} height={16} />
    </span>
  );
});
