import { useState } from "react";
import { api, browserTimezone } from "../../api";
import { ScreenLayout } from "../../shell/ScreenLayout";
import { PageHeader } from "../../components/layout/PageHeader";
import { Card } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { todayISO } from "../../components/date/dateUtils";

type Status = "idle" | "working" | "ready" | "error";

/**
 * Data export (`v1.md §14`, P5). Single user-initiated download of everything
 * the account owns that the API exposes today. Format is a single JSON
 * document — PROVISIONAL pending Q3 (single JSON vs CSV archive); reviews join
 * the snapshot once their backend lands (`docs/left.md`).
 */
export function ExportScreen() {
  const [status, setStatus] = useState<Status>("idle");
  const [url, setUrl] = useState<string | null>(null);
  const [fileName, setFileName] = useState("");
  const [error, setError] = useState("");

  async function runExport() {
    setStatus("working");
    setError("");
    if (url) URL.revokeObjectURL(url);
    setUrl(null);
    try {
      const [account, categories, board, habits, goals] = await Promise.all([
        api.getAccount(),
        api.listCategories(),
        api.board(),
        api.habits(),
        api.goals(),
      ]);
      const snapshot = {
        format: "productivity-os-export/1",
        exported_at: new Date().toISOString(),
        account: { email: account.email, timezone: account.timezone || browserTimezone() },
        categories,
        tasks: board.columns.flatMap((c) => c.tasks.map((t) => ({ ...t, state: c.state }))),
        habits: { active: habits.habits, archived: habits.archived },
        goals,
        reviews: { note: "reviews join this snapshot once the reviews backend lands" },
      };
      const blob = new Blob([JSON.stringify(snapshot, null, 2)], { type: "application/json" });
      const name = `productivity-os-export-${todayISO()}.json`;
      setFileName(name);
      setUrl(URL.createObjectURL(blob));
      setStatus("ready");
    } catch {
      setStatus("error");
      setError("Could not gather your data. Try again.");
    }
  }

  return (
    <ScreenLayout>
      <PageHeader
        eyebrow="Export"
        title="Export my data"
        subtitle="One file with everything you created — yours to keep, no support needed."
      />
      <Card title="What's included" headingLevel={2}>
        <p className="muted">⚠ Provisional format (single JSON) pending Q3.</p>
        <ul className="ui-list">
          <li>Categories</li>
          <li>Tasks (all four states)</li>
          <li>Habits, active and archived</li>
          <li>Goals with their progress states</li>
        </ul>
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        <div className="review-actions">
          {status === "ready" && url && (
            <a className="ui-btn ui-btn--secondary" href={url} download={fileName}>
              Download {fileName}
            </a>
          )}
          <Button onClick={runExport} loading={status === "working"}>
            {status === "ready" ? "Export again" : "Export my data"}
          </Button>
        </div>
      </Card>
    </ScreenLayout>
  );
}
