// Package habits owns daily habits and their completion history (v1.md §9).
// Service is its entire public surface (ADR-0002).
package habits

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

// AccountZone resolves an account's timezone, so "today" is the user's today.
// cmd/server wires this to the account module.
type AccountZone interface {
	Zone(ctx context.Context, accountID uuid.UUID) (*time.Location, error)
}

// CategoryChecker is the slice of the categories module habits needs, to validate
// a category_id a caller assigns to a habit. cmd/server wires this to
// categories.Service (ADR-0009).
type CategoryChecker interface {
	AssignableToAccount(ctx context.Context, accountID, categoryID uuid.UUID) (bool, error)
}

// Habit is a habit record.
type Habit struct {
	ID         uuid.UUID
	Name       string
	CategoryID *uuid.UUID // optional (ADR-0009)
	Target     *string    // optional free-text descriptor, display-only (MX3, v1.md §9)
	ArchivedAt *time.Time
	CreatedAt  time.Time // M7 Phase R1: needed to bound a habit's "active days" in a report range
}

// HabitInput is the data to create a habit.
type HabitInput struct {
	Name       string
	CategoryID *uuid.UUID
	Target     *string
}

// RangeCount is one habit's completion count over a date range (M6/M7 foundation).
type RangeCount struct {
	HabitID uuid.UUID
	Name    string
	Count   int
}

// HabitCompletion is one habit's completion on one date (M8 export).
type HabitCompletion struct {
	HabitID uuid.UUID
	Date    timezone.Date
}

// HabitHistoryEntry is one habit's completion dates within a caller-chosen range —
// the frontend's "This Month" heatmap (R2, docs/left.md Phase 6). Archived is true
// for a habit that is currently archived; it and its history still appear (a past
// range's heatmap should stay complete even for a habit archived since).
type HabitHistoryEntry struct {
	HabitID     uuid.UUID
	Name        string
	Archived    bool
	Completions []timezone.Date
}

// HabitWeekEntry is one active habit's current streak and which of a requested
// week's 7 dates it was completed on — the frontend's "This Week" grid (R2-follow-up,
// docs/left.md Phase 6).
type HabitWeekEntry struct {
	HabitID       uuid.UUID
	Name          string
	CurrentStreak int
	Completed     []timezone.Date
}

// ArchivedHabitName is an archived habit's id and name only, for a lightweight
// reference list (docs/left.md Phase 6 "This Week" grid).
type ArchivedHabitName struct {
	HabitID uuid.UUID
	Name    string
}

// WeekView is the ISO week (Monday-first, account timezone) containing a
// requested date: its 7 dates, every active habit's current streak and
// completions within that week, and the caller's archived habits by name only.
type WeekView struct {
	WeekStart timezone.Date
	Days      []timezone.Date
	Habits    []HabitWeekEntry
	Archived  []ArchivedHabitName
}

// HabitView is an active habit with its streak and per-date state, for the list.
type HabitView struct {
	ID              uuid.UUID
	Name            string
	CategoryID      *uuid.UUID
	Target          *string
	CurrentStreak   int
	CompletedOnDate bool // completed on the requested view date
	Last30Days      int  // completions in the 30 days ending today
}

// ErrHabitNotFound is returned when a habit is missing or not the caller's.
var ErrHabitNotFound = errors.New("habits: habit not found")

// ValidationError carries per-field messages for a 400 VALIDATION_ERROR.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "habits: validation failed" }

// Service is the habits module's published interface.
type Service interface {
	// CreateHabit adds an active habit. *ValidationError if CategoryID is set but
	// not the caller's or archived.
	CreateHabit(ctx context.Context, accountID uuid.UUID, in HabitInput) (Habit, error)

	// SetHabitCategory changes a habit's category (nil clears it) — its own
	// endpoint, separate from UpdateHabit, matching the ADR-0009 pattern shared
	// with tasks/goals. ErrHabitNotFound / *ValidationError as for CreateHabit.
	SetHabitCategory(ctx context.Context, accountID, habitID uuid.UUID, categoryID *uuid.UUID) error

	// UpdateHabit replaces a habit's name and target descriptor (MX3; v1.md §9).
	// name is always required (as for CreateHabit); target nil clears it.
	// ErrHabitNotFound if missing or not the caller's; *ValidationError for a bad
	// name.
	UpdateHabit(ctx context.Context, accountID, habitID uuid.UUID, name string, target *string) (Habit, error)

	// ArchiveHabit / UnarchiveHabit toggle the archived flag. Completions are
	// never touched (Q11). ErrHabitNotFound if missing or not the caller's.
	ArchiveHabit(ctx context.Context, accountID, habitID uuid.UUID) error
	UnarchiveHabit(ctx context.Context, accountID, habitID uuid.UUID) error

	// MarkComplete / UnmarkComplete set a habit's completion for one date.
	// MarkComplete is idempotent. Future dates are allowed (Q9).
	MarkComplete(ctx context.Context, accountID, habitID uuid.UUID, date timezone.Date) error
	UnmarkComplete(ctx context.Context, accountID, habitID uuid.UUID, date timezone.Date) error

	// ListActive returns the caller's active habits with the current streak and
	// whether each was completed on viewDate.
	ListActive(ctx context.Context, accountID uuid.UUID, viewDate timezone.Date) ([]HabitView, error)

	// ListArchived returns the caller's archived habits.
	ListArchived(ctx context.Context, accountID uuid.UUID) ([]Habit, error)

	// CountByCategory implements categories.Counter for the categories overview
	// (ADR-0009). Archived habits are excluded — they are already hidden from the
	// habits list, so they should not inflate a category's shown count.
	CountByCategory(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]int, error)

	// CompletionCountsInRange returns every habit (active and archived, so a past
	// range's history is complete) with how many of its completions fall in
	// [from, to] inclusive — 0 for a habit with none (M6/M7 foundation).
	// *ValidationError if to is before from.
	CompletionCountsInRange(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]RangeCount, error)

	// ListAll returns every habit the caller owns, active and archived, as raw
	// records (M8 export).
	ListAll(ctx context.Context, accountID uuid.UUID) ([]Habit, error)

	// AllCompletions returns every completion of every habit the caller owns,
	// active and archived (M8 export).
	AllCompletions(ctx context.Context, accountID uuid.UUID) ([]HabitCompletion, error)

	// History returns every habit (active and archived, each flagged) with its
	// completion dates within [from, to] inclusive — the frontend's "This Month"
	// heatmap (R2). *ValidationError if to is before from or the range exceeds the
	// server-side bound (92 days).
	History(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]HabitHistoryEntry, error)

	// Week returns the ISO week containing date, batched into one call — the
	// frontend's "This Week" grid (docs/left.md Phase 6), replacing 7 individual
	// ListActive-shaped calls with one.
	Week(ctx context.Context, accountID uuid.UUID, date timezone.Date) (WeekView, error)
}
