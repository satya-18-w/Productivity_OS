// Package tasks owns tasks and the fixed four-column board (v1.md §7, §8). Service
// is its entire public surface (ADR-0002). Every state change is recorded in
// task_transitions so later milestones can count "tasks that entered DONE".
package tasks

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

// AccountZone resolves an account's IANA timezone to a location, so the HTTP layer
// can convert a `from`/`to` date range into instants for DoneCountInRange without
// client-side tz math (ADR-0005). cmd/server wires this to the account module.
type AccountZone interface {
	Zone(ctx context.Context, accountID uuid.UUID) (*time.Location, error)
}

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

// Priority is a task's optional, purely-informational priority label (MX3-follow-up;
// v1.md §7). It has no effect on ordering, board layout, or any other behaviour.
type Priority string

// The three fixed priority levels.
const (
	High   Priority = "HIGH"
	Medium Priority = "MEDIUM"
	Low    Priority = "LOW"
)

func validPriority(p Priority) bool {
	switch p {
	case High, Medium, Low:
		return true
	default:
		return false
	}
}

// CategoryChecker is the slice of the categories module tasks needs, to validate a
// category_id a caller assigns to a task. cmd/server wires this to
// categories.Service (ADR-0009).
type CategoryChecker interface {
	AssignableToAccount(ctx context.Context, accountID, categoryID uuid.UUID) (bool, error)
}

// GoalChecker is the slice of the goals module tasks needs, to validate a goal_id a
// caller links a task to (MX3; v1.md §10). cmd/server wires this to goals.Service
// (ADR-0009 pattern).
type GoalChecker interface {
	AssignableToAccount(ctx context.Context, accountID, goalID uuid.UUID) (bool, error)
}

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
	CategoryID  *uuid.UUID // optional (ADR-0009)
	GoalID      *uuid.UUID // optional link to a goal (MX3; v1.md §10)
	Priority    *Priority  // optional, display-only (MX3-follow-up; v1.md §7)
	CreatedAt   string     // RFC3339
	UpdatedAt   string     // RFC3339
}

// TaskInput is the editable field set (v1.md §7: title, description, due date;
// category_id added by ADR-0009; goal_id added by MX3; priority added by
// MX3-follow-up).
type TaskInput struct {
	Title       string
	Description string
	DueDate     *timezone.Date
	CategoryID  *uuid.UUID
	GoalID      *uuid.UUID
	Priority    *Priority
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
	// *ValidationError if CategoryID is set but not the caller's or archived, or
	// GoalID is set but not the caller's.
	CreateTask(ctx context.Context, accountID uuid.UUID, in TaskInput) (Task, error)

	// UpdateTask changes a task's title, description, due date, category, and goal
	// link (not its state). ErrTaskNotFound if it is missing or not the caller's;
	// *ValidationError for a bad category or goal.
	UpdateTask(ctx context.Context, accountID, taskID uuid.UUID, in TaskInput) error

	// MoveTask sets a task's state to any of the four, recording a transition
	// unless the state is unchanged. ErrTaskNotFound / *ValidationError.
	MoveTask(ctx context.Context, accountID, taskID uuid.UUID, to State) error

	// DeleteTask removes a task (and its transitions, by cascade).
	DeleteTask(ctx context.Context, accountID, taskID uuid.UUID) error

	// Board returns the caller's four columns, each newest-first.
	Board(ctx context.Context, accountID uuid.UUID) (Board, error)

	// CountByCategory implements categories.Counter for the categories overview
	// (ADR-0009).
	CountByCategory(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]int, error)

	// DoneCountInRange returns how many distinct tasks recorded at least one
	// -> DONE transition in [from, to) — a task that bounces in and out of DONE
	// counts once (M6/M7 foundation, Q10-style resolution for tasks).
	DoneCountInRange(ctx context.Context, accountID uuid.UUID, from, to time.Time) (int, error)

	// ProgressByGoal implements goals.ProgressReader for the goals list endpoint
	// (MX3) — for every goal the caller has linked at least one task to, how many
	// of those tasks are currently DONE (done) and how many total (total), keyed
	// by goal id. A goal with no linked tasks is absent from both maps.
	ProgressByGoal(ctx context.Context, accountID uuid.UUID) (done, total map[uuid.UUID]int, err error)

	// AssignableToAccount reports whether taskID exists and belongs to accountID —
	// used by timeline.TaskChecker to validate a task_id a caller links a time
	// block to (MX-TL, mirrors categories.Service/goals.Service's identical
	// method).
	AssignableToAccount(ctx context.Context, accountID, taskID uuid.UUID) (bool, error)

	// CategoriesForTasks returns every one of the caller's tasks mapped to its own
	// category id (nil if the task is uncategorized) — implements
	// timeline.TaskChecker's bulk lookup so a task-linked time block can resolve
	// its inherited category without N+1 (MX-TL).
	CategoriesForTasks(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]*uuid.UUID, error)
}
