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

// Goal is a goal record.
type Goal struct {
	ID          uuid.UUID
	Title       string
	Description string
	TargetDate  *timezone.Date
	Progress    Progress
	CreatedAt   string // RFC3339
	UpdatedAt   string // RFC3339
}

// GoalInput is the editable field set.
type GoalInput struct {
	Title       string
	Description string
	TargetDate  *timezone.Date
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
	CreateGoal(ctx context.Context, accountID uuid.UUID, in GoalInput) (Goal, error)
	UpdateGoal(ctx context.Context, accountID, goalID uuid.UUID, in GoalInput) error
	SetProgress(ctx context.Context, accountID, goalID uuid.UUID, p Progress) error
	DeleteGoal(ctx context.Context, accountID, goalID uuid.UUID) error
	ListGoals(ctx context.Context, accountID uuid.UUID) ([]Goal, error)
}
