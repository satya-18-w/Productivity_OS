// Package reviews owns daily and weekly review answers against a fixed,
// non-editable prompt set (v1.md §11, §12). It does not compute or store any
// reference data (time-by-category, habit counts, task throughput) — that is
// assembled by the frontend from each domain module's own range-capable read
// endpoints (M6 decision), keeping this module thin and self-contained. Service is
// its entire public surface (ADR-0002).
package reviews

import (
	"context"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

// Prompt is one fixed review question.
type Prompt struct {
	Key  string
	Text string
}

// DailyPrompts is the fixed, non-editable daily-review prompt set (v1.md §11).
// Placeholder wording (Q1, resolved 2026-09-04) — the product owner may replace
// the Text values freely; Key is the stable identity an answer is stored under.
var DailyPrompts = []Prompt{
	{Key: "went_well", Text: "What went well today?"},
	{Key: "not_as_planned", Text: "What didn't go as planned?"},
	{Key: "differently_tomorrow", Text: "What will you do differently tomorrow?"},
	{Key: "grateful_for", Text: "One thing you're grateful for."},
}

// WeeklyPrompts is the fixed, non-editable weekly-review prompt set (v1.md §12).
// Placeholder wording (Q2, resolved 2026-09-04).
var WeeklyPrompts = []Prompt{
	{Key: "highlights", Text: "What were the highlights of this week?"},
	{Key: "struggled_with", Text: "What did you struggle with?"},
	{Key: "time_intended", Text: "Did your time go where you intended?"},
	{Key: "next_priority", Text: "What is the one priority for next week?"},
}

// Answers is a {prompt key: free text} map, saved and loaded whole — never
// queried per-key.
type Answers map[string]string

// DailyReview is one account's daily review for one date. UpdatedAt is "" if the
// date has never been saved — GetDaily returns this rather than an error, since an
// unanswered date is simply blank (v1.md §11), not missing.
type DailyReview struct {
	Date      timezone.Date
	Answers   Answers
	UpdatedAt string // RFC3339, "" if never saved
}

// WeeklyReview is the weekly equivalent, keyed by ISO year and week.
type WeeklyReview struct {
	ISOYear   int
	ISOWeek   int
	Answers   Answers
	UpdatedAt string // RFC3339, "" if never saved
}

// ValidationError carries per-field messages for a 400 VALIDATION_ERROR.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "reviews: validation failed" }

// Service is the reviews module's published interface. Every method is scoped to
// the account id the caller passes, which it takes only from the request context.
type Service interface {
	// GetDaily returns the caller's daily review for date — a zero-answers review
	// if none has been saved yet.
	GetDaily(ctx context.Context, accountID uuid.UUID, date timezone.Date) (DailyReview, error)

	// SaveDaily upserts the caller's whole daily answer set for date. "Complete"
	// and "edit" are the same operation (v1.md §11). Keys not in DailyPrompts are
	// dropped silently; an empty map is allowed (a partially-filled review).
	SaveDaily(ctx context.Context, accountID uuid.UUID, date timezone.Date, answers Answers) (DailyReview, error)

	// GetWeekly returns the caller's weekly review for the given ISO year/week —
	// a zero-answers review if none has been saved yet. *ValidationError if
	// isoWeek is out of [1, 53].
	GetWeekly(ctx context.Context, accountID uuid.UUID, isoYear, isoWeek int) (WeeklyReview, error)

	// SaveWeekly upserts the caller's whole weekly answer set. Keys not in
	// WeeklyPrompts are dropped silently. *ValidationError as for GetWeekly.
	SaveWeekly(ctx context.Context, accountID uuid.UUID, isoYear, isoWeek int, answers Answers) (WeeklyReview, error)

	// ListDaily returns every daily review the caller has saved (M8 export).
	ListDaily(ctx context.Context, accountID uuid.UUID) ([]DailyReview, error)

	// ListWeekly returns every weekly review the caller has saved (M8 export).
	ListWeekly(ctx context.Context, accountID uuid.UUID) ([]WeeklyReview, error)
}
