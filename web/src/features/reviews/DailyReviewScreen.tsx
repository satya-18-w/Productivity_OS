import { useCallback, useEffect, useId, useState, type FormEvent } from "react";
import { useSearchParams } from "react-router-dom";
import { api, type ComparisonRow, type HabitView } from "../../api";
import { ScreenLayout } from "../../shell/ScreenLayout";
import { PageHeader } from "../../components/layout/PageHeader";
import { Card } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { Field } from "../../components/ui/Field";
import { Textarea } from "../../components/ui/Textarea";
import { Chip } from "../../components/ui/Chip";
import { ErrorState } from "../../components/productivity/states";
import { DateStepper } from "../../components/date/DateStepper";
import { todayISO, formatFullDate, parseISODate } from "../../components/date/dateUtils";
import { categoryColor } from "../../components/productivity/categoryColor";
import { CheckIcon } from "../../components/ui/icons";
import { fmtDuration } from "../timeline/timelineFormat";
import { DAILY_REVIEW_PROMPTS, emptyAnswers, fetchDailyReview, saveDailyReview, type DailyReviewAnswers } from "./reviewData";

const ISO_RE = /^\d{4}-\d{2}-\d{2}$/;

/**
 * Daily review (`v1.md §11`, Q1). No design spec / reference image — built from
 * shared primitives + form patterns. Reference totals (actual time per category,
 * habits completed) come from the real Timeline/Habits APIs; the review itself
 * (four fixed prompts, free text) runs on an in-memory mock — no reviews backend
 * exists yet (`docs/left.md`, "Phase 10 — Daily Review").
 */
export function DailyReviewScreen() {
  const [params, setParams] = useSearchParams();
  const formId = useId();

  const rawDate = params.get("date");
  const date =
    rawDate && ISO_RE.test(rawDate) && !Number.isNaN(parseISODate(rawDate).getTime())
      ? rawDate
      : todayISO();

  const setDate = useCallback(
    (iso: string) =>
      setParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          if (iso === todayISO()) next.delete("date");
          else next.set("date", iso);
          return next;
        },
        { replace: true },
      ),
    [setParams],
  );

  const [categories, setCategories] = useState<ComparisonRow[]>([]);
  const [habits, setHabits] = useState<HabitView[]>([]);
  const [refError, setRefError] = useState(false);
  const [refLoading, setRefLoading] = useState(true);

  const loadReference = useCallback(async () => {
    setRefLoading(true);
    setRefError(false);
    try {
      const [cmp, list] = await Promise.all([api.comparison(date), api.habits(date)]);
      setCategories(cmp.categories.filter((c) => c.actual_seconds > 0));
      setHabits(list.habits);
    } catch {
      setRefError(true);
    } finally {
      setRefLoading(false);
    }
  }, [date]);

  useEffect(() => {
    void loadReference();
  }, [loadReference]);

  const [answers, setAnswers] = useState<DailyReviewAnswers>(emptyAnswers());
  const [hasExisting, setHasExisting] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [reviewLoading, setReviewLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState("");

  useEffect(() => {
    let cancelled = false;
    setReviewLoading(true);
    void fetchDailyReview(date).then((existing) => {
      if (cancelled) return;
      setAnswers(existing?.answers ?? emptyAnswers());
      setHasExisting(!!existing);
      setDirty(false);
      setSaveError("");
      setReviewLoading(false);
    });
    return () => {
      cancelled = true;
    };
  }, [date]);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setSaveError("");
    try {
      await saveDailyReview(date, answers);
      setHasExisting(true);
      setDirty(false);
    } catch {
      setSaveError("Could not save the review.");
    } finally {
      setSaving(false);
    }
  }

  function setAnswer(key: keyof DailyReviewAnswers, value: string) {
    setAnswers((a) => ({ ...a, [key]: value }));
    setDirty(true);
  }

  return (
    <ScreenLayout>
      <PageHeader
        eyebrow="Reviews"
        title="Daily review"
        subtitle="Four fixed prompts, answered in your own words — no ratings or scores."
      />

      <DateStepper value={date} onChange={setDate} label="Review date" />

      <Card title={`${formatFullDate(date)} at a glance`} headingLevel={2}>
        {refLoading ? (
          <p className="muted">Loading…</p>
        ) : refError ? (
          <ErrorState message="Could not load this date's totals." action={<Button onClick={loadReference}>Retry</Button>} />
        ) : (
          <div className="review-ref__cols">
            <div>
              <h3 className="review-ref__heading">Actual time by category</h3>
              {categories.length === 0 ? (
                <p className="muted">No time logged for this date.</p>
              ) : (
                <ul className="review-ref__chips">
                  {categories.map((c) => (
                    <li key={c.category_id ?? "uncategorized"}>
                      <Chip dotColor={categoryColor(c.category_id)}>
                        {c.category_name} · {fmtDuration(c.actual_seconds)}
                      </Chip>
                    </li>
                  ))}
                </ul>
              )}
            </div>
            <div>
              <h3 className="review-ref__heading">Habits</h3>
              {habits.length === 0 ? (
                <p className="muted">No habits yet.</p>
              ) : (
                <ul className="review-ref__habits">
                  {habits.map((h) => (
                    <li key={h.id} className={h.completed_on_date ? "is-done" : undefined}>
                      {h.completed_on_date && <CheckIcon width={14} height={14} aria-hidden="true" />}
                      {h.name}
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>
        )}
      </Card>

      <Card title="Prompts" headingLevel={2}>
        {reviewLoading ? (
          <p className="muted">Loading…</p>
        ) : (
          <form id={formId} onSubmit={submit} className="review-form">
            {DAILY_REVIEW_PROMPTS.map((p) => (
              <Field key={p.key} label={p.label} htmlFor={`${formId}-${p.key}`}>
                <Textarea
                  id={`${formId}-${p.key}`}
                  rows={3}
                  maxLength={5000}
                  value={answers[p.key]}
                  onChange={(e) => setAnswer(p.key, e.target.value)}
                />
              </Field>
            ))}
            {saveError && (
              <p className="error" role="alert">
                {saveError}
              </p>
            )}
            <div className="review-actions">
              {hasExisting && !dirty && !saving && <span className="muted">Saved</span>}
              <Button type="submit" loading={saving}>
                {hasExisting ? "Save changes" : "Save review"}
              </Button>
            </div>
          </form>
        )}
      </Card>
    </ScreenLayout>
  );
}
