import { IconButton } from "../../components/ui/IconButton";
import { Menu } from "../../components/ui/Menu";
import { MoreIcon, FlameIcon } from "../../components/ui/icons";

/** Current streak — a plain number with a small static flame glyph (no animation, VP3). */
export function Streak({ value }: { value: number }) {
  return (
    <span className="habit-streak" title="Current streak">
      <FlameIcon className="habit-streak__flame" width={13} height={13} />
      <span className="habit-streak__n">{value}</span>
      <span className="ui-visually-hidden">day streak</span>
    </span>
  );
}

/** Last-30 count — plain tabular N/30 from `HabitView.last_30_days` (V1 scope). */
export function Last30({ value }: { value: number }) {
  return (
    <span className="habit-last30" title={`${value} of the last 30 days`}>
      <span aria-hidden="true">{value}/30</span>
      <span className="ui-visually-hidden">{value} of the last 30 days</span>
    </span>
  );
}

export function HabitActions({ name, onArchive }: { name: string; onArchive: () => void }) {
  return (
    <Menu
      label={`Actions for ${name}`}
      trigger={
        <IconButton label={`Actions for ${name}`} size="sm">
          <MoreIcon />
        </IconButton>
      }
      items={[{ key: "archive", label: "Archive", onSelect: onArchive }]}
    />
  );
}
