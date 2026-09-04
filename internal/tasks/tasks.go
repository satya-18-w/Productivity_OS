// Package tasks owns tasks and the fixed four-column board (v1.md §7, §8). Service
// is its entire public surface (ADR-0002). Every state change is recorded in
// task_transitions so later milestones can count "tasks that entered DONE".
package tasks

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

// State is a task's column. The four are fixed and map exactly to the board
// columns, in this order (v1.md §8).
type State string

// The four fixed states, in board order.
const (
	Backlog    State = "BACKLOG"
	Todo       State = "TODO"
	InProgress State = "IN_PROGRESS"
	Done       State = "DONE"
)

// Columns is the fixed board order.
var Columns = []State{Backlog, Todo, InProgress, Done}

func validState(s State) bool {
	for _, c := range Columns {
		if c == s {
			return true
		}
	}
	return false
}

// Task is a single task.
type Task struct {
	ID          uuid.UUID
	Title       string
	Description string
	DueDate     *timezone.Date // plain calendar date, no time (v1.md §7)
	State       State
	CreatedAt   string // RFC3339
	UpdatedAt   string // RFC3339
}

// TaskInput is the editable field set (v1.md §7: title, description, due date).
type TaskInput struct {
	Title       string
	Description string
	DueDate     *timezone.Date
}

// Column is one board column with its tasks, newest first.
type Column struct {
	State State
	Tasks []Task
}

// Board is the whole board: always four columns, in Columns order.
type Board struct {
	Columns []Column
}

// ErrTaskNotFound is returned when a task is missing or not the caller's.
var ErrTaskNotFound = errors.New("tasks: task not found")

// ValidationError carries per-field messages for a 400 VALIDATION_ERROR.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "tasks: validation failed" }

// Service is the tasks module's published interface. Every method scopes to the
// account id it is passed, taken only from the request context.
type Service interface {
	// CreateTask adds a task in BACKLOG and records the creation transition.
	CreateTask(ctx context.Context, accountID uuid.UUID, in TaskInput) (Task, error)

	// UpdateTask changes a task's title, description, and due date (not its
	// state). ErrTaskNotFound if it is missing or not the caller's.
	UpdateTask(ctx context.Context, accountID, taskID uuid.UUID, in TaskInput) error

	// MoveTask sets a task's state to any of the four, recording a transition
	// unless the state is unchanged. ErrTaskNotFound / *ValidationError.
	MoveTask(ctx context.Context, accountID, taskID uuid.UUID, to State) error

	// DeleteTask removes a task (and its transitions, by cascade).
	DeleteTask(ctx context.Context, accountID, taskID uuid.UUID) error

	// Board returns the caller's four columns, each newest-first.
	Board(ctx context.Context, accountID uuid.UUID) (Board, error)
}
