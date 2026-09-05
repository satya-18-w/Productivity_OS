import type { Category, Task, TaskState } from "../../api";
import { ListRow } from "../../components/productivity/ListRow";
import { Checkbox } from "../../components/ui/Checkbox";
import { Chip } from "../../components/ui/Chip";
import { IconButton } from "../../components/ui/IconButton";
import { Menu, type MenuItem } from "../../components/ui/Menu";
import { MoreIcon, CalendarIcon } from "../../components/ui/icons";
import { formatShortDate } from "../../components/date/dateUtils";
import { categoryColor } from "../../components/productivity/categoryColor";
import { STATE_LABELS, STATE_ORDER, isOverdue, categoryNameFor } from "./taskGroups";

export interface TaskRowProps {
  task: Task;
  today: string;
  categories: Category[];
  onToggleDone: (task: Task, done: boolean) => void;
  onMove: (task: Task, state: TaskState) => void;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
}

export function TaskRow({ task, today, categories, onToggleDone, onMove, onEdit, onDelete }: TaskRowProps) {
  const categoryName = categoryNameFor(task.category_id, categories);
  const done = task.state === "DONE";
  const overdue = isOverdue(task, today);
  const dueToday = task.due_date === today;

  const menuItems: MenuItem[] = [
    { key: "edit", label: "Edit", onSelect: () => onEdit(task) },
    { key: "sep1", separator: true },
    ...STATE_ORDER.filter((s) => s !== task.state).map((s) => ({
      key: s,
      label: `Move to ${STATE_LABELS[s]}`,
      onSelect: () => onMove(task, s),
    })),
    { key: "sep2", separator: true },
    { key: "delete", label: "Delete", danger: true, onSelect: () => onDelete(task) },
  ];

  return (
    <ListRow
      className="task-row"
      done={done}
      lead={
        <Checkbox
          checked={done}
          aria-label={`Mark "${task.title}" ${done ? "not done" : "done"}`}
          onChange={(e) => onToggleDone(task, e.target.checked)}
        />
      }
      title={
        <button type="button" className="task-row__title" onClick={() => onEdit(task)}>
          {task.title}
        </button>
      }
      meta={
        <span className="task-row__meta-line">
          <span className="chip task-row__state">{STATE_LABELS[task.state]}</span>
          {categoryName && (
            <Chip dotColor={categoryColor(task.category_id)}>{categoryName}</Chip>
          )}
          {task.due_date && (
            <span className={`task-row__due${overdue ? " task-row__due--overdue" : ""}`}>
              <CalendarIcon width={13} height={13} />
              {dueToday ? "Today" : formatShortDate(task.due_date)}
              {overdue ? " · overdue" : ""}
            </span>
          )}
        </span>
      }
      trail={
        <Menu
          label={`Actions for ${task.title}`}
          trigger={
            <IconButton label={`Actions for ${task.title}`} size="sm">
              <MoreIcon />
            </IconButton>
          }
          items={menuItems}
        />
      }
    />
  );
}
