import { useId, useState, type FormEvent } from "react";
import { api, ApiError } from "../../api";
import { Dialog } from "../../components/ui/Dialog";
import { Button } from "../../components/ui/Button";
import { Field } from "../../components/ui/Field";
import { Input } from "../../components/ui/Input";

export interface HabitDialogProps {
  open: boolean;
  onClose: () => void;
  onSaved: () => void | Promise<void>;
}

/**
 * Create a habit — **name only** (`v1.md §9`: a habit has just a name).
 * Rename / delete are not V1 and have no endpoint yet — see docs/left.md.
 */
export function HabitDialog({ open, onClose, onSaved }: HabitDialogProps) {
  const formId = useId();
  const [name, setName] = useState("");
  const [error, setError] = useState("");
  const [fieldError, setFieldError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setFieldError("");
    setBusy(true);
    try {
      await api.createHabit(name);
      await onSaved();
    } catch (err) {
      if (err instanceof ApiError && err.code === "VALIDATION_ERROR" && err.fields) {
        setFieldError(Object.values(err.fields)[0] ?? "Check the name.");
      } else {
        setError("Could not create the habit.");
      }
      setBusy(false);
    }
  }

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Add habit"
      actions={
        <>
          <Button variant="secondary" onClick={onClose} disabled={busy}>
            Cancel
          </Button>
          <Button type="submit" form={formId} loading={busy}>
            Add
          </Button>
        </>
      }
    >
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      <form id={formId} onSubmit={submit}>
        <Field label="Name" htmlFor={`${formId}-name`} required error={fieldError || undefined}>
          <Input
            id={`${formId}-name`}
            required
            maxLength={100}
            autoFocus
            value={name}
            invalid={!!fieldError}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Read for 20 minutes"
          />
        </Field>
      </form>
    </Dialog>
  );
}
