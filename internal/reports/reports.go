// Package reports is a read-only composition over timeline, habits, and tasks —
// the five deterministic reports of v1.md §13. It owns no table and performs no
// writes; every figure is computed from the other modules' own range-capable read
// methods (mostly the M6 Phase 1 foundation). Service is its entire public surface
// (ADR-0002).
package reports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/habits"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
	"github.com/satya-18-w/productivity-os/internal/timeline"
)

// AccountZone resolves an account's IANA timezone to a location, needed to
// convert the Task throughput report's date range into instants (ADR-0005).
// cmd/server wires this to the account module.
type AccountZone interface {
	Zone(ctx context.Context, accountID uuid.UUID) (*time.Location, error)
}

// TimelineReader is the slice of the timeline module reports needs. timeline.Service
// already satisfies this structurally.
type TimelineReader interface {
	ComparisonRange(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) (timeline.RangeComparison, error)
	DailyActualTotals(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]timeline.DayTotal, error)
}

// HabitsReader is the slice of the habits module reports needs. habits.Service
// already satisfies this structurally.
type HabitsReader interface {
	CompletionCountsInRange(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]habits.RangeCount, error)

	// ListAll supplies each habit's CreatedAt/ArchivedAt, needed to bound its
	// "active days" within a report range (a habit created or archived partway
	// through the range shouldn't be judged against the full range length).
	ListAll(ctx context.Context, accountID uuid.UUID) ([]habits.Habit, error)
}

// TasksReader is the slice of the tasks module reports needs. tasks.Service
// already satisfies this structurally.
type TasksReader interface {
	DoneCountInRange(ctx context.Context, accountID uuid.UUID, from, to time.Time) (int, error)
}

// ValidationError carries per-field messages for a 400 VALIDATION_ERROR.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "reports: validation failed" }

// CategoryTimeRow is one row of the "Time by category" report — actual time only,
// distinct from the fuller "Planned vs actual" report (v1.md §13).
type CategoryTimeRow struct {
	CategoryID    *uuid.UUID
	CategoryName  string
	ActualSeconds int64
}

// HabitCompletionRow is one row of the "Habit completion" report. RangeDays is the
// number of days within [From, To] the habit actually existed and was active (not
// archived) — not necessarily the full range length, e.g. a habit created or
// archived partway through. The caller (frontend) derives a rate as
// CompletedDays/RangeDays itself.
type HabitCompletionRow struct {
	HabitID       uuid.UUID
	Name          string
	CompletedDays int
	RangeDays     int
}

// TaskThroughputReport is the "Task throughput" report — a single figure for the
// whole range (Q10: distinct tasks with >=1 -> DONE transition in range).
type TaskThroughputReport struct {
	From      timezone.Date
	To        timezone.Date
	DoneCount int
}

// Report is the combined document over all five sub-reports for one range — the
// shape `GET /api/reports` returns (docs/left.md Phase 9). PlannedVsActual excludes
// the Uncategorized bucket (planned time is only meaningful per named category);
// TimeByCategory keeps it (Q8).
type Report struct {
	From              timezone.Date
	To                timezone.Date
	TimeByCategory    []CategoryTimeRow
	PlannedVsActual   []timeline.CategoryTotals
	HabitCompletion   []HabitCompletionRow
	TaskThroughput    int
	DailyActualTotals []timeline.DayTotal
}

// Service is the reports module's published interface. Every method is scoped to
// the account id the caller passes and takes a caller-chosen, inclusive date
// range; every method returns *ValidationError if To is before From.
type Service interface {
	// TimeByCategory is v1.md §13 report 1: total actual time per category.
	TimeByCategory(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]CategoryTimeRow, error)

	// PlannedVsActualByCategory is report 2: planned total, actual total, and
	// difference per category.
	PlannedVsActualByCategory(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]timeline.CategoryTotals, error)

	// HabitCompletion is report 3: for each habit (active and archived), the
	// number of completed days and the number of days it was active, across the
	// range.
	HabitCompletion(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]HabitCompletionRow, error)

	// TaskThroughput is report 4: the number of distinct tasks that entered DONE
	// within the range (Q10).
	TaskThroughput(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) (TaskThroughputReport, error)

	// DailyActualTotals is report 5: total actual time for each day in the range.
	DailyActualTotals(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]timeline.DayTotal, error)

	// Report combines all five sub-reports into the one document
	// GET /api/reports returns. *ValidationError if To is before From, or the
	// range exceeds the server-side bound.
	Report(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) (Report, error)
}
