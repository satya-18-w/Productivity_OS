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

// Habit is a habit record.
type Habit struct {
	ID         uuid.UUID
	Name       string
	ArchivedAt *time.Time
}

// HabitView is an active habit with its streak and per-date state, for the list.
type HabitView struct {
	ID              uuid.UUID
	Name            string
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
	// CreateHabit adds an active habit.
	CreateHabit(ctx context.Context, accountID uuid.UUID, name string) (Habit, error)

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
}
