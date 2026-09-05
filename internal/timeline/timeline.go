// Package timeline owns time blocks (planned and actual) and the per-date views
// over them — the product's core (v1.md §3–§6). Categories live in their own module
// (ADR-0009); timeline reaches them through CategoryStore. Service is its entire
// public surface (ADR-0002).
package timeline

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

// AccountZone resolves an account's IANA timezone to a location. cmd/server wires
// this to the account module's published interface, keeping timeline decoupled
// from account (ADR-0002).
type AccountZone interface {
	Zone(ctx context.Context, accountID uuid.UUID) (*time.Location, error)
}

// CategoryStore is the slice of the categories module that timeline needs:
// validating a category a caller assigns to a block, and resolving category names
// for the read models. cmd/server wires this to categories.Service (ADR-0009).
type CategoryStore interface {
	AssignableToAccount(ctx context.Context, accountID, categoryID uuid.UUID) (bool, error)
	NamesForAccount(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]string, error)
}

// TaskChecker is the slice of the tasks module that timeline needs: validating a
// task a caller links a block to, and resolving each task's own category so a
// task-linked block can inherit it (MX-TL). cmd/server wires this to tasks.Service.
type TaskChecker interface {
	AssignableToAccount(ctx context.Context, accountID, taskID uuid.UUID) (bool, error)
	CategoriesForTasks(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]*uuid.UUID, error)
}

// BlockKind is "planned" or "actual" — fixed when a block is created (v1.md §3, §4).
type BlockKind string

// The two block kinds.
const (
	Planned BlockKind = "planned"
	Actual  BlockKind = "actual"
)

// Block is a time block: a start instant, an end instant strictly after it, and
// either an optional category or an optional link to a task, never both (MX-TL).
type Block struct {
	ID           uuid.UUID
	Kind         BlockKind
	StartsAt     time.Time
	EndsAt       time.Time
	CategoryID   *uuid.UUID // only meaningful when TaskID is nil
	CategoryName *string    // populated by read models only; resolved via TaskID when linked
	TaskID       *uuid.UUID // optional link to a task (MX-TL)
}

// BlockInput is the data to create a block.
type BlockInput struct {
	Kind       BlockKind
	StartsAt   time.Time
	EndsAt     time.Time
	CategoryID *uuid.UUID
	TaskID     *uuid.UUID
}

// ErrBlockNotFound is returned when a block is missing or not the caller's; the
// handler maps it to 404.
var ErrBlockNotFound = errors.New("timeline: time block not found")

// ErrTaskNotFound is returned by BlocksForTask when the task is missing or not
// the caller's; the handler maps it to 404.
var ErrTaskNotFound = errors.New("timeline: task not found")

// ValidationError carries per-field messages for a 400 VALIDATION_ERROR.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "timeline: validation failed" }

// Service is the timeline module's published interface. Every method is scoped to
// the account id the caller passes, which it takes only from the request context.
type Service interface {
	// AddBlock creates a planned or actual block. *ValidationError for a bad kind,
	// a non-positive range, a category that is not the caller's or is archived, a
	// task that is not the caller's (MX-TL), or both CategoryID and TaskID set on
	// the same input (a task-linked block inherits the task's category and cannot
	// also carry its own).
	AddBlock(ctx context.Context, accountID uuid.UUID, in BlockInput) (Block, error)

	// EditBlock changes a block's start, end, category, and task link. Kind is
	// immutable. ErrBlockNotFound if it is missing or not the caller's;
	// *ValidationError as for AddBlock.
	EditBlock(ctx context.Context, accountID, blockID uuid.UUID, starts, ends time.Time, categoryID, taskID *uuid.UUID) error

	// DeleteBlock removes a block. ErrBlockNotFound if missing or not the caller's.
	DeleteBlock(ctx context.Context, accountID, blockID uuid.UUID) error

	// Timeline returns the caller's planned and actual blocks that overlap the
	// given date, resolved in the account's timezone (v1.md §5).
	Timeline(ctx context.Context, accountID uuid.UUID, date timezone.Date) (DayTimeline, error)

	// TimelineRange is Timeline batched over every day in [from, to] inclusive —
	// one query instead of one per day, for the Week/Month views (v1.md §5's
	// Week/Month amendment; same block data, no new fields). *ValidationError if
	// to is before from or the range exceeds the server-side bound (62 days,
	// enough for a month grid's 42 cells).
	TimelineRange(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) (RangeTimeline, error)

	// Comparison returns per-category planned/actual/difference totals for the
	// given date, with a nil-category "Uncategorized" bucket. Overlapping blocks
	// are summed; a midnight-spanning block counts only its portion inside the
	// date's window (v1.md §6, N4).
	Comparison(ctx context.Context, accountID uuid.UUID, date timezone.Date) (DayComparison, error)

	// ComparisonRange is Comparison summed over every day in [from, to] inclusive
	// (M6/M7 foundation). *ValidationError if to is before from.
	ComparisonRange(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) (RangeComparison, error)

	// DailyActualTotals returns total actual (not planned) time for each day in
	// [from, to] inclusive, in the account's timezone (M7 "Daily actual totals"
	// report, v1.md §13). *ValidationError if to is before from.
	DailyActualTotals(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]DayTotal, error)

	// CountByCategory implements categories.Counter for the categories overview
	// (ADR-0009).
	CountByCategory(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]int, error)

	// ListAllBlocks returns every planned and actual block the caller owns,
	// unbounded, ordered by start (M8 export).
	ListAllBlocks(ctx context.Context, accountID uuid.UUID) ([]Block, error)

	// BlocksForTask returns every block (planned and actual, across any date)
	// linked to taskID, ordered by start — the reverse direction of a block's
	// TaskID link (v1.md §7). ErrTaskNotFound if taskID is missing or not the
	// caller's.
	BlocksForTask(ctx context.Context, accountID, taskID uuid.UUID) ([]Block, error)
}

// PositionedBlock is a block with its placement on a specific date's 24-hour
// grid, computed server-side in the account's timezone (no client-side tz math).
type PositionedBlock struct {
	Block
	StartMinute int  // wall-clock minute-of-day on the queried date, 0–1440
	EndMinute   int  // wall-clock minute-of-day, 0–1440
	FromPrevDay bool // the block started before this date
	ToNextDay   bool // the block ends after this date

	// The block's own wall-clock values in the account zone, for the editor.
	LocalDate   string // YYYY-MM-DD (the block's start date)
	LocalStart  string // HH:MM
	LocalEnd    string // HH:MM
	EndsNextDay bool   // the block's own end is on the day after LocalDate
}

// DayTimeline is a date's planned and actual blocks, positioned for display.
type DayTimeline struct {
	Date    timezone.Date
	Planned []PositionedBlock
	Actual  []PositionedBlock
}

// RangeTimeline is TimelineRange's result — every day in [From, To] inclusive,
// each already split into Planned/Actual exactly as Timeline would return it.
type RangeTimeline struct {
	From timezone.Date
	To   timezone.Date
	Days []DayTimeline
}

// CategoryTotals is one row of a DayComparison. CategoryID is nil for the
// Uncategorized bucket.
type CategoryTotals struct {
	CategoryID        *uuid.UUID
	CategoryName      string
	PlannedSeconds    int64
	ActualSeconds     int64
	DifferenceSeconds int64 // actual - planned
}

// DayComparison is the per-category planned-vs-actual view for one date.
type DayComparison struct {
	Date       timezone.Date
	Categories []CategoryTotals
}

// RangeComparison is the per-category planned-vs-actual view summed over
// [From, To] inclusive.
type RangeComparison struct {
	From       timezone.Date
	To         timezone.Date
	Categories []CategoryTotals
}

// DayTotal is one date's total actual time (v1.md §13 "Daily actual totals").
type DayTotal struct {
	Date          timezone.Date
	ActualSeconds int64
}
