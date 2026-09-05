import { useState, type DragEvent } from "react";
import type { Category, Task, TaskState } from "../../api";
import { Badge } from "../../components/ui/Badge";
import { STATE_LABELS } from "../tasks/taskGroups";
import { TaskCard } from "./TaskCard";

export interface BoardColumnProps {
  state: TaskState;
  tasks: Task[];
  categories: Category[];
  onDropTask: (taskId: string, to: TaskState) => void;
  onEdit: (task: Task) => void;
  onMove: (task: Task, state: TaskState) => void;
  onDelete: (task: Task) => void;
}

export function BoardColumn({ state, tasks, categories, onDropTask, onEdit, onMove, onDelete }: BoardColumnProps) {
  const [over, setOver] = useState(false);

  function onDragOver(e: DragEvent) {
    if (e.dataTransfer.types.includes("text/task-id")) {
      e.preventDefault();
      e.dataTransfer.dropEffect = "move";
      setOver(true);
    }
  }
  function onDrop(e: DragEvent) {
    e.preventDefault();
    setOver(false);
    const id = e.dataTransfer.getData("text/task-id");
    if (id) onDropTask(id, state);
  }

  return (
    <section
      className={`board2__col${over ? " board2__col--dragover" : ""}`}
      aria-label={`${STATE_LABELS[state]} — ${tasks.length} task${tasks.length === 1 ? "" : "s"}`}
      onDragOver={onDragOver}
      onDragLeave={() => setOver(false)}
      onDrop={onDrop}
    >
      <div className="board2__col-head">
        <span>{STATE_LABELS[state]}</span>
        <Badge>{tasks.length}</Badge>
      </div>
      {tasks.length === 0 ? (
        <p className="board2__empty">No tasks</p>
      ) : (
        tasks.map((task) => (
          <TaskCard
            key={task.id}
            task={task}
            categories={categories}
            onEdit={onEdit}
            onMove={onMove}
            onDelete={onDelete}
          />
        ))
      )}
    </section>
  );
}
