import { Input } from "../../components/ui/Input";
import { Button } from "../../components/ui/Button";
import { Field } from "../../components/ui/Field";
import { shiftDays, todayISO } from "../../components/date/dateUtils";

export interface DateRangePickerProps {
  from: string;
  to: string;
  onChange: (range: { from: string; to: string }) => void;
}

const PRESETS = [
  { label: "Last 7 days", days: 6 },
  { label: "Last 30 days", days: 29 },
  { label: "Last 90 days", days: 89 },
] as const;

/** The one required Reports control (`v1.md §13`: "a user-chosen date range"). */
export function DateRangePicker({ from, to, onChange }: DateRangePickerProps) {
  return (
    <div className="report-range">
      <Field label="From" htmlFor="report-from">
        <Input
          id="report-from"
          type="date"
          value={from}
          max={to}
          onChange={(e) => e.target.value && onChange({ from: e.target.value, to })}
        />
      </Field>
      <Field label="To" htmlFor="report-to">
        <Input
          id="report-to"
          type="date"
          value={to}
          min={from}
          onChange={(e) => e.target.value && onChange({ from, to: e.target.value })}
        />
      </Field>
      <div className="report-range__presets" role="group" aria-label="Date range presets">
        {PRESETS.map((p) => (
          <Button
            key={p.label}
            variant="ghost"
            size="sm"
            onClick={() => {
              const end = todayISO();
              onChange({ from: shiftDays(end, -p.days), to: end });
            }}
          >
            {p.label}
          </Button>
        ))}
      </div>
    </div>
  );
}
