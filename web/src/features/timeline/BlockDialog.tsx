import { useId, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { api, ApiError, type BlockKind, type Category, type NewBlock, type PositionedBlock, type Task } from "../../api";
import { Dialog } from "../../components/ui/Dialog";
import { Button } from "../../components/ui/Button";
import { Field } from "../../components/ui/Field";
import { Input } from "../../components/ui/Input";
import { Select } from "../../components/ui/Select";
import { Checkbox } from "../../components/ui/Checkbox";
import { SegmentedControl } from "../../components/ui/SegmentedControl";

export type BlockDialogTarget =
  | { mode: "new"; kind?: BlockKind }
  | { mode: "edit"; block: PositionedBlock };

export interface BlockDialogProps {
  open: boolean;
  target: BlockDialogTarget;
  date: string;
  categories: Category[];
  tasks: Task[];
  onClose: () => void;
  onSaved: () => void | Promise<void>;
}

const KIND_OPTIONS = [
  { value: "planned", label: "Planned" },
  { value: "actual", label: "Actual" },
] as const;

export function BlockDialog({ open, target, date, categories, tasks, onClose, onSaved }: BlockDialogProps) {
  const existing = target.mode === "edit" ? target.block : null;
  const formId = useId();

  const [kind, setKind] = useState<BlockKind>(
    existing?.kind ?? (target.mode === "new" ? (target.kind ?? "planned") : "planned"),
  );
  const [formDate, setFormDate] = useState(existing?.local_date ?? date);
  const [start, setStart] = useState(existing?.local_start ?? "09:00");
  const [end, setEnd] = useState(existing?.local_end ?? "10:00");
  const [endsNextDay, setEndsNextDay] = useState(existing?.ends_next_day ?? false);
  const [categoryId, setCategoryId] = useState(existing?.category_id ?? "");
  const [taskId, setTaskId] = useState(existing?.task_id ?? "");
  const [error, setError] = useState("");
  const [fieldError, setFieldError] = useState<{ field: string; message: string } | null>(null);
  const [busy, setBusy] = useState(false);

  const linkedTask = tasks.find((t) => t.id === taskId) ?? null;
  const inheritedCategoryName = linkedTask
    ? (categories.find((c) => c.id === linkedTask.category_id)?.name ?? "— none —")
    : null;

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setFieldError(null);
    setBusy(true);
    const payload: NewBlock = {
      kind,
      date: formDate,
      start,
      end,
      ends_next_day: endsNextDay,
      category_id: taskId ? null : categoryId || null,
      task_id: taskId || null,
    };
    try {
      if (existing) await api.updateBlock(existing.id, payload);
      else await api.createBlock(payload);
      await onSaved();
    } catch (err) {
      if (err instanceof ApiError && err.code === "VALIDATION_ERROR" && err.fields) {
        const [field, message] = Object.entries(err.fields)[0] ?? ["", "Check the fields."];
        setFieldError({ field, message });
      } else {
        setError("Could not save the block.");
      }
      setBusy(false);
    }
  }

  async function remove() {
    if (!existing) return;
    setBusy(true);
    try {
      await api.deleteBlock(existing.id);
      await onSaved();
    } catch {
      setError("Could not delete the block.");
      setBusy(false);
    }
  }

  const err = (field: string) => (fieldError?.field === field ? fieldError.message : undefined);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={existing ? "Edit block" : "Add block"}
      actions={
        <>
          {existing && (
            <Button variant="danger" onClick={remove} disabled={busy}>
              Delete
            </Button>
          )}
          <Button variant="secondary" onClick={onClose} disabled={busy}>
            Cancel
          </Button>
          <Button type="submit" form={formId} loading={busy}>
            {existing ? "Save" : "Add"}
          </Button>
        </>
      }
    >
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      <form id={formId} onSubmit={submit} className="tl2-form">
        {existing ? (
          <p className="tl2-form__kind">{existing.kind === "planned" ? "Planned block" : "Actual block"}</p>
        ) : (
          <SegmentedControl label="Block type" options={KIND_OPTIONS} value={kind} onChange={setKind} />
        )}

        <Field label="Date" htmlFor={`${formId}-date`} error={err("date")}>
          <Input
            id={`${formId}-date`}
            type="date"
            required
            value={formDate}
            invalid={!!err("date")}
            onChange={(e) => setFormDate(e.target.value)}
          />
        </Field>

        <div className="tl2-form__row">
          <Field label="Start" htmlFor={`${formId}-start`} error={err("start")}>
            <Input
              id={`${formId}-start`}
              type="time"
              required
              value={start}
              invalid={!!err("start")}
              onChange={(e) => setStart(e.target.value)}
            />
          </Field>
          <Field label="End" htmlFor={`${formId}-end`} error={err("end")}>
            <Input
              id={`${formId}-end`}
              type="time"
              required
              value={end}
              invalid={!!err("end")}
              onChange={(e) => setEnd(e.target.value)}
            />
          </Field>
        </div>

        <Checkbox
          label="Ends on the next day"
          checked={endsNextDay}
          onChange={(e) => setEndsNextDay(e.target.checked)}
        />

        <Field label="Task" htmlFor={`${formId}-task`} error={err("task_id")} hint="Optional — links this block to a task instead of a standalone category.">
          <Select id={`${formId}-task`} value={taskId} onChange={(e) => setTaskId(e.target.value)}>
            <option value="">— none (standalone) —</option>
            {tasks.map((t) => (
              <option key={t.id} value={t.id}>
                {t.title}
              </option>
            ))}
          </Select>
        </Field>

        {linkedTask ? (
          <Field label="Category" htmlFor={`${formId}-cat-inherited`} hint="Inherited from the linked task.">
            <Input id={`${formId}-cat-inherited`} value={inheritedCategoryName ?? "— none —"} readOnly disabled />
          </Field>
        ) : (
          <Field label="Category" htmlFor={`${formId}-cat`} error={err("category_id")}>
            <Select
              id={`${formId}-cat`}
              value={categoryId}
              invalid={!!err("category_id")}
              onChange={(e) => setCategoryId(e.target.value)}
            >
              <option value="">— none —</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </Select>
          </Field>
        )}

        {existing?.task_id && (
          <p className="tl2-form__linked-task">
            Linked to <Link to={`/tasks?openTask=${existing.task_id}`}>{linkedTask?.title ?? "this task"} →</Link>
          </p>
        )}
      </form>
    </Dialog>
  );
}
