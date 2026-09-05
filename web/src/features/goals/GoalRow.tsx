import type { Goal } from "../../api";
import type { GoalProgress } from "../../components/productivity/StatusBadge";
import { StatusBadge, GOAL_PROGRESS_LABELS } from "../../components/productivity/StatusBadge";
import { IconButton } from "../../components/ui/IconButton";
import { Menu, type MenuItem } from "../../components/ui/Menu";
import { MoreIcon, CalendarIcon } from "../../components/ui/icons";
import { formatShortDate, todayISO } from "../../components/date/dateUtils";
import { PROGRESS_ORDER } from "./goalHelpers";

export interface GoalRowProps {
  goal: Goal;
  onSetProgress: (goal: Goal, progress: GoalProgress) => void;
  onEdit: (goal: Goal) => void;
  onDelete: (goal: Goal) => void;
}

export function GoalRow({ goal, onSetProgress, onEdit, onDelete }: GoalRowProps) {
  const overdue =
    goal.target_date != null && goal.target_date < todayISO() && goal.progress !== "ACHIEVED" && goal.progress !== "ABANDONED";

  const items: MenuItem[] = [
    { key: "edit", label: "Edit", onSelect: () => onEdit(goal) },
    { key: "sep1", separator: true },
    ...PROGRESS_ORDER.filter((p) => p !== goal.progress).map((p) => ({
      key: p,
      label: `Set to ${GOAL_PROGRESS_LABELS[p]}`,
      onSelect: () => onSetProgress(goal, p),
    })),
    { key: "sep2", separator: true },
    { key: "delete", label: "Delete", danger: true, onSelect: () => onDelete(goal) },
  ];

  return (
    <article className="goal-row">
      <div className="goal-row__main">
        <button type="button" className="goal-row__title" onClick={() => onEdit(goal)}>
          {goal.title}
        </button>
        {goal.description && <p className="goal-row__desc">{goal.description}</p>}
        <div className="goal-row__meta">
          <StatusBadge status={goal.progress} />
          {goal.target_date && (
            <span className={`goal-row__target${overdue ? " goal-row__target--overdue" : ""}`}>
              <CalendarIcon width={13} height={13} />
              {formatShortDate(goal.target_date)}
              {overdue ? " · past due" : ""}
            </span>
          )}
        </div>
      </div>
      <Menu
        label={`Actions for ${goal.title}`}
        trigger={
          <IconButton label={`Actions for ${goal.title}`} size="sm">
            <MoreIcon />
          </IconButton>
        }
        items={items}
      />
    </article>
  );
}
