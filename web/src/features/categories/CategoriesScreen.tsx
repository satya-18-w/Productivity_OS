import { useCallback, useEffect, useState } from "react";
import { api, type Category } from "../../api";
import { ScreenLayout } from "../../shell/ScreenLayout";
import { PageHeader } from "../../components/layout/PageHeader";
import { Button } from "../../components/ui/Button";
import { EmptyState, ErrorState } from "../../components/productivity/states";
import { CategoryRow } from "./CategoryRow";
import { CategoryDialog, type CategoryDialogTarget } from "./CategoryDialog";

export function CategoriesScreen() {
  const [categories, setCategories] = useState<Category[] | null>(null);
  const [error, setError] = useState(false);
  const [dialog, setDialog] = useState<CategoryDialogTarget | null>(null);

  const load = useCallback(async () => {
    setError(false);
    try {
      setCategories(await api.listCategories());
    } catch {
      setError(true);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function archive(category: Category) {
    if (!window.confirm(`Archive "${category.name}"? It will no longer be offered for new blocks; existing blocks keep it.`)) {
      return;
    }
    try {
      await api.archiveCategory(category.id);
      await load();
    } catch {
      setError(true);
    }
  }

  return (
    <ScreenLayout>
      <PageHeader
        eyebrow="Categories"
        title="Categories"
        subtitle="Flat labels for your time blocks."
        actions={<Button onClick={() => setDialog({ mode: "new" })}>New category</Button>}
      />

      {error ? (
        <ErrorState message="Could not load your categories." action={<Button onClick={load}>Retry</Button>} />
      ) : !categories ? (
        <p className="muted">Loading…</p>
      ) : categories.length === 0 ? (
        <EmptyState
          title="No categories yet"
          message="Create a category to label your planned and actual time blocks."
          action={<Button onClick={() => setDialog({ mode: "new" })}>New category</Button>}
        />
      ) : (
        <ul className="category-list">
          {categories.map((c) => (
            <CategoryRow
              key={c.id}
              category={c}
              onRename={(cat) => setDialog({ mode: "rename", category: cat })}
              onArchive={archive}
            />
          ))}
        </ul>
      )}

      {dialog && (
        <CategoryDialog
          open
          target={dialog}
          onClose={() => setDialog(null)}
          onSaved={async () => {
            setDialog(null);
            await load();
          }}
        />
      )}
    </ScreenLayout>
  );
}
