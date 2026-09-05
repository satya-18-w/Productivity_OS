import { useId, useState, type FormEvent } from "react";
import { api, ApiError, type Goal, type NewGoal } from "../../api";
import { Dialog } from "../../components/ui/Dialog";
import { Button } from "../../components/ui/Button";
import { Field } from "../../components/ui/Field";
import { Input } from "../../components/ui/Input";
import { Textarea } from "../../components/ui/Textarea";

export type GoalDialogTarget = { mode: "new" } | { mode: "edit"; goal: Goal };

export interface GoalDialogProps {
  open: boolean;
  target: GoalDialogTarget;
  onClose: () => void;
  onSaved: () => void | Promise<void>;
}

/** Create / edit a goal — title, description, target date. No %, no linked tasks (§10). */
export function GoalDialog({ open, target, onClose, onSaved }: GoalDialogProps) {
  const existing = target.mode === "edit" ? target.goal : null;
  const formId = useId();

  const [title, setTitle] = useState(existing?.title ?? "");
  const [description, setDescription] = useState(existing?.description ?? "");
  const [targetDate, setTargetDate] = useState(existing?.target_date ?? "");
  const [error, setError] = useState("");
  const [fieldError, setFieldError] = useState<{ field: string; message: string } | null>(null);
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setFieldError(null);
    setBusy(true);
    const payload: NewGoal = { title, description, target_date: targetDate || null };
    try {
      if (existing) await api.updateGoal(existing.id, payload);
      else await api.createGoal(payload);
      await onSaved();
    } catch (err) {
      if (err instanceof ApiError && err.code === "VALIDATION_ERROR" && err.fields) {
        const [field, message] = Object.entries(err.fields)[0] ?? ["", "Check the fields."];
        setFieldError({ field, message });
      } else {
        setError("Could not save the goal.");
      }
      setBusy(false);
    }
  }

  const err = (f: string) => (fieldError?.field === f ? fieldError.message : undefined);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={existing ? "Edit goal" : "New goal"}
      actions={
        <>
          <Button variant="secondary" onClick={onClose} disabled={busy}>
            Cancel
          </Button>
          <Button type="submit" form={formId} loading={busy}>
            {existing ? "Save" : "Create"}
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
        <Field label="Target date" htmlFor={`${formId}-target`} hint="Optional." error={err("target_date")}>
          <Input
            id={`${formId}-target`}
            type="date"
            value={targetDate}
            invalid={!!err("target_date")}
            onChange={(e) => setTargetDate(e.target.value)}
          />
        </Field>
      </form>
    </Dialog>
  );
}
