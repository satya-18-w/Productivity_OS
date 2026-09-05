import {
  cloneElement,
  isValidElement,
  useEffect,
  useId,
  useRef,
  useState,
  type KeyboardEvent,
  type ReactElement,
  type ReactNode,
} from "react";
import { cx } from "../cx";

export type MenuItem =
  | { key: string; label: ReactNode; onSelect: () => void; danger?: boolean; disabled?: boolean }
  | { key: string; separator: true };

export interface MenuProps {
  /** The trigger element (e.g. an <IconButton>). It receives the menu-button ARIA. */
  trigger: ReactElement;
  items: MenuItem[];
  /** Accessible name for the menu list. */
  label: string;
  align?: "start" | "end";
  className?: string;
}

function isSeparator(it: MenuItem): it is { key: string; separator: true } {
  return "separator" in it;
}

/**
 * A menu-button (WAI-ARIA): click / Enter / Space / ArrowDown opens it, arrow
 * keys move, Enter selects, Esc / click-outside closes, focus returns to the
 * trigger.
 */
export function Menu({ trigger, items, label, align = "end", className }: MenuProps) {
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(0);
  const rootRef = useRef<HTMLDivElement>(null);
  const itemRefs = useRef<Array<HTMLButtonElement | null>>([]);
  const menuId = useId();

  const selectable = items
    .map((it, i) => (isSeparator(it) || it.disabled ? -1 : i))
    .filter((i) => i >= 0);

  useEffect(() => {
    if (!open) return;
    const first = selectable[0] ?? 0;
    setActive(first);
    itemRefs.current[first]?.focus();
    function onDoc(e: MouseEvent) {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) setOpen(false);
    }
    document.addEventListener("mousedown", onDoc);
    return () => document.removeEventListener("mousedown", onDoc);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  function focusTrigger() {
    (rootRef.current?.querySelector("[aria-haspopup]") as HTMLElement | null)?.focus();
  }
  function close() {
    setOpen(false);
    focusTrigger();
  }

  function move(dir: 1 | -1) {
    const pos = selectable.indexOf(active);
    const next = selectable[(pos + dir + selectable.length) % selectable.length];
    setActive(next);
    itemRefs.current[next]?.focus();
  }

  function onMenuKey(e: KeyboardEvent) {
    if (e.key === "Escape") {
      e.preventDefault();
      close();
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      move(1);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      move(-1);
    } else if (e.key === "Home") {
      e.preventDefault();
      setActive(selectable[0]);
      itemRefs.current[selectable[0]]?.focus();
    } else if (e.key === "End") {
      e.preventDefault();
      const last = selectable[selectable.length - 1];
      setActive(last);
      itemRefs.current[last]?.focus();
    }
  }

  const triggerProps: Record<string, unknown> = {
    "aria-haspopup": "menu",
    "aria-expanded": open,
    "aria-controls": open ? menuId : undefined,
    onClick: () => setOpen((v) => !v),
    onKeyDown: (e: KeyboardEvent) => {
      if (e.key === "ArrowDown" || e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        setOpen(true);
      }
    },
  };

  return (
    <div className={cx("ui-menu", className)} ref={rootRef}>
      {isValidElement(trigger)
        ? cloneElement(trigger as ReactElement<Record<string, unknown>>, triggerProps)
        : trigger}
      {open && (
        <div
          id={menuId}
          role="menu"
          aria-label={label}
          className={cx("ui-menu__list", align === "start" && "ui-menu__list--start")}
          onKeyDown={onMenuKey}
        >
          {items.map((it, i) =>
            isSeparator(it) ? (
              <div key={it.key} className="ui-menu__sep" role="separator" />
            ) : (
              <button
                key={it.key}
                ref={(el) => {
                  itemRefs.current[i] = el;
                }}
                type="button"
                role="menuitem"
                tabIndex={i === active ? 0 : -1}
                disabled={it.disabled}
                className={cx("ui-menu__item", it.danger && "ui-menu__item--danger")}
                onClick={() => {
                  it.onSelect();
                  close();
                }}
              >
                {it.label}
              </button>
            ),
          )}
        </div>
      )}
    </div>
  );
}
