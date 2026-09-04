import { useRef, type KeyboardEvent } from "react";
import { cx } from "../cx";

export interface SegmentedOption<T extends string> {
  value: T;
  label: string;
}

export interface SegmentedControlProps<T extends string> {
  options: ReadonlyArray<SegmentedOption<T>>;
  value: T;
  onChange: (value: T) => void;
  /** Accessible name for the group. */
  label: string;
  className?: string;
}

/**
 * Pill view-switcher (design-system.md §4.3). Single selection, keyboard
 * navigable (Arrow keys / Home / End), exposed as an ARIA radiogroup.
 */
export function SegmentedControl<T extends string>({
  options,
  value,
  onChange,
  label,
  className,
}: SegmentedControlProps<T>) {
  const refs = useRef<Array<HTMLButtonElement | null>>([]);

  function move(delta: number, from: number) {
    const next = (from + delta + options.length) % options.length;
    onChange(options[next].value);
    refs.current[next]?.focus();
  }

  function onKeyDown(e: KeyboardEvent<HTMLButtonElement>, index: number) {
    switch (e.key) {
      case "ArrowRight":
      case "ArrowDown":
        e.preventDefault();
        move(1, index);
        break;
      case "ArrowLeft":
      case "ArrowUp":
        e.preventDefault();
        move(-1, index);
        break;
      case "Home":
        e.preventDefault();
        onChange(options[0].value);
        refs.current[0]?.focus();
        break;
      case "End":
        e.preventDefault();
        onChange(options[options.length - 1].value);
        refs.current[options.length - 1]?.focus();
        break;
    }
  }

  return (
    <div className={cx("ui-segmented", className)} role="radiogroup" aria-label={label}>
      {options.map((opt, i) => {
        const selected = opt.value === value;
        return (
          <button
            key={opt.value}
            ref={(el) => {
              refs.current[i] = el;
            }}
            type="button"
            role="radio"
            aria-checked={selected}
            aria-selected={selected}
            tabIndex={selected ? 0 : -1}
            className="ui-segmented__item"
            onClick={() => onChange(opt.value)}
            onKeyDown={(e) => onKeyDown(e, i)}
          >
            {opt.label}
          </button>
        );
      })}
    </div>
  );
}
