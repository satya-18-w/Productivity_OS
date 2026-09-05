// Package goals owns goals and their manual four-state progress (v1.md §10).
// Service is its entire public surface (ADR-0002). A goal is linked to nothing
// else in V1.
package goals

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

// Progress is a goal's manually set state (v1.md §10).
type Progress string

// The four progress states.
const (
	NotStarted Progress = "NOT_STARTED"
	InProgress Progress = "IN_PROGRESS"
	Achieved   Progress = "ACHIEVED"
	Abandoned  Progress = "ABANDONED"
)

func validProgress(p Progress) bool {
	switch p {
	case NotStarted, InProgress, Achieved, Abandoned:
		return true
	default:
		return false
	}
}

// CategoryChecker is the slice of the categories module goals needs, to validate a
// category_id a caller assigns to a goal. cmd/server wires this to
// categories.Service (ADR-0009).
type CategoryChecker interface {
	AssignableToAccount(ctx context.Context, accountID, categoryID uuid.UUID) (bool, error)
}

// ProgressReader is the slice of the tasks module goals needs, to embed each
// goal's derived task-completion progress in the list response (MX3; v1.md §10).
// cmd/server wires this to tasks.Service. done/total are keyed by goal id; a goal
// with no linked tasks is simply absent from both maps (0, 0).
type ProgressReader interface {
	ProgressByGoal(ctx context.Context, accountID uuid.UUID) (done, total map[uuid.UUID]int, err error)
}

// Goal is a goal record.
type Goal struct {
	ID          uuid.UUID
	Title       string
	Description string
	TargetDate  *timezone.Date
	Progress    Progress
	CategoryID  *uuid.UUID // optional (ADR-0009)
	CreatedAt   string     // RFC3339
	UpdatedAt   string     // RFC3339
}

// GoalInput is the editable field set.
type GoalInput struct {
	Title       string
	Description string
	TargetDate  *timezone.Date
	CategoryID  *uuid.UUID
}

// ErrGoalNotFound is returned when a goal is missing or not the caller's.
var ErrGoalNotFound = errors.New("goals: goal not found")

// ValidationError carries per-field messages for a 400 VALIDATION_ERROR.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "goals: validation failed" }

// Service is the goals module's published interface.
type Service interface {
	// CreateGoal / UpdateGoal: *ValidationError if CategoryID is set but not the
	// caller's or archived.
	CreateGoal(ctx context.Context, accountID uuid.UUID, in GoalInput) (Goal, error)
	UpdateGoal(ctx context.Context, accountID, goalID uuid.UUID, in GoalInput) error
	SetProgress(ctx context.Context, accountID, goalID uuid.UUID, p Progress) error
	DeleteGoal(ctx context.Context, accountID, goalID uuid.UUID) error
	ListGoals(ctx context.Context, accountID uuid.UUID) ([]Goal, error)

	// AssignableToAccount reports whether goalID exists and belongs to
	// accountID — used by tasks.GoalChecker to validate a goal_id a caller
	// links a task to (ADR-0009 pattern, mirrors categories.CategoryChecker).
	AssignableToAccount(ctx context.Context, accountID, goalID uuid.UUID) (bool, error)

	// CountByCategory implements categories.Counter for the categories overview
	// (ADR-0009).
	CountByCategory(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]int, error)
}
