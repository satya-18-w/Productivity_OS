import { useId, useRef, useState, type KeyboardEvent, type ReactNode } from "react";
import { cx } from "../cx";

export interface TabItem {
  value: string;
  label: ReactNode;
  content: ReactNode;
}

export interface TabsProps {
  items: TabItem[];
  /** Accessible name for the tab list. */
  label: string;
  /** Controlled value. Omit for uncontrolled. */
  value?: string;
  defaultValue?: string;
  onValueChange?: (value: string) => void;
  className?: string;
}

/** WAI-ARIA tabs: roving tabindex, arrow-key nav, panel wiring. */
export function Tabs({ items, label, value, defaultValue, onValueChange, className }: TabsProps) {
  const baseId = useId();
  const [internal, setInternal] = useState(defaultValue ?? items[0]?.value);
  const current = value ?? internal;
  const refs = useRef<Array<HTMLButtonElement | null>>([]);

  function select(v: string) {
    if (value === undefined) setInternal(v);
    onValueChange?.(v);
  }

  function onKeyDown(e: KeyboardEvent<HTMLButtonElement>, index: number) {
    let next = index;
    if (e.key === "ArrowRight" || e.key === "ArrowDown") next = (index + 1) % items.length;
    else if (e.key === "ArrowLeft" || e.key === "ArrowUp") next = (index - 1 + items.length) % items.length;
    else if (e.key === "Home") next = 0;
    else if (e.key === "End") next = items.length - 1;
    else return;
    e.preventDefault();
    select(items[next].value);
    refs.current[next]?.focus();
  }

  const activeItem = items.find((i) => i.value === current) ?? items[0];

  return (
    <div className={cx("ui-tabs", className)}>
      <div className="ui-tabs__list" role="tablist" aria-label={label}>
        {items.map((item, i) => {
          const selected = item.value === current;
          return (
            <button
              key={item.value}
              ref={(el) => {
                refs.current[i] = el;
              }}
              type="button"
              role="tab"
              id={`${baseId}-tab-${item.value}`}
              aria-selected={selected}
              aria-controls={`${baseId}-panel-${item.value}`}
              tabIndex={selected ? 0 : -1}
              className="ui-tabs__tab"
              onClick={() => select(item.value)}
              onKeyDown={(e) => onKeyDown(e, i)}
            >
              {item.label}
            </button>
          );
        })}
      </div>
      {activeItem && (
        <div
          role="tabpanel"
          id={`${baseId}-panel-${activeItem.value}`}
          aria-labelledby={`${baseId}-tab-${activeItem.value}`}
          className="ui-tabs__panel"
          tabIndex={0}
        >
          {activeItem.content}
        </div>
      )}
    </div>
  );
}
