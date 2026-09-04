import { useCallback, useEffect, useState, type FormEvent } from "react";
import { api, ApiError, type Category } from "../api";

export function Categories() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      setCategories(await api.listCategories());
      setError("");
    } catch {
      setError("Could not load categories.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <div className="stack">
      <section className="card">
        <h2>Categories</h2>
        <p className="muted">Flat labels for your time blocks. Archiving hides a category from new blocks; existing blocks keep it.</p>
        {error && <p className="error" role="alert">{error}</p>}
        <NewCategoryForm onCreated={load} />
      </section>

      <section className="card">
        <h3>Active</h3>
        {loading ? (
          <p className="muted">Loading…</p>
        ) : categories.length === 0 ? (
          <p className="muted">No categories yet.</p>
        ) : (
          <ul className="rows">
            {categories.map((c) => (
              <CategoryRow key={c.id} category={c} onChanged={load} />
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}

function NewCategoryForm({ onCreated }: { onCreated: () => Promise<void> }) {
  const [name, setName] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      await api.createCategory(name);
      setName("");
      await onCreated();
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) setError("A category with this name already exists.");
      else if (err instanceof ApiError && err.fields?.name) setError(err.fields.name);
      else setError("Could not create the category.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form className="inline-form" onSubmit={submit}>
      <input
        type="text"
        placeholder="New category name"
        required
        maxLength={60}
        value={name}
        onChange={(e) => setName(e.target.value)}
      />
      <button type="submit" disabled={busy}>Add</button>
      {error && <span className="field-error">{error}</span>}
    </form>
  );
}

function CategoryRow({ category, onChanged }: { category: Category; onChanged: () => Promise<void> }) {
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(category.name);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function rename(e: FormEvent) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      await api.renameCategory(category.id, name);
      setEditing(false);
      await onChanged();
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) setError("Name already in use.");
      else if (err instanceof ApiError && err.fields?.name) setError(err.fields.name);
      else setError("Could not rename.");
    } finally {
      setBusy(false);
    }
  }

  async function archive() {
    if (!window.confirm(`Archive "${category.name}"? It will no longer be offered for new blocks.`)) return;
    setBusy(true);
    try {
      await api.archiveCategory(category.id);
      await onChanged();
    } catch {
      setError("Could not archive.");
      setBusy(false);
    }
  }

  return (
    <li className="row">
      {editing ? (
        <form className="inline-form" onSubmit={rename}>
          <input type="text" required maxLength={60} value={name} onChange={(e) => setName(e.target.value)} autoFocus />
          <button type="submit" disabled={busy}>Save</button>
          <button type="button" className="link" onClick={() => { setEditing(false); setName(category.name); setError(""); }}>
            Cancel
          </button>
        </form>
      ) : (
        <>
          <span className="row-name">{category.name}</span>
          <span className="row-actions">
            <button type="button" className="link" onClick={() => setEditing(true)}>Rename</button>
            <button type="button" className="link danger" onClick={archive} disabled={busy}>Archive</button>
          </span>
        </>
      )}
      {error && <span className="field-error">{error}</span>}
    </li>
  );
}
