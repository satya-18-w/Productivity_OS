import { useCallback, useEffect, useId, useState, type FormEvent } from "react";
import { useSearchParams } from "react-router-dom";
import { ScreenLayout } from "../../shell/ScreenLayout";
import { PageHeader } from "../../components/layout/PageHeader";
import { Card } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { IconButton } from "../../components/ui/IconButton";
import { Field } from "../../components/ui/Field";
import { Textarea } from "../../components/ui/Textarea";
import { Chip } from "../../components/ui/Chip";
import { ChevronDownIcon } from "../../components/ui/icons";
import {
  todayISO,
  shiftDays,
  isoWeekRange,
  formatShortDate,
  parseISODate,
} from "../../components/date/dateUtils";
import { fmtDuration } from "../timeline/timelineFormat";
import {
  WEEKLY_REVIEW_PROMPTS,
  emptyWeeklyAnswers,
  fetchWeeklyReview,
  saveWeeklyReview,
  mockWeekReference,
  type WeeklyReviewAnswers,
  type WeekReference,
} from "./reviewData";

const ISO_RE = /^\d{4}-\d{2}-\d{2}$/;

/**
 * Weekly review (`v1.md §12`, Q2). No design spec / reference image — built from
 * shared primitives + form patterns, mirroring `DailyReviewScreen`.
 * Both the review record (four fixed Q2 prompts) and the weekly reference
 * totals run on deterministic in-memory mocks — no reviews/weekly-range backend
 * exists yet (`docs/left.md`). Week boundaries are ISO / Monday-first (D8).
 */
export function WeeklyReviewScreen() {
  const [params, setParams] = useSearchParams();
  const formId = useId();

  const rawWeek = params.get("week");
  const anchor =
    rawWeek && ISO_RE.test(rawWeek) && !Number.isNaN(parseISODate(rawWeek).getTime())
      ? rawWeek
      : todayISO();
  const [weekStart, weekEnd] = isoWeekRange(anchor);
  const [thisMonday] = isoWeekRange(todayISO());
  const isThisWeek = weekStart === thisMonday;

  const setAnchor = useCallback(
    (iso: string) =>
      setParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          const [targetMonday] = isoWeekRange(iso);
          if (targetMonday === thisMonday) next.delete("week");
          else next.set("week", iso);
          return next;
        },
        { replace: true },
      ),
    [setParams, thisMonday],
  );

  const [ref, setRef] = useState<WeekReference | null>(null);
  const [answers, setAnswers] = useState<WeeklyReviewAnswers>(emptyWeeklyAnswers());
  const [hasExisting, setHasExisting] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState("");

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    void Promise.all([mockWeekReference(weekStart, weekEnd), fetchWeeklyReview(weekStart)]).then(
      ([reference, existing]) => {
        if (cancelled) return;
        setRef(reference);
        setAnswers(existing?.answers ?? emptyWeeklyAnswers());
        setHasExisting(!!existing);
        setDirty(false);
        setSaveError("");
        setLoading(false);
      },
    );
    return () => {
      cancelled = true;
    };
  }, [weekStart, weekEnd]);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setSaveError("");
    try {
      await saveWeeklyReview(weekStart, answers);
      setHasExisting(true);
      setDirty(false);
    } catch {
      setSaveError("Could not save the review.");
    } finally {
      setSaving(false);
    }
  }

  function setAnswer(key: keyof WeeklyReviewAnswers, value: string) {
    setAnswers((a) => ({ ...a, [key]: value }));
    setDirty(true);
  }

  return (
    <ScreenLayout>
      <PageHeader
        eyebrow="Reviews"
        title="Weekly review"
        subtitle="Four fixed prompts, answered in your own words — no ratings or scores."
      />

      <div className="review-weekstepper">
        <IconButton label="Previous week" size="sm" onClick={() => setAnchor(shiftDays(anchor, -7))}>
          <ChevronDownIcon style={{ transform: "rotate(90deg)" }} width={16} height={16} />
        </IconButton>
        <span className="review-weekstepper__label" aria-live="polite">
          {formatShortDate(weekStart)} – {formatShortDate(weekEnd)}
        </span>
        <IconButton label="Next week" size="sm" onClick={() => setAnchor(shiftDays(anchor, 7))}>
          <ChevronDownIcon style={{ transform: "rotate(-90deg)" }} width={16} height={16} />
        </IconButton>
        <Button variant="secondary" size="sm" onClick={() => setAnchor(todayISO())} disabled={isThisWeek}>
          This week
        </Button>
      </div>

      <Card title="This week at a glance" headingLevel={2}>
        {loading || !ref ? (
          <p className="muted">Loading…</p>
        ) : (
          <>
            <p className="muted">⚠ Sample data — the weekly range endpoints are pending.</p>
            <div className="review-ref__cols">
              <div>
                <h3 className="review-ref__heading">Actual time by category</h3>
                <ul className="review-ref__chips">
                  {ref.categorySeconds.map((c) => (
                    <li key={c.name}>
                      <Chip>{c.name} · {fmtDuration(c.seconds)}</Chip>
                    </li>
                  ))}
                </ul>
              </div>
              <div>
                <h3 className="review-ref__heading">Habits</h3>
                <ul className="review-ref__habits">
                  {ref.habitCounts.map((h) => (
                    <li key={h.name}>
                      {h.name} · {h.count}/7 days
                    </li>
                  ))}
                </ul>
                <h3 className="review-ref__heading">Tasks completed</h3>
                <p className="review-ref__stat">{ref.tasksDone} entered Done this week</p>
              </div>
            </div>
          </>
        )}
      </Card>

      <Card title="Prompts" headingLevel={2}>
        {loading ? (
          <p className="muted">Loading…</p>
        ) : (
          <form id={formId} onSubmit={submit} className="review-form">
            {WEEKLY_REVIEW_PROMPTS.map((p) => (
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
