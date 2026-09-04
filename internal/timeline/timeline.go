// Package timeline owns categories and time blocks (planned and actual) and the
// per-date views over them — the product's core (v1.md §2–§6). Service is its
// entire public surface (ADR-0002).
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

// Category is a flat, user-defined label for time blocks (v1.md §2).
type Category struct {
	ID         uuid.UUID
	Name       string
	ArchivedAt *time.Time // nil while active
}

// BlockKind is "planned" or "actual" — fixed when a block is created (v1.md §3, §4).
type BlockKind string

// The two block kinds.
const (
	Planned BlockKind = "planned"
	Actual  BlockKind = "actual"
)

// Block is a time block: a start instant, an end instant strictly after it, and an
// optional category.
type Block struct {
	ID           uuid.UUID
	Kind         BlockKind
	StartsAt     time.Time
	EndsAt       time.Time
	CategoryID   *uuid.UUID
	CategoryName *string // populated by read models only
}

// BlockInput is the data to create a block.
type BlockInput struct {
	Kind       BlockKind
	StartsAt   time.Time
	EndsAt     time.Time
	CategoryID *uuid.UUID
}

// Sentinel errors mapped to HTTP status codes by the handler.
var (
	ErrCategoryNameTaken = errors.New("timeline: an active category with this name already exists")
	ErrCategoryNotFound  = errors.New("timeline: category not found")
	ErrBlockNotFound     = errors.New("timeline: time block not found")
)

// ValidationError carries per-field messages for a 400 VALIDATION_ERROR.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "timeline: validation failed" }

// Service is the timeline module's published interface. Every method is scoped to
// the account id the caller passes, which it takes only from the request context.
type Service interface {
	// CreateCategory adds an active category. ErrCategoryNameTaken if an active
	// one already has that name (case-insensitive).
	CreateCategory(ctx context.Context, accountID uuid.UUID, name string) (Category, error)

	// RenameCategory changes an active category's name. ErrCategoryNotFound if it
	// does not exist, is not the caller's, or is archived. ErrCategoryNameTaken
	// on a collision.
	RenameCategory(ctx context.Context, accountID, categoryID uuid.UUID, name string) error

	// ArchiveCategory archives an active category so it is no longer offered.
	// Blocks already assigned to it keep it (v1.md §2). ErrCategoryNotFound if it
	// is missing, not the caller's, or already archived.
	ArchiveCategory(ctx context.Context, accountID, categoryID uuid.UUID) error

	// ListActiveCategories returns the caller's active categories, name-ordered.
	ListActiveCategories(ctx context.Context, accountID uuid.UUID) ([]Category, error)

	// AddBlock creates a planned or actual block. *ValidationError for a bad kind,
	// a non-positive range, or a category that is not the caller's or is archived.
	AddBlock(ctx context.Context, accountID uuid.UUID, in BlockInput) (Block, error)

	// EditBlock changes a block's start, end, and category. Kind is immutable.
	// ErrBlockNotFound if it is missing or not the caller's; *ValidationError as
	// for AddBlock.
	EditBlock(ctx context.Context, accountID, blockID uuid.UUID, starts, ends time.Time, categoryID *uuid.UUID) error

	// DeleteBlock removes a block. ErrBlockNotFound if missing or not the caller's.
	DeleteBlock(ctx context.Context, accountID, blockID uuid.UUID) error

	// Timeline returns the caller's planned and actual blocks that overlap the
	// given date, resolved in the account's timezone (v1.md §5).
	Timeline(ctx context.Context, accountID uuid.UUID, date timezone.Date) (DayTimeline, error)

	// Comparison returns per-category planned/actual/difference totals for the
	// given date, with a nil-category "Uncategorized" bucket. Overlapping blocks
	// are summed; a midnight-spanning block counts only its portion inside the
	// date's window (v1.md §6, N4).
	Comparison(ctx context.Context, accountID uuid.UUID, date timezone.Date) (DayComparison, error)
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
