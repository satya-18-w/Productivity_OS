import { useState, type KeyboardEvent } from "react";
import { getThemePreference, setThemePreference, type ThemePreference } from "../theme";
import { SunIcon, MoonIcon, MonitorIcon } from "../components/ui/icons";
import { IconButton } from "../components/ui/IconButton";

const OPTIONS: { value: ThemePreference; label: string; Icon: typeof SunIcon }[] = [
  { value: "light", label: "Light", Icon: SunIcon },
  { value: "dark", label: "Dark", Icon: MoonIcon },
  { value: "system", label: "System", Icon: MonitorIcon },
];

export interface ThemeToggleProps {
  /** Collapsed sidebar → a single cycling button instead of the 3-up radiogroup. */
  compact?: boolean;
}

export function ThemeToggle({ compact }: ThemeToggleProps) {
  const [pref, setPref] = useState<ThemePreference>(() => getThemePreference());

  function choose(next: ThemePreference) {
    setPref(next);
    setThemePreference(next);
  }

  if (compact) {
    const current = OPTIONS.find((o) => o.value === pref) ?? OPTIONS[2];
    const nextIndex = (OPTIONS.findIndex((o) => o.value === pref) + 1) % OPTIONS.length;
    return (
      <IconButton
        label={`Theme: ${current.label}. Switch to ${OPTIONS[nextIndex].label}`}
        onClick={() => choose(OPTIONS[nextIndex].value)}
      >
        <current.Icon />
      </IconButton>
    );
  }

  function onKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    const i = OPTIONS.findIndex((o) => o.value === pref);
    if (e.key === "ArrowRight" || e.key === "ArrowDown") {
      e.preventDefault();
      choose(OPTIONS[(i + 1) % OPTIONS.length].value);
    } else if (e.key === "ArrowLeft" || e.key === "ArrowUp") {
      e.preventDefault();
      choose(OPTIONS[(i - 1 + OPTIONS.length) % OPTIONS.length].value);
    }
  }

  return (
    <div className="theme-toggle">
      <span className="theme-toggle__label" id="theme-toggle-label">
        Theme
      </span>
      <div
        className="theme-toggle__options"
        role="radiogroup"
        aria-labelledby="theme-toggle-label"
        onKeyDown={onKeyDown}
      >
        {OPTIONS.map((o) => {
          const selected = o.value === pref;
          return (
            <button
              key={o.value}
              type="button"
              role="radio"
              aria-checked={selected}
              aria-label={o.label}
              tabIndex={selected ? 0 : -1}
              className="theme-toggle__opt"
              onClick={() => choose(o.value)}
            >
              <o.Icon />
            </button>
          );
        })}
      </div>
    </div>
  );
}
