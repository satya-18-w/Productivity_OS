import { Button } from "../ui/Button";
import { IconButton } from "../ui/IconButton";
import { Input } from "../ui/Input";
import { ChevronDownIcon } from "../ui/icons";
import { shiftDays, todayISO } from "./dateUtils";

export interface DateStepperProps {
  /** The selected date, ISO (`YYYY-MM-DD`). */
  value: string;
  onChange: (iso: string) => void;
  /** Accessible name for the native date input. Default: "Date". */
  label?: string;
  /**
   * Override what `‹`/`›` step by (default: ±1 day) — e.g. Timeline's Week
   * view steps by 7 days, Month by a calendar month. Receives the direction.
   */
  onStep?: (direction: -1 | 1) => void;
  /** Accessible names for the step buttons when `onStep` changes their unit. */
  prevLabel?: string;
  nextLabel?: string;
}

/**
 * `‹ date › + Today` — the one date-navigation control shared by every
 * single-date screen (Timeline, Daily Review). Extracted from Timeline's
 * toolbar (frontend-implementation-plan.md's planned "DateStepper" primitive)
 * so the two screens don't grow duplicate visual systems (CLAUDE.md).
 */
export function DateStepper({
  value,
  onChange,
  label = "Date",
  onStep,
  prevLabel = "Previous day",
  nextLabel = "Next day",
}: DateStepperProps) {
  const isToday = value === todayISO();
  const step = (direction: -1 | 1) =>
    onStep ? onStep(direction) : onChange(shiftDays(value, direction));
  return (
    <div className="date-stepper">
      <IconButton label={prevLabel} size="sm" onClick={() => step(-1)}>
        <ChevronDownIcon style={{ transform: "rotate(90deg)" }} width={16} height={16} />
      </IconButton>
      <Input
        type="date"
        aria-label={label}
        value={value}
        onChange={(e) => e.target.value && onChange(e.target.value)}
        style={{ width: "auto" }}
      />
      <IconButton label={nextLabel} size="sm" onClick={() => step(1)}>
        <ChevronDownIcon style={{ transform: "rotate(-90deg)" }} width={16} height={16} />
      </IconButton>
      <Button variant="secondary" size="sm" onClick={() => onChange(todayISO())} disabled={isToday}>
        Today
      </Button>
    </div>
  );
}
