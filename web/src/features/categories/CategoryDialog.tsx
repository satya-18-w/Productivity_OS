import { useId, useState, type FormEvent } from "react";
import { api, ApiError, type Category } from "../../api";
import { Dialog } from "../../components/ui/Dialog";
import { Button } from "../../components/ui/Button";
import { Field } from "../../components/ui/Field";
import { Input } from "../../components/ui/Input";

export type CategoryDialogTarget = { mode: "new" } | { mode: "rename"; category: Category };

export interface CategoryDialogProps {
  open: boolean;
  target: CategoryDialogTarget;
  onClose: () => void;
  onSaved: () => void | Promise<void>;
}

/** Create / rename a category — **name only** (`v1.md §2`: a flat, user-defined label). */
export function CategoryDialog({ open, target, onClose, onSaved }: CategoryDialogProps) {
  const renaming = target.mode === "rename";
  const formId = useId();
  const [name, setName] = useState(renaming ? target.category.name : "");
  const [error, setError] = useState("");
  const [fieldError, setFieldError] = useState("");
  const [busy, setBusy] = useState(false);
  // Mirrors the backend trim (categories service `validate`): whitespace-only
  // is empty, and padded names are stored trimmed.
  const trimmed = name.trim();

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setFieldError("");
    setBusy(true);
    try {
      if (target.mode === "rename") await api.renameCategory(target.category.id, trimmed);
      else await api.createCategory(trimmed);
      await onSaved();
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setFieldError("A category with this name already exists.");
      } else if (err instanceof ApiError && err.code === "VALIDATION_ERROR" && err.fields) {
        setFieldError(Object.values(err.fields)[0] ?? "Check the name.");
      } else {
        setError(`Could not ${renaming ? "rename" : "create"} the category.`);
      }
      setBusy(false);
    }
  }

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={renaming ? "Rename category" : "New category"}
      actions={
        <>
          <Button variant="secondary" onClick={onClose} disabled={busy}>
            Cancel
          </Button>
          <Button type="submit" form={formId} loading={busy} disabled={trimmed === ""}>
            {renaming ? "Save" : "Create"}
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
            maxLength={60}
            autoFocus
            value={name}
            invalid={!!fieldError}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Deep Work"
          />
        </Field>
      </form>
    </Dialog>
  );
}
