import { forwardRef } from "react";
import { Select, type SelectProps } from "../../components/ui/Select";
import { browserTimezone } from "../../api";

/**
 * IANA timezone picker (Q4: browser-detected default, fallback `UTC`).
 *
 * FEATURE-LOCAL shared primitive: Account (`pages/Account.tsx`) and Register
 * (`pages/Register.tsx`) both need it, but scope for this pass forbids editing
 * `components/ui` / `styles/*` (except `styles/reviews.css`) — so it lives
 * here until it can be promoted to `components/ui/TimezoneSelect.tsx`.
 * Presentation only — a native `Select` over `Intl.supportedValuesOf`.
 */
export function timezoneOptions(current?: string): string[] {
  let names: string[] = [];
  try {
    const supported = (Intl as { supportedValuesOf?: (key: string) => string[] }).supportedValuesOf;
    if (typeof supported === "function") names = supported.call(Intl, "timeZone");
  } catch {
    names = [];
  }
  if (!Array.isArray(names) || names.length === 0) {
    const fallback = browserTimezone();
    return fallback === "UTC" ? ["UTC"] : [fallback, "UTC"];
  }
  // `supportedValuesOf` is not guaranteed to include "UTC" (ICU-dependent) —
  // always offer it, plus the current value when unknown.
  const head: string[] = [];
  if (current && !names.includes(current)) head.push(current);
  if (!names.includes("UTC") && !head.includes("UTC")) head.push("UTC");
  return [...head, ...names];
}

export interface TimezoneSelectProps extends Omit<SelectProps, "children" | "value" | "onChange"> {
  value: string;
  onChange: (timezone: string) => void;
}

/** Native-select timezone picker. Wrap with `<Field>` for label/hint/error. */
export const TimezoneSelect = forwardRef<HTMLSelectElement, TimezoneSelectProps>(
  function TimezoneSelect({ value, onChange, ...rest }, ref) {
    return (
      <Select
        ref={ref}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        {...rest}
      >
        {timezoneOptions(value).map((tz) => (
          <option key={tz} value={tz}>
            {tz}
          </option>
        ))}
      </Select>
    );
  },
);
