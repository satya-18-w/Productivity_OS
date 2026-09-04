import { useState } from "react";
import { cx } from "../cx";
import { IconButton } from "../ui/IconButton";
import { ChevronDownIcon } from "../ui/icons";
import {
  WEEKDAYS_MON_FIRST,
  formatMonthLabel,
  monthGrid,
  shiftMonths,
  toISODate,
  todayISO,
} from "./dateUtils";

export interface MiniCalendarProps {
  /** Selected date, `YYYY-MM-DD`. */
  value: string;
  onChange: (iso: string) => void;
  className?: string;
}

/**
 * Compact month calendar (design-system.md §4.13). Monday-first / ISO weeks (D8).
 * Today is ringed in --brand; the selected day is filled.
 */
export function MiniCalendar({ value, onChange, className }: MiniCalendarProps) {
  const [viewIso, setViewIso] = useState(value);
  const today = todayISO();
  const viewMonth = new Date(viewIso + "T00:00:00").getMonth();
  const grid = monthGrid(viewIso);

  return (
    <div className={cx("ui-minical", className)} role="group" aria-label="Choose a date">
      <div className="ui-minical__head">
        <IconButton
          label="Previous month"
          size="sm"
          onClick={() => setViewIso(shiftMonths(viewIso, -1))}
        >
          <ChevronDownIcon style={{ transform: "rotate(90deg)" }} width={16} height={16} />
        </IconButton>
        <span className="ui-minical__label" aria-live="polite">
          {formatMonthLabel(viewIso)}
        </span>
        <IconButton
          label="Next month"
          size="sm"
          onClick={() => setViewIso(shiftMonths(viewIso, 1))}
        >
          <ChevronDownIcon style={{ transform: "rotate(-90deg)" }} width={16} height={16} />
        </IconButton>
      </div>

      <div className="ui-minical__weekdays" aria-hidden="true">
        {WEEKDAYS_MON_FIRST.map((d) => (
          <span key={d}>{d}</span>
        ))}
      </div>

      <div className="ui-minical__grid">
        {grid.map((d) => {
          const iso = toISODate(d);
          const outside = d.getMonth() !== viewMonth;
          const selected = iso === value;
          const isToday = iso === today;
          return (
            <button
              key={iso}
              type="button"
              className={cx(
                "ui-minical__day",
                outside && "ui-minical__day--outside",
                isToday && "ui-minical__day--today",
                selected && "ui-minical__day--selected",
              )}
              aria-pressed={selected}
              aria-label={d.toLocaleDateString(undefined, {
                weekday: "long",
                day: "numeric",
                month: "long",
                year: "numeric",
              })}
              aria-current={isToday ? "date" : undefined}
              onClick={() => {
                onChange(iso);
                if (outside) setViewIso(iso);
              }}
            >
              {d.getDate()}
            </button>
          );
        })}
      </div>
    </div>
  );
}
