import { useEffect, useId, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { api, ApiError, type Category, type NewTask, type Task, type TaskLinkedBlock } from "../../api";
import { Dialog } from "../../components/ui/Dialog";
import { Button } from "../../components/ui/Button";
import { Badge } from "../../components/ui/Badge";
import { Field } from "../../components/ui/Field";
import { Input } from "../../components/ui/Input";
import { Select } from "../../components/ui/Select";
import { formatShortDate, utcInZone } from "../../components/date/dateUtils";
import { Textarea } from "../../components/ui/Textarea";
import { useAuth } from "../../auth";

export type TaskDialogTarget = { mode: "new" } | { mode: "edit"; task: Task };

export interface TaskDialogProps {
  open: boolean;
  target: TaskDialogTarget;
  categories: Category[];
  onClose: () => void;
  onSaved: () => void | Promise<void>;
}

export function TaskDialog({ open, target, categories, onClose, onSaved }: TaskDialogProps) {
  const existing = target.mode === "edit" ? target.task : null;
  const formId = useId();
  const { account } = useAuth();
  const timeZone = account?.timezone ?? "UTC";

  const [title, setTitle] = useState(existing?.title ?? "");
  const [description, setDescription] = useState(existing?.description ?? "");
  const [dueDate, setDueDate] = useState(existing?.due_date ?? "");
  const [categoryId, setCategoryId] = useState(existing?.category_id ?? "");
  const [error, setError] = useState("");
  const [fieldError, setFieldError] = useState<{ field: string; message: string } | null>(null);
  const [busy, setBusy] = useState(false);

  const [blocks, setBlocks] = useState<TaskLinkedBlock[] | null>(null);
  const [blocksError, setBlocksError] = useState(false);

  useEffect(() => {
    if (!existing) return;
    let live = true;
    setBlocks(null);
    setBlocksError(false);
    api
      .blocksForTask(existing.id)
      .then((b) => {
        if (live) setBlocks(b);
      })
      .catch(() => {
        if (live) setBlocksError(true);
      });
    return () => {
      live = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- keyed on the task's id only
  }, [existing?.id]);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setFieldError(null);
    setBusy(true);
    const payload: NewTask = {
      title,
      description,
      due_date: dueDate || null,
      category_id: categoryId || null,
    };
    try {
      if (existing) await api.updateTask(existing.id, payload);
      else await api.createTask(payload);
      await onSaved();
    } catch (err) {
      if (err instanceof ApiError && err.code === "VALIDATION_ERROR" && err.fields) {
        const [field, message] = Object.entries(err.fields)[0] ?? ["", "Check the fields."];
        setFieldError({ field, message });
      } else {
        setError("Could not save the task.");
      }
      setBusy(false);
    }
  }

  async function remove() {
    if (!existing) return;
    setBusy(true);
    try {
      await api.deleteTask(existing.id);
      await onSaved();
    } catch {
      setError("Could not delete the task.");
      setBusy(false);
    }
  }

  const err = (f: string) => (fieldError?.field === f ? fieldError.message : undefined);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={existing ? "Edit task" : "Add task"}
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
        <Field label="Title" htmlFor={`${formId}-title`} required error={err("title")}>
          <Input
            id={`${formId}-title`}
            required
            maxLength={200}
            autoFocus
            value={title}
            invalid={!!err("title")}
            onChange={(e) => setTitle(e.target.value)}
          />
        </Field>
        <Field label="Description" htmlFor={`${formId}-desc`} error={err("description")}>
          <Textarea
            id={`${formId}-desc`}
            rows={3}
            maxLength={5000}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </Field>
        <Field label="Due date" htmlFor={`${formId}-due`} hint="Optional." error={err("due_date")}>
          <Input
            id={`${formId}-due`}
            type="date"
            value={dueDate}
            invalid={!!err("due_date")}
            onChange={(e) => setDueDate(e.target.value)}
          />
        </Field>
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

        {existing && (
          <div className="tl2-form__scheduled">
            <p className="tl2-form__scheduled-label">Scheduled blocks</p>
            {blocksError ? (
              <p className="error" role="alert">
                Could not load scheduled blocks.
              </p>
            ) : !blocks ? (
              <p className="muted">Loading…</p>
            ) : blocks.length === 0 ? (
              <p className="muted">No time blocks scheduled yet.</p>
            ) : (
              <ul className="tl2-form__scheduled-list">
                {blocks.map((b) => {
                  const start = utcInZone(b.starts_at, timeZone);
                  const end = utcInZone(b.ends_at, timeZone);
                  return (
                    <li key={b.id}>
                      <span>
                        {formatShortDate(start.date)}, {start.time}–{end.time}
                      </span>
                      <Badge tone={b.kind === "actual" ? "brand" : "neutral"}>
                        {b.kind === "actual" ? "Actual" : "Planned"}
                      </Badge>
                      <Link to={`/timeline?date=${start.date}&openBlock=${b.id}`}>View on Timeline →</Link>
                    </li>
                  );
                })}
              </ul>
            )}
          </div>
        )}
      </form>
    </Dialog>
  );
}
