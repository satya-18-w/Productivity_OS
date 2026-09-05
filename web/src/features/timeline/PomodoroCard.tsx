import { useEffect, useRef, useState } from "react";
import { Card } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { SegmentedControl } from "../../components/ui/SegmentedControl";

const PRESETS = [
  { value: "focus", label: "Focus", minutes: 25 },
  { value: "short", label: "Short break", minutes: 5 },
  { value: "long", label: "Long break", minutes: 15 },
] as const;
type PresetId = (typeof PRESETS)[number]["value"];

const OPTIONS = PRESETS.map((p) => ({ value: p.value, label: p.label }));

function minutesFor(preset: PresetId): number {
  return PRESETS.find((p) => p.value === preset)!.minutes;
}

function fmt(totalSeconds: number): string {
  const m = Math.floor(totalSeconds / 60);
  const s = totalSeconds % 60;
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

const RADIUS = 42;
const CIRCUMFERENCE = 2 * Math.PI * RADIUS;

/**
 * Standalone focus-session timer (`v1.md §4`, "Focus timer" amendment,
 * 2026-09-05; `design-system.md` G2) — Timeline Day's rail, per
 * `references/timeline.png`. Deliberately disconnected from block data: it
 * never creates, edits, or reads a time block, and keeps no state once the
 * component unmounts. A session finishing does not log anything.
 */
export function PomodoroCard() {
  const [preset, setPreset] = useState<PresetId>("focus");
  const [secondsLeft, setSecondsLeft] = useState(() => minutesFor("focus") * 60);
  const [running, setRunning] = useState(false);
  const total = minutesFor(preset) * 60;

  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    if (!running) return;
    intervalRef.current = setInterval(() => {
      setSecondsLeft((s) => {
        if (s <= 1) {
          setRunning(false);
          return 0;
        }
        return s - 1;
      });
    }, 1000);
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [running]);

  function choosePreset(next: string) {
    const id = next as PresetId;
    setPreset(id);
    setRunning(false);
    setSecondsLeft(minutesFor(id) * 60);
  }

  function reset() {
    setRunning(false);
    setSecondsLeft(total);
  }

  const fraction = total === 0 ? 0 : secondsLeft / total;
  const done = secondsLeft === 0;

  return (
    <Card padding="compact">
      <h2 className="pomodoro__title">Focus timer</h2>
      <SegmentedControl label="Session type" options={OPTIONS} value={preset} onChange={choosePreset} />
      <div className="pomodoro__ring-wrap">
        <svg viewBox="0 0 100 100" className="pomodoro__ring" role="img" aria-label={`${fmt(secondsLeft)} remaining`}>
          <circle className="pomodoro__ring-track" cx="50" cy="50" r={RADIUS} />
          <circle
            className="pomodoro__ring-fill"
            cx="50"
            cy="50"
            r={RADIUS}
            strokeDasharray={CIRCUMFERENCE}
            strokeDashoffset={CIRCUMFERENCE * (1 - fraction)}
          />
        </svg>
        <span className="pomodoro__time" aria-hidden="true">
          {fmt(secondsLeft)}
        </span>
      </div>
      <div className="pomodoro__actions">
        <Button variant="secondary" size="sm" onClick={reset} disabled={secondsLeft === total && !running}>
          Reset
        </Button>
        <Button size="sm" onClick={() => setRunning((r) => !r)} disabled={done}>
          {running ? "Pause" : done ? "Done" : secondsLeft === total ? "Start" : "Resume"}
        </Button>
      </div>
      <p className="hint pomodoro__hint">A standalone timer — it doesn't log a time block.</p>
    </Card>
  );
}
