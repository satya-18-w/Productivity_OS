import { useState, type DragEvent } from "react";
import type { Category, Task, TaskState } from "../../api";
import { Chip } from "../../components/ui/Chip";
import { IconButton } from "../../components/ui/IconButton";
import { Menu, type MenuItem } from "../../components/ui/Menu";
import { MoreIcon, CalendarIcon } from "../../components/ui/icons";
import { formatShortDate, todayISO } from "../../components/date/dateUtils";
import { categoryColor } from "../../components/productivity/categoryColor";
import { STATE_LABELS, STATE_ORDER, isOverdue, categoryNameFor } from "../tasks/taskGroups";

export interface TaskCardProps {
  task: Task;
  categories: Category[];
  onEdit: (task: Task) => void;
  onMove: (task: Task, state: TaskState) => void;
  onDelete: (task: Task) => void;
}

export function TaskCard({ task, categories, onEdit, onMove, onDelete }: TaskCardProps) {
  const [dragging, setDragging] = useState(false);
  const today = todayISO();
  const overdue = isOverdue(task, today);
  const categoryName = categoryNameFor(task.category_id, categories);

  const menuItems: MenuItem[] = [
    { key: "edit", label: "Edit", onSelect: () => onEdit(task) },
    { key: "sep", separator: true },
    ...STATE_ORDER.filter((s) => s !== task.state).map((s) => ({
      key: s,
      label: `Move to ${STATE_LABELS[s]}`,
      onSelect: () => onMove(task, s),
    })),
    { key: "sep2", separator: true },
    { key: "del", label: "Delete", danger: true, onSelect: () => onDelete(task) },
  ];

  function onDragStart(e: DragEvent) {
    e.dataTransfer.setData("text/task-id", task.id);
    e.dataTransfer.effectAllowed = "move";
    setDragging(true);
  }

  return (
    <article
      className={`board2__card${dragging ? " board2__card--dragging" : ""}`}
      draggable
      onDragStart={onDragStart}
      onDragEnd={() => setDragging(false)}
      aria-label={`${task.title} — ${STATE_LABELS[task.state]}`}
    >
      <div className="board2__card-head">
        <button type="button" className="board2__card-title" onClick={() => onEdit(task)}>
          {task.title}
        </button>
        <Menu
          label={`Actions for ${task.title}`}
          trigger={
            <IconButton label={`Actions for ${task.title}`} size="sm">
              <MoreIcon />
            </IconButton>
          }
          items={menuItems}
        />
      </div>
      {task.description && <p className="board2__card-desc">{task.description}</p>}
      {(categoryName || task.due_date) && (
        <div className="board2__card-foot">
          {categoryName && <Chip dotColor={categoryColor(task.category_id)}>{categoryName}</Chip>}
          {task.due_date && (
            <span className={`board2__card-due${overdue ? " board2__card-due--overdue" : ""}`}>
              <CalendarIcon width={13} height={13} />
              {task.due_date === today ? "Today" : formatShortDate(task.due_date)}
              {overdue ? " · overdue" : ""}
            </span>
          )}
        </div>
      )}
    </article>
  );
}
